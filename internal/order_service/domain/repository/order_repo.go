package repository

import (
	"context"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderRepository defines the interface for order data persistence
type OrderRepository interface {
	// Create saves a new order in the repository
	Create(ctx context.Context, order *entity.Order) error
	
	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	
	// GetByUserID retrieves all orders for a specific user
	GetByUserID(ctx context.Context, userID string) ([]*entity.Order, error)
	
	// Update updates an existing order
	Update(ctx context.Context, order *entity.Order) error
	
	// Delete removes an order by its ID
	Delete(ctx context.Context, id string) error
	
	// UpdateStatus updates the status of an order
	UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus) error
	
	// ListByStatus retrieves orders by their status
	ListByStatus(ctx context.Context, status valueobject.OrderStatus) ([]*entity.Order, error)
	
	// GetOrdersPaginated retrieves orders with pagination
	GetOrdersPaginated(ctx context.Context, page, pageSize int) ([]*entity.Order, int, error)
}
