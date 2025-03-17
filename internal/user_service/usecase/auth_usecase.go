package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists = errors.New("user already exists")
)

type AuthUsecase interface {
	Register(ctx context.Context, user entity.User, password string) (*entity.User, *entity.TokenPair, error)
	Login(ctx context.Context, email, password string) (*entity.TokenPair, error)
	RefreshToken(ctx context.Context, tokenStr string) (*entity.TokenPair, error)
}

type authUsecase struct {
	userUsecase  UserUsecase
	tokenUsecase TokenUsecase
}

// NewAuthUsecase creates a new instance of AuthUsecase
func NewAuthUsecase(userUsecase UserUsecase, tokenUsecase TokenUsecase) AuthUsecase {
	return &authUsecase{
		userUsecase:  userUsecase,
		tokenUsecase: tokenUsecase,
	}
}

// Register creates a new user and generates a token
func (au *authUsecase) Register(ctx context.Context, user entity.User, password string) (*entity.User, *entity.TokenPair, error) {
	// Check if user already exists
	existingUser, err := au.userUsecase.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return nil, nil, ErrUserExists
	}

	// Create user (password handling is done in UserUsecase)
	createdUser, err := au.userUsecase.CreateUser(ctx, &user, password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token pair
	tokenPair, err := au.tokenUsecase.GenerateTokenPair(ctx, createdUser.ID, createdUser.Username, vo.User.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate token: %w", err)
	}
	return createdUser, tokenPair, nil
}

// Login authenticates a user and generates a token
func (au *authUsecase) Login(ctx context.Context, email, password string) (*entity.TokenPair, error) {
	// Get user by email
	user, err := au.userUsecase.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	comparePassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	// Verify password (assuming this is handled in the user entity or repository)
	// This is a placeholder - in a real implementation, you would use a proper password verification method
	if !verifyPassword(user.HashedPassword, string(comparePassword)) {
		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := au.tokenUsecase.GenerateTokenPair(ctx, user.ID, user.Username, user.Role.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenPair, nil
}

// RefreshToken generates a new access token using a refresh token
func (au *authUsecase) RefreshToken(ctx context.Context, tokenStr string) (*entity.TokenPair, error) {
	// Validate refresh token
	claims, err := au.tokenUsecase.ValidateToken(ctx, tokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get new access token
	newAccessToken, err := au.tokenUsecase.RefreshAccessToken(ctx, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user to populate token response
	user, err := au.userUsecase.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Construct and return the token response
	token := &entity.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: tokenStr,
	}
	fmt.Println("token", user)
	return token, nil
}

// verifyPassword checks if the provided password matches the stored hash
// This is a placeholder - in a real implementation, you would use a proper password verification method
func verifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
