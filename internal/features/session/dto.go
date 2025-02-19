package session

import "github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/interfaces"

// requests

type RenewTokensRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
	UserAgent    string `json:"userAgent" validate:"required"`
	ClientIP     string `json:"clientIP" validate:"required"`
}

// responses

type RenewTokensCookiesResponse struct {
	AccessToken  interfaces.TokenDetails `json:"accessToken"`
	RefreshToken interfaces.TokenDetails `json:"refreshToken"`
}
