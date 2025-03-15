package port

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
)

// UserService defines the interface for the user business logic
type UserUsecase interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user entity.User, password string) (*entity.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id string) (*entity.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id string, user entity.User) (*entity.User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
}

// UserRepository defines the interface for user data storage
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
