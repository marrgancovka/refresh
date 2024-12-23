package repo

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"refresh/internal/models"
	"refresh/pkg/myerrors"
)

const (
	checkToken    = `SELECT hash_token FROM sessions WHERE user_id = $1`
	insertSession = `INSERT INTO sessions (hash_token, user_id) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET hash_token = EXCLUDED.hash_token`
)

type Params struct {
	fx.In

	DB     *pgxpool.Pool
	Logger *slog.Logger
}

type Repo struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func New(p Params) *Repo {
	return &Repo{
		db:  p.DB,
		log: p.Logger,
	}
}

func (r *Repo) CheckToken(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	var hashToken string

	row := r.db.QueryRow(ctx, checkToken, userID)
	if err := row.Scan(&hashToken); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return myerrors.ErrInappropriateRefreshToken
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashToken), []byte(refreshToken)); err != nil {
		return myerrors.ErrInappropriateRefreshToken
	}

	return nil
}

func (r *Repo) CreateSession(ctx context.Context, session *models.Session) error {
	if _, err := r.db.Exec(ctx, insertSession, session.HashToken, session.UserID); err != nil {
		return err
	}
	return nil
}
