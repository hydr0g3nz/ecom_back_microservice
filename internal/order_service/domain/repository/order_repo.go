package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderRepository defines the interface for order persistence operations
type OrderRepository interface {
	// Create stores a new order
	Create(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id string) (*entity.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// UpdateStatus updates the status of an order
	UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus) error

	// AddItem adds an item to an order
	AddItem(ctx context.Context, orderID string, item entity.OrderItem) error

	// RemoveItem removes an item from an order
	RemoveItem(ctx context.Context, orderID string, productID string) error

	// UpdateItemQuantity updates the quantity of an item in an order
	UpdateItemQuantity(ctx context.Context, orderID string, productID string, quantity int) error

	// ApplyDiscount applies a discount to an order
	ApplyDiscount(ctx context.Context, orderID string, discount entity.Discount) error
}
