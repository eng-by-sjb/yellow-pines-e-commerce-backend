package helpers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type GeneralResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
}

var Validator = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	err := json.NewDecoder(r.Body).Decode(payload)

	if err != nil || r.Body == http.NoBody {
		return errors.New("invalid request payload")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, statusCode int, v any) {

	// todo success format
	// if v.(ErrorResponse).Status != "error" {
	// }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// todo work on internal server error
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "Something went wrong, try again later.", http.StatusInternalServerError)
		return
	}

}

func WriteError(w http.ResponseWriter, statusCode int, message string, errs ...any) {
	var errorsResponse *ErrorResponse

	if len(errs) == 0 {
		errorsResponse = &ErrorResponse{
			Status:  "error",
			Message: message,
		}
	} else {
		errorsResponse = &ErrorResponse{
			Status:  "error",
			Message: message,
			Errors:  errs,
		}
	}

	WriteJSON(w, statusCode, errorsResponse)
}
