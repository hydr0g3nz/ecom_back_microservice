package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Errors
var (
	ErrEmptyPassword = errors.New("password cannot be empty")
	ErrEmptyHash     = errors.New("hash cannot be empty")
)

func HashPassword(password string) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("password is empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}
func VerifyPassword(plainPassword, hashedPassword string) error {
	if plainPassword == "" {
		return ErrEmptyPassword
	}

	if hashedPassword == "" {
		return ErrEmptyHash
	}

	// CompareHashAndPassword returns nil on success, or an error on failure
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
