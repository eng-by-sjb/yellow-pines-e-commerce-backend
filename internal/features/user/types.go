package user

import (
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
