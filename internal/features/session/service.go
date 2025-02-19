package session

import (
	"context"
	"log"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/interfaces"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type sessionStorer interface {
	create(ctx context.Context, session *Session) error
	findByEntityIDAndUserAgent(ctx context.Context, entityID uuid.UUID, UserAgent string) (*Session, error) // todo: change this to client agent and the postgres Representation
	deleteByID(ctx context.Context, sessionID uuid.UUID) error
	deleteAllByEntityID(ctx context.Context, entityID uuid.UUID) error
	findByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
}

type tokenServicer interface {
	GenerateToken(isRefreshToken bool, entityID string, entityType string) (string, *auth.TokenClaims, error)
	ValidateAccessToken(tokenStr string) (isValid bool, claims *auth.TokenClaims, err error)
	ValidateRefreshToken(tokenStr string) (isValid bool, claims *auth.TokenClaims, err error)
	RefreshTokens(entityID string, entityType string) (*auth.RefreshTokens, error)
}

type service struct {
	sessionStore sessionStorer
	tokenService tokenServicer
}

func NewService(sessionStore sessionStorer, tokenService tokenServicer) *service {
	return &service{
		sessionStore: sessionStore,
		tokenService: tokenService,
	}
}

// todo: mail methods called in for actor model

func (s *service) renewTokens(ctx context.Context, payload *RenewTokensRequest) (*RenewTokensCookiesResponse, error) {
	//validate refresh token
	isValid, claims, err := s.tokenService.ValidateRefreshToken(
		payload.RefreshToken,
	)
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

	session, err := s.sessionStore.findByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(claims.EntityID)
	if err != nil {
		return nil, err
	}

	// if there is no session with that session id, delete all sessions for the
	//user. Detect token reuse
	if session.SessionID == uuid.Nil {
		err := s.sessionStore.deleteAllByEntityID(ctx, entityID)
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
		session.EntityID != entityID ||
		session.UserAgent != payload.UserAgent ||
		session.ClientIP != payload.ClientIP {
		log.Println("100")
		err := s.sessionStore.deleteByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}

		return nil, servererrors.ErrInvalidRefreshToken
	}

	// delete and continue token rotation creation
	err = s.sessionStore.deleteByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	refreshTokens, err := s.tokenService.RefreshTokens(
		session.EntityID.String(),
		session.EntityType,
	)
	if err != nil {
		return nil, err
	}

	newSessionID, err := uuid.Parse(refreshTokens.NewRefreshTokenClaims.ID)
	if err != nil {
		return nil, err
	}

	err = s.sessionStore.create(
		ctx,
		&Session{
			SessionID:    newSessionID,
			EntityID:     session.EntityID,
			EntityType:   session.EntityType,
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
		AccessToken: interfaces.TokenDetails{
			Value:   refreshTokens.NewAccessToken,
			Expires: refreshTokens.NewAccessTokenClaims.ExpiresAt.Time,
		},
		RefreshToken: interfaces.TokenDetails{
			Value:   refreshTokens.NewRefreshToken,
			Expires: refreshTokens.NewRefreshTokenClaims.ExpiresAt.Time,
		},
	}, nil
}

func (s *service) LoginEntity(ctx context.Context, payload *interfaces.LoginEntityRequest) (*interfaces.LoginEntityCookiesResponse, error) {

	accessToken, accessClaims, err := s.tokenService.GenerateToken(
		false,
		payload.EntityID.String(),
		payload.EntityType,
	)
	if err != nil {
		return nil, err
	}

	existingSession, err := s.sessionStore.findByEntityIDAndUserAgent(ctx,
		payload.EntityID,
		payload.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	if existingSession != nil {
		switch {
		case existingSession.ExpiresAt.After(time.Now()) && !existingSession.IsRevoked:
			return &interfaces.LoginEntityCookiesResponse{
				AccessToken: interfaces.TokenDetails{
					Value:   accessToken,
					Expires: accessClaims.ExpiresAt.Time,
				},
				RefreshToken: interfaces.TokenDetails{
					Value:   existingSession.RefreshToken,
					Expires: existingSession.ExpiresAt,
				},
			}, nil

		default:
			s.sessionStore.deleteByID(ctx, existingSession.SessionID)
		}
	}

	refreshToken, refreshClaims, err := s.tokenService.GenerateToken(
		true,
		payload.EntityID.String(),
		payload.EntityType,
	)
	if err != nil {
		return nil, err
	}

	sessionID, err := uuid.Parse(refreshClaims.RegisteredClaims.ID)
	if err != nil {
		return nil, err
	}

	err = s.sessionStore.create(
		ctx,
		&Session{
			SessionID:    sessionID,
			EntityID:     payload.EntityID,
			EntityType:   payload.EntityType,
			RefreshToken: refreshToken,
			ExpiresAt:    refreshClaims.RegisteredClaims.ExpiresAt.Time,
			UserAgent:    payload.UserAgent,
			ClientIP:     payload.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}
	return &interfaces.LoginEntityCookiesResponse{
		AccessToken: interfaces.TokenDetails{
			Value:   accessToken,
			Expires: accessClaims.ExpiresAt.Time,
		},
		RefreshToken: interfaces.TokenDetails{
			Value:   refreshToken,
			Expires: refreshClaims.ExpiresAt.Time,
		},
	}, nil
}

func (s *service) LogoutEntity(ctx context.Context, refreshToken string) error {

	isValid, refreshTokenClaims, err := s.tokenService.ValidateRefreshToken(refreshToken)
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

	entityID, err := uuid.Parse(refreshTokenClaims.EntityID)
	if err != nil {
		return err
	}

	session, err := s.sessionStore.findByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.SessionID == uuid.Nil {
		err := s.sessionStore.deleteAllByEntityID(ctx, entityID)
		if err != nil {
			return err
		}

		return servererrors.ErrSessionNotFound
	}

	if err := s.sessionStore.deleteByID(ctx, sessionID); err != nil {
		return err
	}

	return nil
}
