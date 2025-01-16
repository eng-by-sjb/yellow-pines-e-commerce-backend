package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type UserStorer interface {
	create(ctx context.Context, user *User) error
	findByEmail(ctx context.Context, email string) (*User, error)
	findByID(ctx context.Context, userID uuid.UUID) (*User, error)
}

type SessionStorer interface {
	createSession(ctx context.Context, session *Session) error
	findSessionByUserIDAndUserAgent(ctx context.Context, userID uuid.UUID, UserAgent string) (*Session, error)
	deleteSessionByID(ctx context.Context, sessionID uuid.UUID) error
	deleteAllSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	findSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
}
type Storer interface {
	UserStorer
	SessionStorer
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

func (s *Store) createSession(ctx context.Context, session *Session) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO sessions(session_id,user_id, refresh_token, expires_at, user_agent, client_ip) VALUES($1, $2, $3, $4, $5, $6)",
		session.SessionID,
		session.UserID,
		session.RefreshToken,
		session.ExpiresAt,
		session.UserAgent,
		session.ClientIP,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to insert new session in user store: %w",
			err,
		)
	}

	return nil
}

func (s *Store) findSessionByUserIDAndUserAgent(ctx context.Context, userID uuid.UUID, UserAgent string) (*Session, error) {
	// session := new(Session)
	// err := s.db.QueryRowContext(
	// 	ctx,
	// 	"SELECT * FROM sessions WHERE user_id = $1 AND user_agent = $2",
	// 	userID,
	// 	UserAgent,
	// ).Scan(
	// 	&session.SessionID,
	// 	&session.UserID,
	// 	&session.RefreshToken,
	// 	&session.ExpiresAt,
	// 	&session.IsRevoked,
	// 	&session.UserAgent,
	// 	&session.ClientIP,
	// 	&session.LastUsedAt,
	// 	&session.CreatedAt,
	// 	&session.UpdatedAt,
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf(
	// 		"failed to find session by user id and user agent in user store: %w",
	// 		err,
	// 	)
	// }
	// return session, nil

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT * FROM sessions WHERE user_id = $1 AND user_agent = $2",
		userID,
		UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query db in user store findSessionByUserIDAndUserAgent: %w",
			err,
		)
	}
	defer rows.Close()

	session := new(Session)
	for rows.Next() {
		session, err = scanRowsIntoSession(rows, session)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
}

func (s *Store) deleteAllSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return deleteSessionWithContext(
		ctx,
		s,
		"DELETE FROM sessions WHERE user_id = $1",
		userID,
	)
}

func (s *Store) deleteSessionByID(ctx context.Context, sessionID uuid.UUID) error {
	return deleteSessionWithContext(
		ctx,
		s,
		"DELETE FROM sessions WHERE session_id = $1",
		sessionID,
	)
}

func (s *Store) findSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	// session := new(Session)
	rows, err := s.db.QueryContext(
		ctx,
		"SELECT * FROM sessions WHERE session_id = $1",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find session by id in user store: %w",
			err,
		)
	}
	defer rows.Close()

	session := new(Session)
	for rows.Next() {
		session, err = scanRowsIntoSession(rows, session)
		if err != nil {
			return nil, err
		}
	}

	return session, nil
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
		err = scanRowsIntoUser(rows, user)
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
func scanRowsIntoUser(rows *sql.Rows, user *User) error {
	if user == nil {
		return errors.New(
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
		return fmt.Errorf(
			"failed to scan row into user in user store: %w",
			err,
		)
	}

	return nil
}

func scanRowsIntoSession(rows *sql.Rows, session *Session) (*Session, error) {
	if session == nil {
		return nil, errors.New(
			"scanRowsIntoSession err in user store",
		)
	}

	err := rows.Scan(
		&session.SessionID,
		&session.UserID,
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.IsRevoked,
		&session.UserAgent,
		&session.ClientIP,
		&session.LastUsedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to scan row into session in user store: %w",
			err,
		)
	}

	return session, nil
}

func deleteSessionWithContext(ctx context.Context, s *Store, exec string, args ...any) error {
	_, err := s.db.ExecContext(
		ctx,
		exec,
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to delete session in user store: %w",
			err,
		)
	}

	return nil
}
