package usecase

import (
	"context"
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInternalServerError = errors.New("internal server error")
)

type UserUsecase interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *entity.User, password string) (*entity.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*entity.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id string, user entity.User) (*entity.User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
}

// NewUserUsecase returns a new instance of the user usecase
func NewUserUsecase(ur repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: ur,
	}
}

// userUsecase implements the UserUsecase interface
type userUsecase struct {
	userRepo repository.UserRepository
}

// CreateUser creates a new user
func (uu *userUsecase) CreateUser(ctx context.Context, user *entity.User, password string) (*entity.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user.HashedPassword = string(hashedPassword)
	return uu.userRepo.Create(ctx, *user)
}

// GetUserByID retrieves a user by ID
func (uu *userUsecase) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	return uu.userRepo.GetByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (uu *userUsecase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return uu.userRepo.GetByEmail(ctx, email)
}

// UpdateUser updates an existing user
func (uu *userUsecase) UpdateUser(ctx context.Context, id string, user entity.User) (*entity.User, error) {
	return uu.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user by ID
func (uu *userUsecase) DeleteUser(ctx context.Context, id string) error {
	return uu.userRepo.Delete(ctx, id)
}
