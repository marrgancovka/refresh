package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/fx"
	"log/slog"
)

type Params struct {
	fx.In

	DB     *sql.DB
	Logger *slog.Logger
}

//go:embed postgres/*.sql
var migrationFiles embed.FS

func RunMigrations(p Params) error {
	sourceDriver, err := iofs.New(migrationFiles, "postgres")
	if err != nil {
		return fmt.Errorf("failed to initialize migrations source driver: %w", err)
	}

	dbDriver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		p.Logger.Error("failed to run migrations: ", "myerrors", err)
		return fmt.Errorf("migration up failed: %w", err)
	}

	err = sourceDriver.Close()
	if err != nil {
		p.Logger.Error("failed to close migrations sourceDriver", "myerrors", err)
		return err
	}

	err = dbDriver.Close()
	if err != nil {
		p.Logger.Error("failed to close migrations dbDriver", "myerrors", err)
		return err
	}

	return nil
}
