package admin

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/handlerutils"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/servererrors"
	"github.com/go-chi/chi"
)

type servicer interface {
	// registerAdmin(ctx context.Context, payload *RegisterAdminRequest) error
	loginAdmin(ctx context.Context, payload *LoginAdminRequest) (*LoginAdminCookiesResponse, error)
	logoutAdmin(ctx context.Context, refreshToken string) error
}

type handler struct {
	service servicer
}

func NewHandler(service servicer) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) RegisterRoutes(router *chi.Mux) {
	router.Post(
		"/admin/login",
		handlerutils.MakeHandler(h.loginUserHandler),
	)
	router.Post(
		"/admin/logout",
		handlerutils.MakeHandler(h.logoutUserHandler),
	)
}

func (h *handler) loginUserHandler(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		(30 * time.Second),
	)
	defer cancel()

	var payload *LoginAdminRequest
	var err error
	defer r.Body.Close()

	if err = handlerutils.ParseJSON(r, &payload); err != nil {
		return err
	}

	payload.ClientIP = handlerutils.GetClientIP(r)
	payload.UserAgent = r.UserAgent()

	loginUserResponse, err := h.service.loginAdmin(
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

func (h *handler) logoutUserHandler(w http.ResponseWriter, r *http.Request) error {
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

	if err = h.service.logoutAdmin(ctx, refreshToken.Value); err != nil {
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
