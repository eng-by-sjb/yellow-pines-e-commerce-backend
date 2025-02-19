package session

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
	renewTokens(ctx context.Context, payload *RenewTokensRequest) (*RenewTokensCookiesResponse, error)
}

type handler struct {
	service servicer
}

func NewHandler(sessionService servicer) *handler {
	return &handler{
		service: sessionService,
	}
}

func (h *handler) RegisterRoutes(router *chi.Mux) {
	router.Post(
		"/tokens/renew",
		handlerutils.MakeHandler(h.renewTokensHandler),
	)
}

func (h *handler) renewTokensHandler(w http.ResponseWriter, r *http.Request) error {
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
		"tokens renewed and attached to cookies",
		nil,
	)

}
