package tokenizer

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"log/slog"
	"refresh/internal/models"
	"refresh/pkg/myerrors"
	"time"
)

type Params struct {
	fx.In

	Config Config
	Logger *slog.Logger
}

type Tokenizer struct {
	cfg Config
	log *slog.Logger
}

func New(p Params) *Tokenizer {
	return &Tokenizer{
		cfg: p.Config,
		log: p.Logger,
	}
}

func (t *Tokenizer) GenerateJWT(payload *models.TokenPayload) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": payload.UserID,
		"ip":  payload.UserIP,
		"exp": payload.Exp,
	})

	return token.SignedString(t.cfg.KeyJWT)
}

func (t *Tokenizer) ValidateJWT(tokenString string) (*models.TokenPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, myerrors.ErrInvalidToken
		}
		return t.cfg.KeyJWT, nil
	})
	if err != nil {
		t.log.Error("parsing token", "error", err)
		return nil, myerrors.ErrInvalidToken
	}

	payload, err := parseClaims(token)
	if err != nil {
		t.log.Error("parsing token claims", "error", err)
		return nil, myerrors.ErrInvalidToken
	}

	if payload.Exp.Before(time.Now()) {
		t.log.Error("token expired")
		return nil, myerrors.ErrTokenExpired
	}

	return payload, nil
}

func (t *Tokenizer) GeneratePairToken(payload *models.TokenPayload) (*models.PairToken, error) {
	payload.Exp = time.Now().Add(t.cfg.AccessExpirationTime)
	accessToken, err := t.GenerateJWT(payload)
	if err != nil {
		return nil, err
	}

	payload.Exp = time.Now().Add(t.cfg.RefreshExpirationTime)
	refreshToken, err := t.GenerateJWT(payload)
	if err != nil {
		return nil, err
	}

	return &models.PairToken{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func parseClaims(token *jwt.Token) (*models.TokenPayload, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, myerrors.ErrInvalidToken
	}

	userID, ok := claims["sub"].(uuid.UUID)
	if !ok {
		return nil, errors.New("invalid userID in token claims")
	}

	ip, ok := claims["ip"].(string)
	if !ok {
		return nil, errors.New("invalid IP in token claims")
	}

	exp, ok := claims["exp"].(time.Time)
	if !ok {
		return nil, errors.New("invalid exp in token claims")
	}

	return &models.TokenPayload{
		UserID: userID,
		UserIP: ip,
		Exp:    exp,
	}, nil
}
