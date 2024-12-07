package main

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"log/slog"
	"os"
	"os/signal"
	"refresh/internal/pkg/auth"
	handlerAuth "refresh/internal/pkg/auth/delivery/http"
	"refresh/internal/pkg/auth/repo"
	"refresh/internal/pkg/auth/usecase"
	"refresh/internal/pkg/config"
	"refresh/internal/pkg/db"
	"refresh/internal/pkg/server"
	"refresh/internal/pkg/tokenizer"
	"refresh/migrations"
	"refresh/pkg/logger"
	"syscall"
)

func main() {
	app := fx.New(
		fx.Provide(
			logger.SetupLogger,
			server.NewRouter,

			config.MustLoad,

			db.NewPostgresConn,
			db.NewPostgresPool,

			tokenizer.New,

			fx.Annotate(repo.New, fx.As(new(auth.Repository))),
			fx.Annotate(usecase.New, fx.As(new(auth.Usecase))),
			handlerAuth.New,
		),

		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),

		fx.Invoke(
			server.RunServer,
			migrations.RunMigrations,
		),
	)

	ctx := context.Background()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	if err := app.Start(ctx); err != nil {
		panic(err)
	}

	<-stop
	app.Stop(ctx)
}
