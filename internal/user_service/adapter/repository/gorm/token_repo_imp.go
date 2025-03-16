package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
)

// GormTokenRepository implements TokenRepository interface using GORM
type GormTokenRepository struct {
	db *gorm.DB
}

// NewGormTokenRepository creates a new instance of GormTokenRepository
func NewGormTokenRepository(db *gorm.DB) *GormTokenRepository {
	return &GormTokenRepository{db: db}
}

// FindByUserID retrieves a list of tokens by user ID
func (r *GormTokenRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Token, error) {
	var tokens []*entity.Token
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokens).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tokens, nil
}

// FindByValue retrieves a token by value
func (r *GormTokenRepository) FindByValue(ctx context.Context, value string) (*entity.Token, error) {
	var token entity.Token
	err := r.db.WithContext(ctx).
		Where("token_str = ?", value).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// Save saves a new token
func (r *GormTokenRepository) Save(ctx context.Context, token *entity.Token) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// Delete removes a token by ID
func (r *GormTokenRepository) Delete(ctx context.Context, tokenID string) error {
	return r.db.WithContext(ctx).Where("id = ?", tokenID).Delete(&entity.Token{}).Error
}
func (r *GormTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.Token{}).Error
}
func (r *GormTokenRepository) Update(ctx context.Context, by, token entity.Token) error {
	return r.db.WithContext(ctx).Where("id = ", by.ID).Updates(token).Error
}
func (r *GormTokenRepository) Create(ctx context.Context, token *entity.Token) error {
	return r.db.WithContext(ctx).Create(token).Error
}
func (r *GormTokenRepository) GetByToken(ctx context.Context, tokenStr string) (*entity.Token, error) {
	var token entity.Token
	err := r.db.WithContext(ctx).Where("token_str = ?", tokenStr).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}
