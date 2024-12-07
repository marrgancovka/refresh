package http

import (
	"errors"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
	"refresh/internal/models"
	"refresh/internal/pkg/auth"
	"refresh/pkg/myerrors"
	"refresh/pkg/responser"
)

const RefreshCookieName = "refresh_token"

type Params struct {
	fx.In

	Usecase auth.Usecase
	Logger  *slog.Logger
}

type Handler struct {
	uc  auth.Usecase
	log *slog.Logger
}

func New(p Params) *Handler {
	return &Handler{uc: p.Usecase, log: p.Logger}
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	clientIP := r.RemoteAddr

	if userID == "" {
		h.log.Error("user id in query params not found")
		responser.Send400(w, "id not found")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		h.log.Error("invalid user id", "error", err)
		responser.Send400(w, "uncorrected id")
		return
	}

	tokens, err := h.uc.Authenticate(r.Context(), &models.TokenPayload{UserID: id, UserIP: clientIP})
	if err != nil {
		h.log.Error("authenticate", "error", err)
		responser.Send500(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    tokens.RefreshToken,
		Path:     "/",
		Expires:  tokens.ExpRefreshToken,
		HttpOnly: true,
	})

	responser.Send200(w, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil || cookie == nil {
		h.log.Error("invalid cookie", "error", err)
		responser.Send401(w, "refresh token not found")
	}
	refresh := cookie.Value
	clientIP := r.RemoteAddr

	tokens, err := h.uc.Refresh(r.Context(), refresh, clientIP)
	if err != nil {
		h.log.Error("refresh", "error", err)
		switch {
		case errors.Is(err, myerrors.ErrInvalidToken):
			responser.Send401(w, err.Error())
			return
		case errors.Is(err, myerrors.ErrInappropriateRefreshToken):
			responser.Send401(w, err.Error())
			return
		case errors.Is(err, myerrors.ErrTokenExpired):
			responser.Send401(w, err.Error())
			return
		default:
			responser.Send500(w)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    tokens.RefreshToken,
		Path:     "/",
		Expires:  tokens.ExpRefreshToken,
		HttpOnly: true,
	})

	responser.Send200(w, tokens)
}
