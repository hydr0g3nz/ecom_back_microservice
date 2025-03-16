package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
)

type UserRepository interface {
	// Create stores a new user
	Create(ctx context.Context, user entity.User) (*entity.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByentity.Username retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user entity.User) (*entity.User, error)

	// Delete removes a user by ID
	Delete(ctx context.Context, id string) error
}
