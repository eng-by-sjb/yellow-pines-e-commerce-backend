package interfaces

import (
	"time"

	"github.com/google/uuid"
)

// Requests

type LoginEntityRequest struct {
	EntityID   uuid.UUID `json:"entityID" validate:"required"`
	EntityType string    `json:"entityType" validate:"required"`
	UserAgent  string    `json:"userAgent" validate:"required"`
	ClientIP   string    `json:"clientIP" validate:"required"`
}

// Responses

type LoginEntityCookiesResponse struct {
	AccessToken  TokenDetails `json:"accessToken"`
	RefreshToken TokenDetails `json:"refreshToken"`
}

// TokenDetails represents the data for access and refresh tokens.
type TokenDetails struct {
	Value   string    `json:"value"`
	Expires time.Time `json:"expires"`
}
