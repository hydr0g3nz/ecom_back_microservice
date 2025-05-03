package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
)

type InventoryRepository interface {
	// Create stores a new inventory record
	Create(ctx context.Context, inventory entity.Inventory) (*entity.Inventory, error)

	// GetByProductID retrieves inventory by product ID
	GetByProductID(ctx context.Context, productID string) (*entity.Inventory, error)

	// Update updates an existing inventory
	Update(ctx context.Context, inventory entity.Inventory) (*entity.Inventory, error)

	// UpdateQuantity updates just the quantity of a product
	UpdateQuantity(ctx context.Context, productID string, quantity int) error

	// ReserveStock reserves stock for a product (for order processing)
	ReserveStock(ctx context.Context, productID string, quantity int) error

	// ReleaseReservedStock releases reserved stock back to available
	ReleaseReservedStock(ctx context.Context, productID string, quantity int) error
}
