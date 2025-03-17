package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
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
		errBuilder: utils.NewErrorBuilder(
			"UserUsecase",
		),
	}
}

// userUsecase implements the UserUsecase interface
type userUsecase struct {
	userRepo   repository.UserRepository
	errBuilder *utils.ErrorBuilder
}

// CreateUser creates a new user
func (uu *userUsecase) CreateUser(ctx context.Context, user *entity.User, password string) (*entity.User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, uu.errBuilder.Err(err)
	}
	user.ID = uuid.New().String()
	user.HashedPassword = string(hashedPassword)
	createdUser, err := uu.userRepo.Create(ctx, *user)
	if err != nil {
		return nil, uu.errBuilder.Err(err)
	}
	return createdUser, nil
}

// GetUserByID retrieves a user by ID
func (uu *userUsecase) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := uu.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, uu.errBuilder.Err(err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uu *userUsecase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := uu.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, uu.errBuilder.Err(err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (uu *userUsecase) UpdateUser(ctx context.Context, id string, user entity.User) (*entity.User, error) {
	updatedUser, err := uu.userRepo.Update(ctx, user)
	if err != nil {
		return nil, uu.errBuilder.Err(err)
	}
	return updatedUser, nil
}

// DeleteUser deletes a user by ID
func (uu *userUsecase) DeleteUser(ctx context.Context, id string) error {
	err := uu.userRepo.Delete(ctx, id)
	if err != nil {
		return uu.errBuilder.Err(err)
	}
	return nil
}
