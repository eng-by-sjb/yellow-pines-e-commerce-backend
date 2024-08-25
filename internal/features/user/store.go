package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/severerrors"
	"github.com/google/uuid"
)

type Storer interface {
	create(ctx context.Context, user *User) error
	findByEmail(ctx context.Context, email string) (*User, error)
	findByID(ctx context.Context, userID uuid.UUID) (*User, error)
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) create(ctx context.Context, user *User) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf(
			"context error in user store: %w",
			err,
		)
	}

	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO users(first_name, last_name, email, hashed_password) VALUES($1, $2, $3, $4)",
		user.FirstName,
		user.LastName,
		user.Email,
		user.HashedPassword,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to insert new user in user store: %w",
			err,
		)
	}

	return nil
}

func (s *Store) findByEmail(ctx context.Context, email string) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"context error \"findByEmail\" in user store: %w",
			err,
		)
	}

	user, err := s.getUserWithContext(
		ctx,
		"SELECT * FROM users WHERE email = $1",
		email,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find user by email in user store: %w",
			err,
		)
	}

	return user, nil
}

func (s *Store) findByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := s.getUserWithContext(
		ctx,
		"SELECT * FROM users WHERE id = $1",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find user by id in user store: %w",
			err,
		)
	}

	return user, nil
}

func (s *Store) getUserWithContext(ctx context.Context, query string, args ...any) (*User, error) {

	row, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query db in user store: %w",
			err,
		)
	}
	defer row.Close()

	user := new(User)
	for row.Next() {
		user, err = scanRowsIntoUser(row, user)
		if err != nil {
			return nil, err
		}
	}

	if user.UserID == uuid.Nil { // check for zero value if no user is found
		return nil, fmt.Errorf(
			"no user found in user store: %w",
			severerrors.ErrUserNotFound,
		)
	}

	return user, nil
}

func scanRowsIntoUser(rows *sql.Rows, user *User) (*User, error) {
	err := rows.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.HashedPassword,
		&user.EncryptedTOPTSecret,
		&user.IsTwoFactorAuthEnabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to scan row into user in user store: %w",
			err,
		)
	}

	return user, nil
}
