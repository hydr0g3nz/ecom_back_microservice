package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository interface using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new user repository instance
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// Create stores a new user
func (r *GormUserRepository) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	err := r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user
func (r *GormUserRepository) Update(ctx context.Context, user entity.User) (*entity.User, error) {
	err := r.db.WithContext(ctx).Save(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete removes a user by ID
func (r *GormUserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.User{}).Error
}
