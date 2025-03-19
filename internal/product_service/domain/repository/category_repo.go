package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
)

type CategoryRepository interface {
	// Create stores a new category
	Create(ctx context.Context, category entity.Category) (*entity.Category, error)

	// GetByID retrieves a category by ID
	GetByID(ctx context.Context, id string) (*entity.Category, error)

	// GetByName retrieves a category by name
	GetByName(ctx context.Context, name string) (*entity.Category, error)

	// List retrieves categories with optional filtering
	List(ctx context.Context, offset, limit int) ([]*entity.Category, int, error)

	// GetChildren retrieves child categories for a parent category
	GetChildren(ctx context.Context, parentID string) ([]*entity.Category, error)

	// Update updates an existing category
	Update(ctx context.Context, category entity.Category) (*entity.Category, error)

	// Delete removes a category by ID (soft delete)
	Delete(ctx context.Context, id string) error
}
