package user

import "github.com/google/uuid"

type RegisterUserRequestDTO struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=5,max=10"`
}

type LoginUserRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponseDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

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
