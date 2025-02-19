package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) create(ctx context.Context, admin *Admin) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO admins(first_name, last_name, email, hashed_password) VALUES($1, $2, $3, $4)",
		admin.FirstName,
		admin.LastName,
		admin.Email,
		admin.HashedPassword,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to insert new admin in admin store: %w",
			err,
		)
	}

	return nil
}

func (s *Store) findByEmail(ctx context.Context, email string) (*Admin, error) {
	admin, err := s.getAdminWithContext(
		ctx,
		"SELECT * FROM admins WHERE email = $1",
		email,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find admin by email in admin store: %w",
			err,
		)
	}

	return admin, nil
}

func (s *Store) findByID(ctx context.Context, adminID uuid.UUID) (*Admin, error) {
	admin, err := s.getAdminWithContext(
		ctx,
		"SELECT * FROM admins WHERE admin_id = $1",
		adminID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find admin by id in admin store: %w",
			err,
		)
	}

	return admin, nil
}

func (s *Store) getAdminWithContext(ctx context.Context, query string, args ...any) (*Admin, error) {
	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf(
			"failed to query db in admin store getAdminWithContext: %w",
			err,
		)
	}
	defer rows.Close()

	admin := new(Admin) // initialize admin to its zero values
	for rows.Next() {
		err = scanRowsIntoAdmin(rows, admin)
		if err != nil {
			return nil, err
		}
	}

	return admin, nil
}

// scanRowsIntoAdmin takes in sql rows and a admin that has been initialized to
// its zero values and returns the admin and nil or nil and an error if any.
func scanRowsIntoAdmin(rows *sql.Rows, admin *Admin) error {
	if admin == nil {
		return errors.New(
			"scanRowsIntoAdmin err in admin store",
		)
	}

	err := rows.Scan(
		&admin.AdminID,
		&admin.FirstName,
		&admin.LastName,
		&admin.Email,
		&admin.HashedPassword,
		&admin.EncryptedTOPTSecret,
		&admin.IsTwoFactorAuthEnabled,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to scan row into admin in admin store: %w",
			err,
		)
	}

	return nil
}
