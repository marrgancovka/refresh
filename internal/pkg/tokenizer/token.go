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
		"exp": payload.Exp.Unix(),
	})

	return token.SignedString(t.cfg.KeyJWT)
}

func (t *Tokenizer) ValidateJWT(tokenString string) (*models.TokenPayload, error) {
	t.log.Debug(tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, myerrors.ErrInvalidToken
		}

		//if exp, ok := token.Claims.(jwt.MapClaims)["exp"]; ok {
		//	t.log.Debug("exp, ok: ", "exp", exp, "ok", ok)
		//	//expTime := time.Unix(exp, 0)
		//	//t.log.Debug("", "exp", expTime)
		//	if exp.(time.Time).Before(time.Now()) {
		//		return nil, myerrors.ErrTokenExpired
		//	}
		//}

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
	pair := &models.PairToken{}
	payload.Exp = time.Now().Add(t.cfg.AccessExpirationTime)
	pair.ExpAccessToken = payload.Exp
	accessToken, err := t.GenerateJWT(payload)
	if err != nil {
		t.log.Error("generating access token", "error", err)
		return nil, err
	}
	pair.AccessToken = accessToken

	payload.Exp = time.Now().Add(t.cfg.RefreshExpirationTime)
	pair.ExpRefreshToken = payload.Exp
	refreshToken, err := t.GenerateJWT(payload)
	if err != nil {
		t.log.Error("generating refresh token", "error", err)
		return nil, err
	}
	pair.RefreshToken = refreshToken

	return pair, nil
}

func parseClaims(token *jwt.Token) (*models.TokenPayload, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, myerrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, errors.New("invalid userID in token claims")
	}

	ip, ok := claims["ip"].(string)
	if !ok {
		return nil, errors.New("invalid IP in token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid exp in token claims")
	}
	expTime := time.Unix(int64(exp), 0)

	return &models.TokenPayload{
		UserID: userID,
		UserIP: ip,
		Exp:    expTime,
	}, nil
}
