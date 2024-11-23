package user

import (
	"context"
	"strings"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/severerrors"
	"github.com/google/uuid"
)

type Servicer interface {
	registerUser(ctx context.Context, newUser *RegisterUserRequest) error
	loginUser(ctx context.Context, payload *LoginUserRequest) (*LoginUserResponse, error)
	logoutUser(ctx context.Context, userID *uuid.UUID) (*User, error)
}

type Service struct {
	store      Storer
	tokenMaker auth.TokenMaker
}

func NewService(store Storer, tokenMaker auth.TokenMaker) *Service {
	return &Service{
		store:      store,
		tokenMaker: tokenMaker,
	}
}

func (s *Service) registerUser(ctx context.Context, newUser *RegisterUserRequest) error {
	newUser.FirstName = strings.TrimSpace(newUser.FirstName)
	newUser.LastName = strings.TrimSpace(newUser.LastName)
	newUser.Email = strings.TrimSpace(newUser.Email)

	user, err := s.store.findByEmail(ctx, newUser.Email)
	if err != nil {
		return err
	}

	//todo: to Decouple service from external depends. put uuid.Nil in a
	// todo: --- variable somewhere and import it here
	if user.UserID != uuid.Nil {
		return severerrors.ErrUserAlreadyExists
	}

	hashedPassword, err := auth.HashPassword(
		newUser.Password,
	)
	if err != nil {
		return err
	}

	err = s.store.create(ctx, &User{
		FirstName:      newUser.FirstName,
		LastName:       newUser.LastName,
		Email:          newUser.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) loginUser(ctx context.Context, payload *LoginUserRequest) (*LoginUserResponse, error) {
	u, err := s.store.findByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	if !auth.ComparePassword(u.HashedPassword, payload.Password) {
		return nil, severerrors.ErrInvalidCredentials
	}

	accessToken, _, err := s.tokenMaker.GenerateToken(
		false,
		u.UserID.String(),
	)
	if err != nil {
		return nil, err
	}

	existingSession, err := s.store.findSessionByUserIDAndUserAgent(ctx, u.UserID, payload.UserAgent)
	if err != nil {
		return nil, err
	}

	if existingSession != nil {

		switch {
		case existingSession.ExpiresAt.After(time.Now()) && !existingSession.IsRevoked:
			return &LoginUserResponse{
				SessionID:    existingSession.SessionID.String(),
				AccessToken:  accessToken,
				RefreshToken: existingSession.RefreshToken,
			}, nil

		default:
			s.store.deleteSessionByID(ctx, existingSession.SessionID)
		}
	}

	refreshToken, refreshClaims, err := s.tokenMaker.GenerateToken(
		true,
		u.UserID.String(),
	)
	if err != nil {
		return nil, err
	}

	sessionID, err := uuid.Parse(refreshClaims.RegisteredClaims.ID)
	if err != nil {
		return nil, err
	}

	err = s.store.createSession(
		ctx,
		&Session{
			SessionID:    sessionID,
			UserID:       u.UserID,
			RefreshToken: refreshToken,
			ExpiresAt:    refreshClaims.RegisteredClaims.ExpiresAt.Time,
			UserAgent:    payload.UserAgent,
			ClientIP:     payload.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}

	return &LoginUserResponse{
		SessionID:    refreshClaims.RegisteredClaims.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) logoutUser(ctx context.Context, userID *uuid.UUID) (*User, error) {
	panic("unimplemented")
}
