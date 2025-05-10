package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"gorm.io/gorm"
)

// GormCategoryRepository implements CategoryRepository interface using GORM
type GormCategoryRepository struct {
	db *gorm.DB
}

// NewGormCategoryRepository creates a new instance of GormCategoryRepository
func NewGormCategoryRepository(db *gorm.DB) *GormCategoryRepository {
	return &GormCategoryRepository{db: db}
}

// Create stores a new category
func (r *GormCategoryRepository) Create(ctx context.Context, category entity.Category) (*entity.Category, error) {
	categoryModel := model.NewCategoryModel(&category)
	err := r.db.WithContext(ctx).Create(categoryModel).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "name") {
			return nil, entity.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return categoryModel.ToEntity(), nil
}

// GetByID retrieves a category by ID
func (r *GormCategoryRepository) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	var categoryModel model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&categoryModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCategoryNotFound
		}
		return nil, err
	}
	return categoryModel.ToEntity(), nil
}

// GetByName retrieves a category by name
func (r *GormCategoryRepository) GetByName(ctx context.Context, name string) (*entity.Category, error) {
	var categoryModel model.Category
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&categoryModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCategoryNotFound
		}
		return nil, err
	}
	return categoryModel.ToEntity(), nil
}

// List retrieves categories with optional filtering
func (r *GormCategoryRepository) List(ctx context.Context, offset, limit int) ([]*entity.Category, int, error) {
	var categoryModels []model.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Category{})

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("level ASC, name ASC").Find(&categoryModels).Error; err != nil {
		return nil, 0, err
	}

	// Convert to entities
	categories := make([]*entity.Category, len(categoryModels))
	for i, categoryModel := range categoryModels {
		categories[i] = categoryModel.ToEntity()
	}

	return categories, int(total), nil
}

// GetChildren retrieves child categories for a parent category
func (r *GormCategoryRepository) GetChildren(ctx context.Context, parentID string) ([]*entity.Category, error) {
	var categoryModels []model.Category

	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("name ASC").Find(&categoryModels).Error; err != nil {
		return nil, err
	}

	// Convert to entities
	categories := make([]*entity.Category, len(categoryModels))
	for i, categoryModel := range categoryModels {
		categories[i] = categoryModel.ToEntity()
	}

	return categories, nil
}

// Update updates an existing category
func (r *GormCategoryRepository) Update(ctx context.Context, category entity.Category) (*entity.Category, error) {
	// Check if category exists
	existingCategory, err := r.GetByID(ctx, category.ID)
	if err != nil {
		return nil, err
	}

	// Create category model
	categoryModel := model.NewCategoryModel(&category)

	// Preserve created_at timestamp
	categoryModel.CreatedAt = existingCategory.CreatedAt

	// Update category
	err = r.db.WithContext(ctx).Save(categoryModel).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "name") {
			return nil, entity.ErrCategoryAlreadyExists
		}
		return nil, err
	}

	return categoryModel.ToEntity(), nil
}

// Delete removes a category by ID (soft delete)
func (r *GormCategoryRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Category{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrCategoryNotFound
	}
	return nil
}
