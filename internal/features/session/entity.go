package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID    uuid.UUID `json:"session_id"`
	EntityID     uuid.UUID `json:"user_id"`
	EntityType   string    `json:"entity_type"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsRevoked    bool      `json:"is_revoked"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	LastUsedAt   time.Time `json:"last_used_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
