package models

import (
	"github.com/google/uuid"
)

type TokensResponse struct {
	AccessToken string `json:"access_token"`
	RefreshType string `json:"refresh_token"`
}

type Session struct {
	UserID       uuid.UUID `json:"-"`
	RefreshToken string    `json:"-"`
}
