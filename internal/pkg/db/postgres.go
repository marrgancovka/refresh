package db

import (
	"database/sql"
	"fmt"
	"go.uber.org/fx"
	"log/slog"
)

func getConnStr(cfg *Config) string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
	)
}

type PostgresParams struct {
	fx.In

	Cfg    Config
	Logger *slog.Logger
}

func NewPostgresConn(p PostgresParams) (*sql.DB, error) {
	connStr := getConnStr(&p.Cfg)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		p.Logger.Error("open connection: " + err.Error())
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		p.Logger.Error("ping database: " + err.Error())
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, nil
}
