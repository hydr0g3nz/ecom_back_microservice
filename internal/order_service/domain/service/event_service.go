// internal/order_service/domain/service/event_service.go
package service

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// Event types for order service
const (
	EventTypeOrderCreated      = "order.created"
	EventTypeOrderUpdated      = "order.updated"
	EventTypeOrderCancelled    = "order.cancelled"
	EventTypeOrderCompleted    = "order.completed"
	EventTypeReserveInventory  = "inventory.reserve"
	EventTypeReleaseInventory  = "inventory.release"
	EventTypeInventoryReserved = "inventory.reserved"
	EventTypeInventoryReleased = "inventory.released"
	EventTypePaymentRequested  = "payment.requested"
	EventTypePaymentProcessed  = "payment.processed"
	EventTypePaymentFailed     = "payment.failed"
)

// EventPublisher defines the interface for publishing events
type EventPublisherService interface {
	// PublishOrderCreated publishes an event that a new order has been created
	PublishOrderCreated(ctx context.Context, order *entity.Order) error

	// PublishOrderUpdated publishes an event that an order has been updated
	PublishOrderUpdated(ctx context.Context, order *entity.Order) error

	// PublishOrderCancelled publishes an event that an order has been cancelled
	PublishOrderCancelled(ctx context.Context, order *entity.Order) error

	// PublishOrderCompleted publishes an event that an order has been completed
	PublishOrderCompleted(ctx context.Context, order *entity.Order) error

	// PublishReserveInventory publishes a request to reserve inventory for an order
	PublishReserveInventory(ctx context.Context, order *entity.Order) error

	// PublishReleaseInventory publishes a request to release reserved inventory
	PublishReleaseInventory(ctx context.Context, order *entity.Order) error

	// PublishPaymentRequest publishes a request to process payment for an order
	PublishPaymentRequest(ctx context.Context, order *entity.Order) error
	Close() error
}

// EventSubscriber defines the interface for subscribing to events
type EventSubscriberService interface {
	// SubscribeToInventoryEvents subscribes to inventory-related events
	SubscribeToInventoryEvents(ctx context.Context) error

	// SubscribeToPaymentEvents subscribes to payment-related events
	SubscribeToPaymentEvents(ctx context.Context) error
	Close() error
}

// EventService combines both EventPublisher and EventSubscriber
