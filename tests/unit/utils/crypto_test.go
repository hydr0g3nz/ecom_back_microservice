package utils_test

import (
	"testing"

	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils" // Replace with your actual module name
)

func TestHashPassword(t *testing.T) {
	// Test with valid password
	password := "securePassword123"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword returned an error for valid password: %v", err)
	}
	if hash == nil || len(hash) == 0 {
		t.Error("HashPassword returned an empty hash for valid password")
	}

	// Test with empty password
	_, err = utils.HashPassword("")
	if err == nil {
		t.Error("HashPassword did not return an error for empty password")
	}
	if err != nil && err.Error() != "password is empty" {
		t.Errorf("HashPassword returned unexpected error for empty password: %v", err)
	}

	// Test that the same password produces different hashes (due to bcrypt salt)
	hash1, _ := utils.HashPassword(password)
	hash2, _ := utils.HashPassword(password)
	if string(hash1) == string(hash2) {
		t.Error("HashPassword produced identical hashes for the same password")
	}

	// Test with a long password (within bcrypt's limits)
	longPassword := "ThisIsAReasonablyLongPasswordThatShouldWorkWithBcrypt123456789"
	hash3, err := utils.HashPassword(longPassword)
	if err != nil {
		t.Errorf("HashPassword returned an error for long password: %v", err)
	}
	if hash3 == nil || len(hash3) == 0 {
		t.Error("HashPassword returned an empty hash for long password")
	}
}

func TestVerifyPassword(t *testing.T) {
	// Test with valid password and its hash
	password := "securePassword123"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	err = utils.VerifyPassword(password, string(hash))
	if err != nil {
		t.Errorf("VerifyPassword returned an error for valid password and hash: %v", err)
	}

	// Test with incorrect password
	wrongPassword := "wrongPassword123"
	err = utils.VerifyPassword(wrongPassword, string(hash))
	if err == nil {
		t.Error("VerifyPassword did not return an error for incorrect password")
	}

	// Test with empty password
	err = utils.VerifyPassword("", string(hash))
	if err != utils.ErrEmptyPassword {
		t.Errorf("VerifyPassword returned unexpected error for empty password: %v", err)
	}

	// Test with empty hash
	err = utils.VerifyPassword(password, "")
	if err != utils.ErrEmptyHash {
		t.Errorf("VerifyPassword returned unexpected error for empty hash: %v", err)
	}
}
