package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
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
		return servererrors.New(
			http.StatusBadRequest,
			servererrors.ErrInvalidRequestPayload.Error(),
			nil,
		)
	}

	if err = validate.StructFields(payload); err != nil {
		return servererrors.New(
			http.StatusUnprocessableEntity,
			servererrors.ErrValidationFailed.Error(),
			err,
		)
	}

	if err = h.service.registerUser(ctx, payload); err != nil {
		switch {
		case errors.Is(err, servererrors.ErrUserAlreadyExists):
			return servererrors.New(
				http.StatusConflict,
				servererrors.ErrUserAlreadyExists.Error(),
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
		case errors.Is(err, servererrors.ErrInvalidCredentials):
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrInvalidCredentials.Error(),
				nil,
			)
		default:
			return err
		}
	}

	cookies := []handlerutils.Cookie{
		{
			Name:    "accessToken",
			Value:   loginUserResponse.AccessToken.Value,
			Expires: loginUserResponse.AccessToken.Expires,
		},
		{
			Name:    "refreshToken",
			Value:   loginUserResponse.RefreshToken.Value,
			Expires: loginUserResponse.RefreshToken.Expires,
		},
	}
	handlerutils.SetCookies(
		w,
		cookies,
	)

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusCreated,
		"access and refresh tokens attached to cookies",
		nil,
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
			return servererrors.New(
				http.StatusForbidden,
				servererrors.ErrNoRefreshTokenCookie.Error(),
				nil,
			)
		default:
			return err
		}
	}

	cookiesNames := []string{
		"accessToken",
		"refreshToken",
	}

	if err = h.service.logoutUser(ctx, refreshToken.Value); err != nil {
		handlerutils.ClearCookie(
			w,
			&cookiesNames,
		)
		// todo: handle err well
		return err
	}

	handlerutils.ClearCookie(
		w,
		&cookiesNames,
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
	payload := new(RenewTokensRequest)

	payload.ClientIP = handlerutils.GetClientIP(r)
	payload.UserAgent = r.UserAgent()

	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			return servererrors.New(
				http.StatusForbidden,
				servererrors.ErrNoRefreshTokenCookie.Error(),
				nil,
			)
		default:
			return err
		}
	}

	payload.RefreshToken = refreshToken.Value

	renewedTokens, err := h.service.renewTokens(
		ctx,
		payload,
	)
	if err != nil {
		switch {
		case errors.Is(err, servererrors.ErrInvalidRefreshToken):
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrInvalidRefreshToken.Error(),
				nil,
			)

		case errors.Is(err, servererrors.ErrSessionNotFound):
			return servererrors.New(
				http.StatusUnauthorized,
				servererrors.ErrSessionNotFound.Error(),
				nil,
			)
		default:
			return err
		}
	}

	cookies := []handlerutils.Cookie{
		{
			Name:    "accessToken",
			Value:   renewedTokens.AccessToken.Value,
			Expires: renewedTokens.AccessToken.Expires,
		},
		{
			Name:    "refreshToken",
			Value:   renewedTokens.RefreshToken.Value,
			Expires: renewedTokens.RefreshToken.Expires,
		},
	}
	handlerutils.SetCookies(
		w,
		cookies,
	)

	return handlerutils.WriteSuccessJSON(
		w,
		http.StatusOK,
		"tokens renewed",
		renewedTokens,
	)

}
