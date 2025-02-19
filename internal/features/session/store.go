package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	sessionFields = "session_id, entity_id, entity_type, refresh_token, expires_at, is_revoked, user_agent, client_ip, last_used_at, created_at, updated_at"
)

type store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{
		db: db,
	}
}

func (s *store) create(ctx context.Context, session *Session) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO sessions(session_id, entity_id, entity_type, refresh_token, expires_at, user_agent, client_ip) VALUES($1, $2, $3, $4, $5, $6, $7)",
		session.SessionID,
		session.EntityID,
		session.EntityType,
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

func (s *store) findByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	query := fmt.Sprintf("SELECT %s FROM sessions WHERE session_id = $1", sessionFields)
	rows, err := s.db.QueryContext(
		ctx,
		query,
		sessionID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf(
				"failed to find session by id in session store: %w",
				err,
			)
		default:
			return nil, err
		}
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

func (s *store) findByEntityIDAndUserAgent(ctx context.Context, entityID uuid.UUID, UserAgent string) (*Session, error) {
	query := fmt.Sprintf("SELECT %s FROM sessions WHERE entity_id = $1 AND user_agent = $2", sessionFields)
	rows, err := s.db.QueryContext(
		ctx,
		query,
		entityID,
		UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query db in session store findSessionByUserIDAndUserAgent: %w",
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

func (s *store) deleteAllByEntityID(ctx context.Context, entityID uuid.UUID) error {
	return deleteSessionWithContext(
		ctx,
		s,
		"DELETE FROM sessions WHERE entity_id = $1",
		entityID,
	)
}

func (s *store) deleteByID(ctx context.Context, sessionID uuid.UUID) error {
	return deleteSessionWithContext(
		ctx,
		s,
		"DELETE FROM sessions WHERE session_id = $1",
		sessionID,
	)
}

func scanRowsIntoSession(rows *sql.Rows, session *Session) (*Session, error) {
	if session == nil {
		return nil, errors.New(
			"scanRowsIntoSession err in session store",
		)
	}

	err := rows.Scan(
		&session.SessionID,
		&session.EntityID,
		&session.EntityType,
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
			"failed to scan row into session in session store: %w",
			err,
		)
	}

	return session, nil
}

func deleteSessionWithContext(ctx context.Context, s *store, exec string, args ...any) error {
	_, err := s.db.ExecContext(
		ctx,
		exec,
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to delete session in session store: %w",
			err,
		)
	}

	return nil
}
