package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/dto"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/severerrors"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/validate"
	"github.com/go-chi/chi"
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
	router.Post(
		"/register",
		handlerutils.MakeHandler(h.registerUserHandler),
	)
	router.Post(
		"/login",
		handlerutils.MakeHandler(h.loginUserHandler),
	)
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload *dto.RegisterUserRequest
	var err error
	defer r.Body.Close()

	if err = handlerutils.ParseJSON(r, &payload); err != nil {
		return severerrors.New(
			http.StatusBadRequest,
			severerrors.ErrInvalidRequestPayload.Error(),
			nil,
		)
	}

	if err = validate.StructFields(payload); err != nil {
		return severerrors.New(
			http.StatusUnprocessableEntity,
			severerrors.ErrValidationFailed.Error(),
			err,
			)
	}

	if err = h.service.registerUser(ctx, payload); err != nil {
		switch {
		case errors.Is(err, severerrors.ErrUserAlreadyExists):
			return severerrors.New(
				http.StatusConflict,
				severerrors.ErrUserAlreadyExists.Error(),
				nil,
			)
		default:
			return err
		}
	}

	return handlerutils.WriteSuccessJSON(
			w,
		http.StatusCreated,
		"user created",
		nil,
		)
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
