package user

import (
	"context"
	"errors"
	"net/http"
	"time"

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
	router.Post(
		"/logout",
		handlerutils.MakeHandler(h.logoutUserHandler),
	)
	router.Post(
		"/tokens/renew",
		handlerutils.MakeHandler(h.renewTokensHandler),
	)
}

func (h *Handler) registerUserHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload *RegisterUserRequest
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

func (h *Handler) loginUserHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload *LoginUserRequest
	var err error
	defer r.Body.Close()

	if err = handlerutils.ParseJSON(r, &payload); err != nil {
		return err
	}

	payload.ClientIP = handlerutils.GetClientIP(r)
	payload.UserAgent = r.UserAgent()

	loginUserResponse, err := h.service.loginUser(
		ctx,
		payload,
	)
	if err != nil {

		switch {
		case errors.Is(err, severerrors.ErrInvalidCredentials):
			return severerrors.New(
				http.StatusUnauthorized,
				severerrors.ErrInvalidCredentials.Error(),
				nil,
			)
		default:
			return err
		}
	}

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusCreated,
		"access and refresh tokens created",
		loginUserResponse,
	)
}

func (h *Handler) logoutUserHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var err error
	defer r.Body.Close()

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return severerrors.New(
				http.StatusForbidden,
				severerrors.ErrNoRefreshTokenCookie.Error(),
				nil,
			)
		default:
			return err
		}
	}

	if err = h.service.logoutUserHandler(ctx, refreshToken.Value); err != nil {
		handlerutils.ClearCookie(
			w,
			"refreshToken",
		)
		// todo: handle err well
		return err
	}

	handlerutils.ClearCookie(
		w,
		"refreshToken",
	)

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"user logged out",
		nil,
	)
}

func (h *Handler) renewTokensHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var err error

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return severerrors.New(
				http.StatusForbidden,
				severerrors.ErrNoRefreshTokenCookie.Error(),
				nil,
			)
		default:
			return err
		}
	}

	renewedTokens, err := h.service.renewTokens(
		ctx,
		refreshToken.Value,
	)
	if err != nil {
		switch {
		case errors.Is(err, severerrors.ErrInvalidRefreshToken):
			return severerrors.New(
				http.StatusUnauthorized,
				severerrors.ErrInvalidRefreshToken.Error(),
				nil,
			)
		default:
			return err
		}
	}

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"tokens renewed",
		renewedTokens,
	)

}
