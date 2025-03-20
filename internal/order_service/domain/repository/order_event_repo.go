package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// OrderEventRepository defines the interface for event sourcing operations
type OrderEventRepository interface {
	// SaveEvent stores a new order event
	SaveEvent(ctx context.Context, event *entity.OrderEvent) error

	// GetEventsByOrderID retrieves all events for an order
	GetEventsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderEvent, error)

	// GetEventsByType retrieves events of a specific type
	GetEventsByType(ctx context.Context, eventType entity.EventType) ([]*entity.OrderEvent, error)

	// GetEventsByOrderIDAndType retrieves events for an order of a specific type
	GetEventsByOrderIDAndType(ctx context.Context, orderID string, eventType entity.EventType) ([]*entity.OrderEvent, error)

	// GetLatestEventByOrderID retrieves the most recent event for an order
	GetLatestEventByOrderID(ctx context.Context, orderID string) (*entity.OrderEvent, error)

	// RebuildOrderFromEvents reconstructs an order from its event history
	RebuildOrderFromEvents(ctx context.Context, orderID string) (*entity.Order, error)
}
