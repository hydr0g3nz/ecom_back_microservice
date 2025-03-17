package entity

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInternalServerError = errors.New("internal server error")
	ErrUserExists          = errors.New("user already exists")
	ErrTokenHasBeenRevoked = errors.New("token has been revoked")
)
