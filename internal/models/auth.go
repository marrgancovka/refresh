package models

import (
	"github.com/google/uuid"
	"time"
)

type PairToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenPayload struct {
	UserID uuid.UUID `json:"user_id"`
	UserIP string    `json:"user_ip"`
	Exp    time.Time `json:"exp"`
}

type Session struct {
	UserID    uuid.UUID `json:"user_id"`
	HashToken string    `json:"hash_token"`
}
