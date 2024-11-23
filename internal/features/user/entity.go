package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID                 uuid.UUID `json:"user_id"`
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	Email                  string    `json:"email"`
	HashedPassword         string    `json:"-"`
	EncryptedTOPTSecret    string    `json:"-"`
	IsTwoFactorAuthEnabled bool      `json:"is_two_factor_auth_enabled"`
	CreatedAt              string    `json:"created_at"`
	UpdatedAt              string    `json:"updated_at"`
}

type Session struct {
	SessionID    uuid.UUID `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsRevoked    bool      `json:"is_revoked"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	LastUsedAt   time.Time `json:"last_used_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
