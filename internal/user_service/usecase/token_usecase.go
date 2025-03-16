package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/jwt_service"
	"gorm.io/gorm"
)

// TokenUsecase defines the interface for token operations
type TokenUsecase interface {
	GenerateTokenPair(ctx context.Context, userID, username, role string) (*entity.TokenPair, error)
	ValidateToken(ctx context.Context, tokenValue string) (*jwt_service.CustomClaims, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	RevokeToken(ctx context.Context, tokenValue string) error
}

type tokenUsecase struct {
	tokenRepo  repository.TokenRepository
	jwtService jwt_service.TokenService
}

// NewTokenUsecase creates a new instance of TokenUsecase
func NewTokenUsecase(tokenRepo repository.TokenRepository, jwtService jwt_service.TokenService) TokenUsecase {
	return &tokenUsecase{
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (tu *tokenUsecase) GenerateTokenPair(ctx context.Context, userID, username, role string) (*entity.TokenPair, error) {
	// Generate tokens using the JWT service
	accessToken, err := tu.jwtService.GenerateAccessToken(userID, username, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tu.jwtService.GenerateRefreshToken(userID, username, role)
	if err != nil {
		return nil, err
	}

	// Validate the tokens to get expiration times
	accessClaims, err := tu.jwtService.ValidateToken(accessToken)
	if err != nil {
		return nil, err
	}

	refreshClaims, err := tu.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Store tokens in repository
	accessTokenEntity := &entity.Token{
		UserID:    userID,
		Type:      entity.AccessToken,
		Token:     accessToken,
		ExpiresAt: time.Unix(accessClaims.ExpiresAt.Unix(), 0),
	}

	refreshTokenEntity := &entity.Token{
		UserID:    userID,
		Type:      entity.RefreshToken,
		Token:     refreshToken,
		ExpiresAt: time.Unix(refreshClaims.ExpiresAt.Unix(), 0),
	}

	// Save tokens to repository
	if err := tu.tokenRepo.Create(ctx, accessTokenEntity); err != nil {
		return nil, err
	}

	if err := tu.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		// Try to clean up the access token if refresh token creation fails
		_ = tu.tokenRepo.Delete(ctx, accessTokenEntity.Token)
		return nil, err
	}

	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates a token
func (tu *tokenUsecase) ValidateToken(ctx context.Context, tokenValue string) (*jwt_service.CustomClaims, error) {
	// First check if the token exists in the repository
	token, err := tu.tokenRepo.GetByToken(ctx, tokenValue)
	if err != nil {
		return nil, err
	}

	// Check if token has been revoked or expired in the database
	if errors.Is(err, gorm.ErrRecordNotFound) || token.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has been revoked")
	}

	// Then verify the JWT token
	claims, err := tu.jwtService.ValidateToken(tokenValue)
	if err != nil {
		// If token is expired, we should remove it from the repository
		if errors.Is(err, jwt_service.ErrExpiredToken) {
			_ = tu.tokenRepo.Delete(ctx, tokenValue)
		}
		return nil, err
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (tu *tokenUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate the refresh token
	claims, err := tu.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	// Check if it's actually a refresh token
	if claims.TokenType != jwt_service.RefreshToken {
		return "", errors.New("not a refresh token")
	}

	// Generate a new access token
	accessToken, err := tu.jwtService.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", err
	}

	// Validate the new token to get expiration time
	accessClaims, err := tu.jwtService.ValidateToken(accessToken)
	if err != nil {
		return "", err
	}

	// Store the new access token
	newAccessToken := &entity.Token{
		UserID:    claims.UserID,
		Type:      entity.AccessToken,
		Token:     accessToken,
		ExpiresAt: time.Unix(accessClaims.ExpiresAt.Unix(), 0),
	}

	if err := tu.tokenRepo.Create(ctx, newAccessToken); err != nil {
		return "", err
	}

	return accessToken, nil
}

// RevokeToken revokes a token
func (tu *tokenUsecase) RevokeToken(ctx context.Context, id string) error {

	return tu.tokenRepo.Delete(ctx, id)
}
