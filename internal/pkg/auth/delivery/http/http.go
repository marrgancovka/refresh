package http

import (
	"errors"
	"github.com/google/uuid"
	"go.uber.org/fx"
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
}

type Handler struct {
	uc auth.Usecase
}

func New(p Params) *Handler {
	return &Handler{uc: p.Usecase}
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	clientIP := r.RemoteAddr

	if userID == "" {
		responser.Send400(w, "id not found")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		responser.Send400(w, "uncorrected id")
		return
	}

	tokens, err := h.uc.Authenticate(r.Context(), &models.TokenPayload{UserID: id, UserIP: clientIP})
	if err != nil {
		responser.Send500(w)
		return
	}

	responser.Send200(w, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refresh := r.Context().Value(RefreshCookieName).(string)
	clientIP := r.RemoteAddr

	tokens, err := h.uc.Refresh(r.Context(), refresh, clientIP)
	if err != nil {
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
	responser.Send200(w, tokens)
}
