package jwt_service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Common errors
var (
	ErrInvalidToken         = errors.New("token is invalid")
	ErrExpiredToken         = errors.New("token has expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrEmptyClaims          = errors.New("empty claims provided")
)

// Config holds the JWT configuration
type Config struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// TokenType defines the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// CustomClaims extends standard JWT claims with custom app fields
type CustomClaims struct {
	jwt.RegisteredClaims
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
}

// TokenService defines the JWT service interface
type TokenService interface {
	GenerateAccessToken(userID, username, role string) (string, error)
	GenerateRefreshToken(userID, username, role string) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
	GetTokenFromBearerString(bearerToken string) string
}

// JWTService implements TokenService
type JWTService struct {
	config Config
}

// NewJWTService creates a new instance of JWTService
func NewJWTService(config Config) TokenService {
	return &JWTService{
		config: config,
	}
}

// GenerateAccessToken generates a new access token
func (s *JWTService) GenerateAccessToken(userID, username, role string) (string, error) {
	return s.generateToken(userID, username, role, s.config.AccessTokenDuration, AccessToken)
}

// GenerateRefreshToken generates a new refresh token
func (s *JWTService) GenerateRefreshToken(userID, username, role string) (string, error) {
	return s.generateToken(userID, username, role, s.config.RefreshTokenDuration, RefreshToken)
}

// generateToken creates a new token with provided claims
func (s *JWTService) generateToken(userID, username, role string, duration time.Duration, tokenType TokenType) (string, error) {
	if userID == "" {
		return "", ErrEmptyClaims
	}

	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates the token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetTokenFromBearerString extracts the token from an Authorization header value
func (s *JWTService) GetTokenFromBearerString(bearerToken string) string {
	const bearerPrefix = "Bearer "
	if len(bearerToken) > len(bearerPrefix) && bearerToken[:len(bearerPrefix)] == bearerPrefix {
		return bearerToken[len(bearerPrefix):]
	}
	return bearerToken
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID, username, role string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, username, role)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(userID, username, role)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
