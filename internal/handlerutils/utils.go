package handlerutils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

// MakeHandler is a middleware that takes handler that returns an error and
// return a HandlerFunc to create a centralized error handling, logging and etc.
func MakeHandler(handler APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			log.Println(err)
			var serverError *servererrors.ServerError

			if errors.As(err, &serverError) {
				switch serverError.StatusCode {
				case http.StatusBadRequest:
					WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusConflict:
					WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusUnprocessableEntity:
					WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)

				case http.StatusUnauthorized:
					WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				case http.StatusForbidden:
					WriteErrorJSON(
						w,
						serverError.StatusCode,
						serverError.Error(),
						serverError.Errors,
					)
				}
			} else {
				WriteErrorJSON(
					w,
					http.StatusInternalServerError,
					"something went wrong",
					nil,
				)
			}
		}
	}
}

type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func WriteErrorJSON(w http.ResponseWriter, statusCode int, message string, errs any) error {
	var errResponse *ServerResponse

	if errs != nil {
		errResponse = &ServerResponse{
			Status:  "error",
			Message: message,
			Errors:  errs,
		}
	} else {
		errResponse = &ServerResponse{
			Status:  "error",
			Message: message,
		}
	}

	return writeJSON(w, statusCode, errResponse)
}

func WriteSuccessJSON(w http.ResponseWriter, statusCode int, message string, data any) error {
	var successResponse *ServerResponse

	if data != nil {
		successResponse = &ServerResponse{
			Status:  "success",
			Message: message,
			Data:    data,
		}
	} else {
		successResponse = &ServerResponse{
			Status:  "success",
			Message: message,
		}
	}

	return writeJSON(w, statusCode, successResponse)
}

func ParseJSON(r *http.Request, payload any) error {
	return json.NewDecoder(r.Body).Decode(payload)
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(v)
}

// GetClientIP returns the IP of the request.
func GetClientIP(r *http.Request) string {
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = r.Header.Get("X-Forwarded-For")
	} else if real := r.Header.Get("X-Real-IP"); real != "" {
		clientIP = real
	}
	return clientIP
}

func ClearCookie(w http.ResponseWriter, name string) {
	cookie := http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}
