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
	logoutUserHandler(ctx context.Context, refreshToken string) error
	renewTokens(ctx context.Context, refreshToken string) (*RenewedTokensResponse, error)
}

type Service struct {
	store        Storer
	TokenService auth.TokenServicer
}

func NewService(store Storer, tokenMaker auth.TokenServicer) *Service {
	return &Service{
		store:        store,
		TokenService: tokenMaker,
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

	accessToken, _, err := s.TokenService.GenerateToken(
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

	refreshToken, refreshClaims, err := s.TokenService.GenerateToken(
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

func (s *Service) logoutUserHandler(ctx context.Context, refreshToken string) error {
	isValid, claims, err := s.TokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	if !isValid {
		return severerrors.ErrInvalidRefreshToken
	}

	sessionID, err := uuid.Parse(claims.ID)
	if err != nil {
		return err
	}

	if session, err := s.store.findSessionByID(ctx, sessionID); err != nil {
		return err
	} else if session.SessionID == uuid.Nil {
		return severerrors.ErrSessionNotFound
	}

	if err := s.store.deleteSessionByID(ctx, sessionID); err != nil {
		return err
	}

	return nil
}

func (s *Service) renewTokens(ctx context.Context, refreshToken string) (*RenewedTokensResponse, error) {
	//validate refresh token
	isValid, claims, err := s.TokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, severerrors.ErrInvalidRefreshToken
	}

	sessionID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, err
	}

	session, err := s.store.findSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	// if there is no session with that session id, delete all sessions for the
	//user. Detect token reuse
	if session.SessionID == uuid.Nil {
		err := s.store.deleteAllSessionsByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}

		return nil, severerrors.ErrSessionNotFound
	}

	if session.IsRevoked {
		return nil, severerrors.ErrInvalidRefreshToken
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, severerrors.ErrInvalidRefreshToken
	}

	refreshTokens, err := s.TokenService.RefreshTokens(
		session.UserID.String(),
	)
	if err != nil {
		return nil, err
	}

	newSessionID, err := uuid.Parse(refreshTokens.RefreshTokenClaims.ID)
	if err != nil {
		return nil, err
	}

	err = s.store.createSession(
		ctx,
		&Session{
			SessionID:    newSessionID,
			UserID:       session.UserID,
			RefreshToken: refreshTokens.RefreshToken,
			ExpiresAt:    refreshTokens.RefreshTokenClaims.ExpiresAt.Time,
			UserAgent:    session.UserAgent,
			ClientIP:     session.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}

	return &RenewedTokensResponse{
		AccessToken:  refreshTokens.AccessToken,
		RefreshToken: refreshTokens.RefreshToken,
	}, nil
}
