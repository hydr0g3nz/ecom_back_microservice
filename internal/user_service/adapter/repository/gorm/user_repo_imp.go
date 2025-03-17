package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/adapter/repository/gorm/model"
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
	userModel := model.NewUserModel(&user)
	err := r.db.WithContext(ctx).Create(userModel).Error
	if err != nil {
		return nil, err
	}
	return userModel.ToEntity(), nil
}

// GetByID retrieves a user by ID
func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.ToEntity(), nil
}

// GetByEmail retrieves a user by email
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.ToEntity(), nil
}

// GetByUsername retrieves a user by username
func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.ToEntity(), nil
}

// Update updates an existing user
func (r *GormUserRepository) Update(ctx context.Context, user entity.User) (*entity.User, error) {
	userModel := model.NewUserModel(&user)
	err := r.db.WithContext(ctx).Save(userModel).Error
	if err != nil {
		return nil, err
	}
	return userModel.ToEntity(), nil
}

// Delete removes a user by ID
func (r *GormUserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

// GetByToken retrieves a user by token
func (r *GormUserRepository) GetByToken(ctx context.Context, token string) (*entity.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user.ToEntity(), nil
}
