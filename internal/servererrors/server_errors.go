package servererrors

import (
	"errors"
)

var (
	ErrInvalidRequestPayload = errors.New("invalid request payload")
	ErrValidationFailed      = errors.New("validation failed for one or more fields")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidAccessToken    = errors.New("invalid access token")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrExpiredAccessToken    = errors.New("access token expired")
	ErrExpiredRefreshToken   = errors.New("refresh token expired")
	ErrInternalServerError   = errors.New("internal server error")
	ErrSessionNotFound       = errors.New("session not found")
	ErrUnauthorizedAccess    = errors.New("unauthorized access")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbiddenAccess       = errors.New("forbidden access")
	ErrRequestTimeout        = errors.New("request timeout")
	ErrNoRefreshTokenCookie  = errors.New("missing refresh token cookie")
	ErrProductAlreadyExists  = errors.New("product already exists")
)

type ServerError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Errors     any    `json:"errors,omitempty"`
}

func New(statusCode int, message string, errs any) *ServerError {
	// note to caller: this func can accept "nil" for the "errs" parameter
	// as it is optional
	if errs != nil {
		return &ServerError{
			StatusCode: statusCode,
			Message:    message,
			Errors:     errs,
		}
	}

	return &ServerError{
		StatusCode: statusCode,
		Message:    message,
		Errors:     nil,
	}
}

func (er *ServerError) Error() string {
	return er.Message
}
