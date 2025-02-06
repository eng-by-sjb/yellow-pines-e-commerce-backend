package user

import (
	"context"
	"strings"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type Servicer interface {
	registerUser(ctx context.Context, newUser *RegisterUserRequest) error
	loginUser(ctx context.Context, payload *LoginUserRequest) (*LoginUserCookiesResponse, error)
	logoutUser(ctx context.Context, refreshToken string) error
	renewTokens(ctx context.Context, payload *RenewTokensRequest) (*RenewTokensCookiesResponse, error)
}

type Service struct {
	store        Storer
	TokenService auth.TokenServicer
}

func NewService(store Storer, tokenService auth.TokenServicer) *Service {
	return &Service{
		store:        store,
		TokenService: tokenService,
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
		return servererrors.ErrUserAlreadyExists
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

func (s *Service) loginUser(ctx context.Context, payload *LoginUserRequest) (*LoginUserCookiesResponse, error) {
	u, err := s.store.findByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	if !auth.ComparePassword(u.HashedPassword, payload.Password) {
		return nil, servererrors.ErrInvalidCredentials
	}

	accessToken, accessClaims, err := s.TokenService.GenerateToken(
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
			return &LoginUserCookiesResponse{
				AccessToken: TokenDetails{
					Value:   accessToken,
					Expires: accessClaims.ExpiresAt.Time,
				},
				RefreshToken: TokenDetails{
					Value:   existingSession.RefreshToken,
					Expires: existingSession.ExpiresAt,
				},
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

	return &LoginUserCookiesResponse{
		AccessToken: TokenDetails{
			Value:   accessToken,
			Expires: accessClaims.ExpiresAt.Time,
		},
		RefreshToken: TokenDetails{
			Value:   refreshToken,
			Expires: refreshClaims.ExpiresAt.Time,
		},
	}, nil
}

func (s *Service) logoutUser(ctx context.Context, refreshToken string) error {
	isValid, refreshTokenClaims, err := s.TokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err
	}

	if !isValid {
		return servererrors.ErrInvalidRefreshToken
	}

	sessionID, err := uuid.Parse(refreshTokenClaims.ID)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(refreshTokenClaims.UserID)
	if err != nil {
		return err
	}

	session, err := s.store.findSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.SessionID == uuid.Nil {
		err := s.store.deleteAllSessionsByUserID(ctx, userID)
		if err != nil {
			return err
		}

		return servererrors.ErrSessionNotFound
	}

	if err := s.store.deleteSessionByID(ctx, sessionID); err != nil {
		return err
	}

	return nil
}

func (s *Service) renewTokens(ctx context.Context, payload *RenewTokensRequest) (*RenewTokensCookiesResponse, error) {
	//validate refresh token
	isValid, claims, err := s.TokenService.ValidateRefreshToken(payload.RefreshToken)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, servererrors.ErrInvalidRefreshToken
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

		return nil, servererrors.ErrSessionNotFound
	}

	if session.IsRevoked {
		return nil, servererrors.ErrInvalidRefreshToken
	}

	// if jwt and session is valid but the info in the payload of req and jwt is
	// invalid, then its compromised. Delete that session
	if session.ExpiresAt.Before(time.Now()) ||
		!session.ExpiresAt.Equal(claims.ExpiresAt.Time) ||
		session.UserID != userID ||
		session.UserAgent != payload.UserAgent ||
		session.ClientIP != payload.ClientIP {
		err := s.store.deleteSessionByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}

		return nil, servererrors.ErrInvalidRefreshToken
	}

	// delete and continue token rotation creation
	err = s.store.deleteSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	refreshTokens, err := s.TokenService.RefreshTokens(
		session.UserID.String(),
	)
	if err != nil {
		return nil, err
	}

	newSessionID, err := uuid.Parse(refreshTokens.NewRefreshTokenClaims.ID)
	if err != nil {
		return nil, err
	}

	err = s.store.createSession(
		ctx,
		&Session{
			SessionID:    newSessionID,
			UserID:       session.UserID,
			RefreshToken: refreshTokens.NewRefreshToken,
			ExpiresAt:    refreshTokens.NewRefreshTokenClaims.ExpiresAt.Time,
			UserAgent:    payload.UserAgent,
			ClientIP:     payload.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}

	return &RenewTokensCookiesResponse{
		AccessToken: TokenDetails{
			Value:   refreshTokens.NewAccessToken,
			Expires: refreshTokens.NewAccessTokenClaims.ExpiresAt.Time,
		},
		RefreshToken: TokenDetails{
			Value:   refreshTokens.NewRefreshToken,
			Expires: refreshTokens.NewRefreshTokenClaims.ExpiresAt.Time,
		},
	}, nil
}
