package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// PaymentRepository defines the interface for payment persistence operations
type PaymentRepository interface {
	// Create stores a new payment
	Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)

	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id string) (*entity.Payment, error)

	// GetByOrderID retrieves payments for an order
	GetByOrderID(ctx context.Context, orderID string) ([]*entity.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)

	// UpdateStatus updates the status of a payment
	UpdateStatus(ctx context.Context, id string, status valueobject.PaymentStatus) error
}
