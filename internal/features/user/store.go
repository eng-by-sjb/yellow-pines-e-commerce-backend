package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query db in user store getUserWithContext: %w",
			err,
		)
	}
	defer rows.Close()

	user := new(User) // initial user to its zero values
	for rows.Next() {
		user, err = scanRowsIntoUser(rows, user)
		if err != nil {
			return nil, err

			// switch {
			// case errors.Is(err, sql.ErrNoRows):
			// 	return nil, fmt.Errorf(
			// 		"no user found in user store: %w",
			// 		severerrors.ErrUserNotFound,
			// 	)

			// default:
			// 	return nil, err
			// }
			// // if errors.Is(err, sql.ErrNoRows) {
			// // }
		}
	}

	return user, nil
}

// scanRowsIntoUser takes in sql rows and a user that has been initialized to
// its zero values and returns the user and nil or nil and an error if any.
func scanRowsIntoUser(rows *sql.Rows, user *User) (*User, error) {
	if user == nil {
		return nil, errors.New(
			"scanRowsIntoUser err in user store",
		)
	}

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
