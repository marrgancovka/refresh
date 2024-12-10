package http

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"refresh/internal/models"
	mock_auth "refresh/internal/pkg/auth/mocks"
	"refresh/pkg/logger"
	"refresh/pkg/myerrors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandler_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_auth.NewMockUsecase(ctrl)
	handler := &Handler{
		uc:  mockUsecase,
		log: logger.SetupLogger(),
	}

	tests := []struct {
		name         string
		query        string
		setupMocks   func()
		expectedCode int
	}{
		{
			name:  "Success case",
			query: "?id=550e8400-e29b-41d4-a716-446655440000",
			setupMocks: func() {
				mockUsecase.EXPECT().
					Authenticate(context.Background(), &models.TokenPayload{
						UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						UserIP: "127.0.0.1:8080",
					}).
					Return(&models.PairToken{
						AccessToken:     "access_token",
						RefreshToken:    "refresh_token",
						ExpRefreshToken: time.Now().Add(time.Hour),
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Missing id",
			query:        "",
			setupMocks:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid id format",
			query:        "?id=invalid-uuid",
			setupMocks:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "Error from Usecase",
			query: "?id=550e8400-e29b-41d4-a716-446655440000",
			setupMocks: func() {
				mockUsecase.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("usecase error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest(http.MethodGet, "/authenticate"+tt.query, nil)
			req.RemoteAddr = "127.0.0.1:8080"
			rec := httptest.NewRecorder()

			handler.Authenticate(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_auth.NewMockUsecase(ctrl)
	handler := &Handler{
		uc:  mockUsecase,
		log: logger.SetupLogger(),
	}
	tests := []struct {
		name         string
		cookie       *http.Cookie
		setupMocks   func()
		expectedCode int
	}{
		{
			name: "Success case",
			cookie: &http.Cookie{
				Name:  RefreshCookieName,
				Value: "valid_refresh_token",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().
					Refresh(context.Background(), "valid_refresh_token", "127.0.0.1:8080").
					Return(&models.PairToken{
						AccessToken:     "access_token",
						RefreshToken:    "refresh_token",
						ExpRefreshToken: time.Now().Add(time.Hour),
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Missing cookie",
			cookie:       nil,
			setupMocks:   func() {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid refresh token",
			cookie: &http.Cookie{
				Name:  RefreshCookieName,
				Value: "invalid_refresh_token",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().
					Refresh(gomock.Any(), "invalid_refresh_token", gomock.Any()).
					Return(nil, myerrors.ErrInvalidToken)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Unexpected error",
			cookie: &http.Cookie{
				Name:  RefreshCookieName,
				Value: "valid_refresh_token",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().
					Refresh(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unexpected error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			req.RemoteAddr = "127.0.0.1:8080"
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			handler.Refresh(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}
