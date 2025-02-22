package admin

import (
	"context"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/interfaces"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type adminStorer interface {
	create(ctx context.Context, admin *Admin) error
	findByEmail(ctx context.Context, email string) (*Admin, error)
	findByID(ctx context.Context, adminID uuid.UUID) (*Admin, error)
}
type sessionServicer interface {
	LoginEntity(ctx context.Context, payload *interfaces.LoginEntityRequest) (*interfaces.LoginEntityCookiesResponse, error)
	LogoutEntity(ctx context.Context, refreshToken string) error
}

type service struct {
	sessionService sessionServicer
	adminStore     adminStorer
}

func NewService(adminStore adminStorer, sessionService sessionServicer) *service {
	return &service{
		adminStore:     adminStore,
		sessionService: sessionService,
	}
}

func (s *service) loginAdmin(ctx context.Context, payload *LoginAdminRequest) (*LoginAdminCookiesResponse, error) {
	// check email to find user
	admin, err := s.adminStore.findByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	// compare payload password Against found user password
	if !admin.comparePassword(payload.Password) {
		return nil, servererrors.ErrInvalidCredentials
	}

	resp, err := s.sessionService.LoginEntity(
		ctx,
		&interfaces.LoginEntityRequest{
			EntityID:   admin.AdminID,
			EntityType: "admin",
			UserAgent:  payload.UserAgent,
			ClientIP:   payload.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}

	return &LoginAdminCookiesResponse{
		AccessToken: interfaces.TokenDetails{
			Value:   resp.AccessToken.Value,
			Expires: resp.AccessToken.Expires,
		},
		RefreshToken: interfaces.TokenDetails{
			Value:   resp.RefreshToken.Value,
			Expires: resp.RefreshToken.Expires,
		},
	}, nil
}

func (s *service) logoutAdmin(ctx context.Context, refreshToken string) error {
	return s.sessionService.LogoutEntity(ctx, refreshToken)
}

