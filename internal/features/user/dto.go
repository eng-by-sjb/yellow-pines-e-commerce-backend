package user

import "time"

// Requests
type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	LastName  string `json:"lastName" validate:"required,min=2,max=15,noAllRepeatingChars"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=5,max=10"`
}

type LoginUserRequest struct {
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
	UserAgent string `json:"userAgent" validate:"required"`
	ClientIP  string `json:"clientIP" validate:"required"`
}

type RenewTokensRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	UserAgent    string `json:"userAgent" validate:"required"`
	ClientIP     string `json:"clientIP" validate:"required"`
}

// Responses
type LoginUserCookiesResponse struct {
	AccessToken  TokenDetails `json:"accessToken"`
	RefreshToken TokenDetails `json:"refreshToken"`
}

type RenewTokensResponse struct {
// TokenDetails represents the data for access and refresh tokens.
type TokenDetails struct {
	Value   string    `json:"value"`
	Expires time.Time `json:"expires"`
}
