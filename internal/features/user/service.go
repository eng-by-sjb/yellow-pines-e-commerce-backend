package user

import (
	"context"
	"fmt"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/auth"
)

type Servicer interface {
	createUser(ctx context.Context, newUser *RegisterUserRequestDTO) (bool, error)
	getUserByEmail(ctx context.Context, email *string) (*User, error)
	getUserByID(ctx context.Context, id *int) (*User, error)
}

type Service struct {
	store Storer
}

func NewService(store Storer) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) createUser(ctx context.Context, newUser *RegisterUserRequestDTO) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	user, err := s.store.findByEmail(ctx, newUser.Email)

	if err != nil {
		fmt.Print((*newUser).Email)
		return false, err
	}

	if user != nil {
		return true, nil
	}

	hashedPassword, err := auth.HashPassword(
		newUser.Password,
	)
	if err != nil {
		return false, err
	}

	if err := ctx.Err(); err != nil {
		return false, err
	}
	err = s.store.create(ctx, &User{
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Email:     newUser.Email,
		Password:  hashedPassword,
	})

	if err != nil {
		return false, err
	}

	return false, nil
}

func (s *Service) getUserByEmail(ctx context.Context, email *string) (*User, error) {
	panic("unimplemented")
}

func (s *Service) getUserByID(ctx context.Context, id *int) (*User, error) {
	panic("unimplemented")
}
