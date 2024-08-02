package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/helpers"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service Servicer
}

func NewHandler(service Servicer) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *chi.Mux) {
	router.Post("/register", h.registerUserHandler)
	router.Post("/login", h.loginUserHandler)
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload RegisterUserRequestDTO
	var err error
	defer r.Body.Close()

	err = helpers.ParseJSON(r, &payload)
	if err != nil {
		helpers.WriteError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if err := helpers.Validator.Struct(payload); err != nil {
		validationErrors := make([]helpers.ValidationError, 0, 10)

		var message string
		var field string
		var code string

		for _, err := range err.(validator.ValidationErrors) {
			field = fmt.Sprint(
				strings.ToLower(err.Field()[:1]),
				err.Field()[1:],
			)
			code = fmt.Sprintf(
				"%s_%s",
				strings.ToUpper(err.Field()),
				strings.ToUpper(err.Tag()),
			)

			switch err.Tag() {
			case "required":
				message = fmt.Sprintf(
					"%s is required",
					field,
				)

			case "email":
				message = fmt.Sprintf(
					"%s is not a valid email",
					field,
				)

			case "min":
				message = fmt.Sprintf(
					"%s must be at least %s characters long",
					field,
					err.Param(),
				)

			case "max":
				message = fmt.Sprintf(
					"%s must be at most %s characters long",
					field,
					err.Param(),
				)

			default:
				message = fmt.Sprintf(
					"%s is not valid",
					field,
				)
			}

			validationErrors = append(validationErrors, helpers.ValidationError{
				Field:   field,
				Message: message,
				Code:    code,
			})
		}

		helpers.WriteError(
			w,
			http.StatusBadRequest,
			"validation failed for one or more fields",
			validationErrors,
		)
		return
	}

	newUser := &RegisterUserRequestDTO{
		FirstName: strings.TrimSpace(payload.FirstName),
		LastName:  strings.TrimSpace(payload.LastName),
		Email:     strings.TrimSpace(payload.Email),
		Password:  payload.Password,
	}

	exists, err := h.service.createUser(ctx, newUser)

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			helpers.WriteError(
				w,
				http.StatusBadGateway,
				"request timed out",
			)
			return
		}

		helpers.WriteError(
			w,
			http.StatusRequestTimeout,
			"request cancelled",
		)

	default:
	}

	if err != nil {
		helpers.WriteError(
			w,
			http.StatusInternalServerError,
			"something went wrong",
		)
		return
	}

	if exists {
		helpers.WriteError(
			w,
			http.StatusConflict,
			fmt.Sprintf(
				"user with email '%s' already exists",
				newUser.Email,
			))
		return
	}

	// todo: sort Response for created user
	helpers.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) loginUserHandler(w http.ResponseWriter, r *http.Request) {

}
