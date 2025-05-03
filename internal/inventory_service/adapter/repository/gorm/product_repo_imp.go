package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"gorm.io/gorm"
)

// GormProductRepository implements ProductRepository interface using GORM
type GormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository creates a new instance of GormProductRepository
func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

// Create stores a new product
func (r *GormProductRepository) Create(ctx context.Context, product entity.Product) (*entity.Product, error) {
	productModel := model.NewProductModel(&product)
	err := r.db.WithContext(ctx).Create(productModel).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "sku") {
			return nil, entity.ErrProductSKUExists
		}
		return nil, err
	}
	return productModel.ToEntity(), nil
}

// GetByID retrieves a product by ID
func (r *GormProductRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
	var productModel model.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&productModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrProductNotFound
		}
		return nil, err
	}
	return productModel.ToEntity(), nil
}

// GetBySKU retrieves a product by SKU
func (r *GormProductRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var productModel model.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&productModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrProductNotFound
		}
		return nil, err
	}
	return productModel.ToEntity(), nil
}

// List retrieves products with optional filtering
func (r *GormProductRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*entity.Product, int, error) {
	var productModels []model.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Product{})

	// Apply filters if any
	if len(filters) > 0 {
		for key, value := range filters {
			if key == "status" {
				query = query.Where("status = ?", value)
			} else if key == "category_id" {
				query = query.Where("category_id = ?", value)
			} else if key == "name" {
				query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", value))
			} else if key == "price_min" {
				query = query.Where("price >= ?", value)
			} else if key == "price_max" {
				query = query.Where("price <= ?", value)
			}
		}
	}

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&productModels).Error; err != nil {
		return nil, 0, err
	}

	// Convert to entities
	products := make([]*entity.Product, len(productModels))
	for i, productModel := range productModels {
		products[i] = productModel.ToEntity()
	}

	return products, int(total), nil
}

// Update updates an existing product
func (r *GormProductRepository) Update(ctx context.Context, product entity.Product) (*entity.Product, error) {
	// Check if product exists
	existingProduct, err := r.GetByID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	// Create product model
	productModel := model.NewProductModel(&product)

	// Preserve created_at timestamp
	productModel.CreatedAt = existingProduct.CreatedAt

	// Update product
	err = r.db.WithContext(ctx).Save(productModel).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "sku") {
			return nil, entity.ErrProductSKUExists
		}
		return nil, err
	}

	return productModel.ToEntity(), nil
}

// Delete removes a product by ID (soft delete)
func (r *GormProductRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Product{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrProductNotFound
	}
	return nil
}

// GetByCategory retrieves products by category ID
func (r *GormProductRepository) GetByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*entity.Product, int, error) {
	var productModels []model.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Product{}).Where("category_id = ?", categoryID)

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&productModels).Error; err != nil {
		return nil, 0, err
	}

	// Convert to entities
	products := make([]*entity.Product, len(productModels))
	for i, productModel := range productModels {
		products[i] = productModel.ToEntity()
	}

	return products, int(total), nil
}
