package http

import (
	"go.uber.org/fx"
	"net/http"
)

type Params struct {
	fx.In
}

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {

}
