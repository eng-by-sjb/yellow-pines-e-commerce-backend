package user

import (
	"context"
	"strings"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/interfaces"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/google/uuid"
)

type userStorer interface {
	create(ctx context.Context, user *User) error
	findByEmail(ctx context.Context, email string) (*User, error)
	findByID(ctx context.Context, userID uuid.UUID) (*User, error)
}

type sessionServicer interface {
	LoginEntity(ctx context.Context, payload *interfaces.LoginEntityRequest) (*interfaces.LoginEntityCookiesResponse, error)
	LogoutEntity(ctx context.Context, refreshToken string) error
}

type service struct {
	userStore      userStorer
	sessionService sessionServicer
}

func NewService(userStore userStorer, sessionService sessionServicer) *service {
	return &service{
		userStore:      userStore,
		sessionService: sessionService,
	}
}

func (s *service) registerUser(ctx context.Context, newUser *RegisterUserRequest) error {
	newUser.FirstName = strings.TrimSpace(newUser.FirstName)
	newUser.LastName = strings.TrimSpace(newUser.LastName)
	newUser.Email = strings.TrimSpace(newUser.Email)

	user, err := s.userStore.findByEmail(ctx, newUser.Email)
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

	err = s.userStore.create(ctx, &User{
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

func (s *service) loginUser(ctx context.Context, payload *LoginUserRequest) (*LoginUserCookiesResponse, error) {
	// check email to find user
	u, err := s.userStore.findByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	// compare payload password Against found user password
	if !u.comparePassword(payload.Password) {
		return nil, servererrors.ErrInvalidCredentials
	}

	resp, err := s.sessionService.LoginEntity(
		ctx,
		&interfaces.LoginEntityRequest{
			EntityID:   u.UserID,
			EntityType: "user",
			UserAgent:  payload.UserAgent,
			ClientIP:   payload.ClientIP,
		},
	)
	if err != nil {
		return nil, err
	}

	return &LoginUserCookiesResponse{
		AccessToken: TokenDetails{
			Value:   resp.AccessToken.Value,
			Expires: resp.AccessToken.Expires,
		},
		RefreshToken: TokenDetails{
			Value:   resp.RefreshToken.Value,
			Expires: resp.RefreshToken.Expires,
		},
	}, nil
}

func (s *service) logoutUser(ctx context.Context, refreshToken string) error {
	return s.sessionService.LogoutEntity(ctx, refreshToken)
}
