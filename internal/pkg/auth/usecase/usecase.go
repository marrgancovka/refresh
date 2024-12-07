package usecase

import (
	"context"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	"refresh/internal/models"
	"refresh/internal/pkg/auth"
	"refresh/internal/pkg/tokenizer"
)

type Params struct {
	fx.In

	Repo        auth.Repository
	Tokenizer   tokenizer.Tokenizer
	TokenConfig tokenizer.Config
}

type Usecase struct {
	r        auth.Repository
	t        tokenizer.Tokenizer
	cfgToken tokenizer.Config
}

func New(p Params) *Usecase {
	return &Usecase{r: p.Repo, t: p.Tokenizer, cfgToken: p.TokenConfig}
}

func (uc *Usecase) Authenticate(ctx context.Context, payload *models.TokenPayload) (*models.PairToken, error) {
	pair, err := uc.t.GeneratePairToken(payload)
	if err != nil {
		return nil, err
	}

	hashRefreshToken, err := hashToken(pair.RefreshToken)
	if err != nil {
		return nil, err
	}
	session := &models.Session{
		UserID:    payload.UserID,
		HashToken: hashRefreshToken,
	}

	err = uc.r.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (uc *Usecase) Refresh(ctx context.Context, refreshToken string, ip string) (*models.PairToken, error) {
	payload, err := uc.t.ValidateJWT(refreshToken)
	if err != nil {
		return nil, err
	}

	err = uc.r.CheckToken(ctx, payload.UserID, refreshToken)
	if err != nil {
		return nil, err
	}

	if payload.UserIP != ip {
		payload.UserIP = ip
		// warning to email
	}

	pair, err := uc.t.GeneratePairToken(payload)
	if err != nil {
		return nil, err
	}

	hashRefreshToken, err := hashToken(pair.RefreshToken)
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		UserID:    payload.UserID,
		HashToken: hashRefreshToken,
	}
	err = uc.r.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func hashToken(token string) (string, error) {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedToken), nil
}
