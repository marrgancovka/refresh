package auth

import (
	"context"
	"github.com/google/uuid"
	"refresh/internal/models"
)

type Usecase interface {
	Authenticate(ctx context.Context, payload *models.TokenPayload) (*models.PairToken, error)
	Refresh(ctx context.Context, refreshToken string, ip string) (*models.PairToken, error)
}

type Repository interface {
	CheckToken(ctx context.Context, userID uuid.UUID, refreshToken string) error
	CreateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, session *models.Session) error
}
