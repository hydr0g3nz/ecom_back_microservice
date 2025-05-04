package service

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
)

// Event types for inventory service
const (
	EventTypeStockUpdated             = "inventory.stock.updated"
	EventTypeStockReserved            = "inventory.stock.reserved"
	EventTypeStockReservationFailed   = "inventory.stock.reservation_failed"
	EventTypeStockReleased            = "inventory.stock.released"
	EventTypeStockDeducted            = "inventory.stock.deducted"
	EventTypeStockLow                 = "inventory.stock.low"
	EventTypeOrderReservationCreated  = "order.reservation.created"
	EventTypeOrderReservationCanceled = "order.reservation.canceled"
	EventTypeOrderReservationExpired  = "order.reservation.expired"
)

// EventPublisherService defines the interface for publishing inventory events
type EventPublisherService interface {
	// PublishStockUpdated publishes an event that stock has been updated
	PublishStockUpdated(ctx context.Context, item *entity.InventoryItem) error

	// PublishStockReserved publishes an event that stock has been reserved for an order
	PublishStockReserved(ctx context.Context, reservation *entity.InventoryReservation) error

	// PublishStockReservationFailed publishes an event that stock reservation has failed
	PublishStockReservationFailed(ctx context.Context, orderID string, sku string, reason string) error

	// PublishStockReleased publishes an event that reserved stock has been released
	PublishStockReleased(ctx context.Context, reservation *entity.InventoryReservation) error

	// PublishStockDeducted publishes an event that stock has been deducted
	PublishStockDeducted(ctx context.Context, transaction *entity.StockTransaction) error

	// PublishStockLow publishes an event that stock is below reorder level
	PublishStockLow(ctx context.Context, item *entity.InventoryItem) error

	// Close closes the publisher connections
	Close() error
}

// EventSubscriberService defines the interface for subscribing to inventory-related events
type EventSubscriberService interface {
	// SubscribeToOrderEvents subscribes to order-related events
	SubscribeToOrderEvents(ctx context.Context) error

	// HandleOrderCreated handles the event when an order is created
	HandleOrderCreated(ctx context.Context, orderData []byte) error

	// HandleOrderCancelled handles the event when an order is cancelled
	HandleOrderCancelled(ctx context.Context, orderData []byte) error

	// HandleReservationRequest handles reservation requests
	HandleReservationRequest(ctx context.Context, reservationData []byte) error

	// HandleReleaseRequest handles release requests
	HandleReleaseRequest(ctx context.Context, releaseData []byte) error

	// Close closes the subscriber connections
	Close() error
}
