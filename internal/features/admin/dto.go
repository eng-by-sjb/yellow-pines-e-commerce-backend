package admin

import "github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/interfaces"

// Requests

type LoginAdminRequest struct {
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
	UserAgent string `json:"userAgent" validate:"required"`
	ClientIP  string `json:"clientIP" validate:"required"`
}

// Responses
type LoginAdminCookiesResponse struct {
	AccessToken  interfaces.TokenDetails `json:"accessToken"`
	RefreshToken interfaces.TokenDetails `json:"refreshToken"`
}
