package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"log/slog"
	"refresh/internal/models"
	"refresh/internal/pkg/auth"
	"refresh/internal/pkg/tokenizer"
)

type Params struct {
	fx.In

	Repo      auth.Repository
	Tokenizer *tokenizer.Tokenizer
	Logger    *slog.Logger
}

type Usecase struct {
	r   auth.Repository
	t   *tokenizer.Tokenizer
	log *slog.Logger
}

func New(p Params) *Usecase {
	return &Usecase{r: p.Repo, t: p.Tokenizer, log: p.Logger}
}

func (uc *Usecase) Authenticate(ctx context.Context, payload *models.TokenPayload) (*models.PairToken, error) {
	pair, err := uc.t.GeneratePairToken(payload)
	if err != nil {
		uc.log.Error("failed to generate pair token", "error", err)
		return nil, err
	}

	hashRefreshToken := hashToken(pair.RefreshToken)

	session := &models.Session{
		UserID:    payload.UserID,
		HashToken: hashRefreshToken,
	}

	err = uc.r.CreateSession(ctx, session)
	if err != nil {
		uc.log.Error("failed to create session", "error", err)
		return nil, err
	}

	return pair, nil
}

func (uc *Usecase) Refresh(ctx context.Context, refreshToken string, ip string) (*models.PairToken, error) {
	payload, err := uc.t.ValidateJWT(refreshToken)
	if err != nil {
		uc.log.Error("failed to validate refresh token", "error", err)
		return nil, err
	}

	hashedToken := sha256.Sum256([]byte(refreshToken))
	err = uc.r.CheckToken(ctx, payload.UserID, string(hashedToken[:]))
	if err != nil {
		uc.log.Error("token inappropriate", "error", err)
		return nil, err
	}

	if payload.UserIP != ip {
		uc.log.Info("IP address did not match")
		payload.UserIP = ip
		uc.sendEmail()
	}

	pair, err := uc.t.GeneratePairToken(payload)
	if err != nil {
		uc.log.Error("failed to generate pair token", "error", err)
		return nil, err
	}

	hashRefreshToken := hashToken(pair.RefreshToken)
	uc.log.Debug(hashRefreshToken)
	session := &models.Session{
		UserID:    payload.UserID,
		HashToken: hashRefreshToken,
	}
	err = uc.r.CreateSession(ctx, session)
	if err != nil {
		uc.log.Error("failed to create session", "error", err)
		return nil, err
	}

	return pair, nil
}

func hashToken(token string) string {
	hashed := sha256.Sum256([]byte(token))
	hashedToken, _ := bcrypt.GenerateFromPassword(hashed[:], bcrypt.DefaultCost)

	return string(hashedToken)
}

func (uc *Usecase) sendEmail() {
	email := gomail.NewMessage()
	email.SetHeader("From", "example@example.com")
	email.SetHeader("To", "example_user@example.com")
	email.SetHeader("Subject", "Предупреждение!")
	email.SetBody("text/html", fmt.Sprintf("Ваш ip адресс сменился"))
	d := gomail.NewDialer("smtp", 465, "example@example.com", "password")
	_ = d.DialAndSend(email)
	uc.log.Info("email sent")
}
