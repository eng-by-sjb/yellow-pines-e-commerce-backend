package user

import (
	"context"
	"log"
	"strings"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/dto"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/severerrors"
	"github.com/google/uuid"
)

type Servicer interface {
	registerUser(ctx context.Context, newUser *dto.RegisterUserRequest) error
	loginUser(ctx context.Context, payload *dto.LoginUserRequest) error
	getUserByID(ctx context.Context, userID *uuid.UUID) (*User, error)
}

type Service struct {
	store Storer
}

func NewService(store Storer) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) registerUser(ctx context.Context, newUser *dto.RegisterUserRequest) error {
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

func (s *Service) getUserByEmail(ctx context.Context, email *string) (*User, error) {
	panic("unimplemented")
}

func (s *Service) getUserByID(ctx context.Context, id *int) (*User, error) {
	panic("unimplemented")
}
