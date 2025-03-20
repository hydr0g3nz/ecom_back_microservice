package repository

import (
	"context"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderReadRepository defines the interface for read-only order operations (CQRS read model)
type OrderReadRepository interface {
	// GetByID retrieves an order by ID (read model)
	GetByID(ctx context.Context, id string) (*entity.Order, error)

	// GetByUserID retrieves orders for a user
	GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error)

	// FindByStatus retrieves orders with a specific status
	FindByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error)

	// FindByDateRange retrieves orders created within a date range
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.Order, int, error)

	// FindByProductID retrieves orders containing a specific product
	FindByProductID(ctx context.Context, productID string, page, pageSize int) ([]*entity.Order, int, error)

	// Search searches for orders based on various criteria
	Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error)
}
