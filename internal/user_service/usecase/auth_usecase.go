package usecase

import (
	"context"
	"fmt"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

type AuthUsecase interface {
	Register(ctx context.Context, user entity.User, password string) (*entity.User, *entity.TokenPair, error)
	Login(ctx context.Context, email, password string) (*entity.TokenPair, error)
	RefreshToken(ctx context.Context, tokenStr string) (*entity.TokenPair, error)
}

type authUsecase struct {
	userUsecase  UserUsecase
	tokenUsecase TokenUsecase
	errBuilder   *utils.ErrorBuilder
}

// NewAuthUsecase creates a new instance of AuthUsecase
func NewAuthUsecase(userUsecase UserUsecase, tokenUsecase TokenUsecase) AuthUsecase {

	return &authUsecase{
		userUsecase:  userUsecase,
		tokenUsecase: tokenUsecase,
		errBuilder:   utils.NewErrorBuilder("AuthUsecase"),
	}
}

// Register creates a new user and generates a token
func (au *authUsecase) Register(ctx context.Context, user entity.User, password string) (*entity.User, *entity.TokenPair, error) {
	// Check if user already exists
	existingUser, err := au.userUsecase.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return nil, nil, au.errBuilder.Err(entity.ErrUserExists)
	}
	user.Role = vo.User
	// Create user (password handling is done in UserUsecase)
	createdUser, err := au.userUsecase.CreateUser(ctx, &user, password)
	if err != nil {
		return nil, nil, au.errBuilder.Err(err)
	}

	// Generate token pair
	tokenPair, err := au.tokenUsecase.GenerateTokenPair(ctx, createdUser.ID, vo.User.String())
	if err != nil {
		return nil, nil, au.errBuilder.Err(err)
	}
	return createdUser, tokenPair, nil
}

// Login authenticates a user and generates a token
func (au *authUsecase) Login(ctx context.Context, email, password string) (*entity.TokenPair, error) {
	// Get user by email
	user, err := au.userUsecase.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, entity.ErrInvalidCredentials
	}
	// Verify password (assuming this is handled in the user entity or repository)
	// This is a placeholder - in a real implementation, you would use a proper password verification method
	if err := utils.VerifyPassword(password, user.HashedPassword); err != nil {
		fmt.Println("error", err)
		return nil, entity.ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := au.tokenUsecase.GenerateTokenPair(ctx, user.ID, user.Role.String())
	if err != nil {
		return nil, au.errBuilder.Err(err)
	}

	return tokenPair, nil
}

// RefreshToken generates a new access token using a refresh token
func (au *authUsecase) RefreshToken(ctx context.Context, tokenStr string) (*entity.TokenPair, error) {
	// Validate refresh token
	_, err := au.tokenUsecase.ValidateToken(ctx, tokenStr)
	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	// Get new access token
	newAccessToken, err := au.tokenUsecase.RefreshAccessToken(ctx, tokenStr)
	if err != nil {
		return nil, au.errBuilder.Err(err)
	}

	// // Get user to populate token response
	// user, err := au.userUsecase.GetUserByID(ctx, claims.UserID)
	// if err != nil {
	// 	return nil, au.errBuilder.Err(err)
	// }

	// Construct and return the token response
	token := &entity.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: tokenStr,
	}
	return token, nil
}
