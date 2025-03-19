package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
)

type ProductRepository interface {
	// Create stores a new product
	Create(ctx context.Context, product entity.Product) (*entity.Product, error)

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id string) (*entity.Product, error)

	// GetBySKU retrieves a product by SKU
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)

	// List retrieves products with optional filtering
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*entity.Product, int, error)

	// Update updates an existing product
	Update(ctx context.Context, product entity.Product) (*entity.Product, error)

	// Delete removes a product by ID (soft delete)
	Delete(ctx context.Context, id string) error

	// GetByCategory retrieves products by category ID
	GetByCategory(ctx context.Context, categoryID string, offset, limit int) ([]*entity.Product, int, error)
}
