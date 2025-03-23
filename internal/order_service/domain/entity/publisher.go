package entity

import (
	"context"
)

// OrderEventPublisher defines the interface for publishing order events
type OrderEventPublisher interface {
	// PublishOrderCreated publishes an event when an order is created
	PublishOrderCreated(ctx context.Context, order *Order) error

	// PublishOrderUpdated publishes an event when an order is updated
	PublishOrderUpdated(ctx context.Context, order *Order) error

	// PublishOrderCancelled publishes an event when an order is cancelled
	PublishOrderCancelled(ctx context.Context, order *Order, reason string) error

	// PublishPaymentProcessed publishes an event when a payment is processed
	PublishPaymentProcessed(ctx context.Context, order *Order, payment *Payment) error

	// PublishShippingUpdated publishes an event when shipping information is updated
	PublishShippingUpdated(ctx context.Context, order *Order, shipping *Shipping) error

	// Close closes the publisher resources
	Close() error
}
