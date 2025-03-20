package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// ShippingRepository defines the interface for shipping persistence operations
type ShippingRepository interface {
	// Create stores a new shipping record
	Create(ctx context.Context, shipping *entity.Shipping) (*entity.Shipping, error)

	// GetByID retrieves a shipping record by ID
	GetByID(ctx context.Context, id string) (*entity.Shipping, error)

	// GetByOrderID retrieves shipping for an order
	GetByOrderID(ctx context.Context, orderID string) (*entity.Shipping, error)

	// Update updates an existing shipping record
	Update(ctx context.Context, shipping *entity.Shipping) (*entity.Shipping, error)

	// UpdateStatus updates the status of a shipping record
	UpdateStatus(ctx context.Context, id string, status valueobject.ShippingStatus) error

	// UpdateTrackingInfo updates the tracking information for a shipment
	UpdateTrackingInfo(ctx context.Context, id string, carrier string, trackingNumber string) error
}
