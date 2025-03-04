package admin

import (
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/google/uuid"
)

type Admin struct {
	AdminID                uuid.UUID `json:"admin_id"`
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	Email                  string    `json:"email"`
	HashedPassword         string    `json:"-"`
	EncryptedTOPTSecret    string    `json:"-"`
	IsTwoFactorAuthEnabled bool      `json:"is_two_factor_auth_enabled"`
	CreatedAt              string    `json:"created_at"`
	UpdatedAt              string    `json:"updated_at"`
}

func (u *Admin) comparePassword(passwordStr string) bool {
	return auth.ComparePassword(u.HashedPassword, passwordStr)
}
