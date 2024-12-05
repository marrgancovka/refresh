package auth

import (
	"github.com/google/uuid"
	"refresh/internal/models"
)

type Usecase interface {
	Authenticate(id uuid.UUID) (*models.TokensResponse, error)
	Refresh(token string) (*models.TokensResponse, error)
}

type Repository interface {
	CreateSession(session *models.Session) error
	UpdateSession(session *models.Session) error
}
