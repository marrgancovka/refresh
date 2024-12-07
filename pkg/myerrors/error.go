package myerrors

import "errors"

var (
	ErrInvalidToken              = errors.New("invalid token")
	ErrTokenExpired              = errors.New("token expired")
	ErrInappropriateRefreshToken = errors.New("inappropriate refresh token")
)
