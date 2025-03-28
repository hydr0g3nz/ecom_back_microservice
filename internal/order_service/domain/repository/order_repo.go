// internal/order_service/domain/repository/order_repo.go
package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

type OrderRepository interface {
	// Create stores a new order
	Create(ctx context.Context, order entity.Order) (*entity.Order, error)

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id string) (*entity.Order, error)

	// GetByUserID retrieves orders for a specific user
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Order, int, error)

	// List retrieves orders with optional filtering
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*entity.Order, int, error)

	// Update updates an existing order
	Update(ctx context.Context, order entity.Order) (*entity.Order, error)

	// UpdateStatus updates the status of an order
	UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus, comment string) (*entity.Order, error)

	// Delete removes an order by ID (soft delete or mark as cancelled)
	Delete(ctx context.Context, id string) error
}
