package server

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
	handlerEmployee "refresh/internal/pkg/auth/delivery/http"
)

type RouterParams struct {
	fx.In

	Handler *handlerEmployee.Handler
	Logger  *slog.Logger
}

type Router struct {
	handler *mux.Router
}

func NewRouter(p RouterParams) *Router {
	api := mux.NewRouter().PathPrefix("/api").Subrouter()

	v1 := api.PathPrefix("/v1").Subrouter()
	v1.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	auth := v1.PathPrefix("/auth").Subrouter()

	auth.HandleFunc("/login", p.Handler.Authenticate).Methods(http.MethodGet)
	auth.HandleFunc("/refresh", p.Handler.Refresh).Methods(http.MethodGet)

	router := &Router{
		handler: api,
	}

	p.Logger.Info("registered router")

	return router
}
