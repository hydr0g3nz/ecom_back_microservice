package event

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// Publisher defines the interface for publishing domain events
type Publisher interface {
	// Publish publishes a domain event
	Publish(ctx context.Context, event entity.DomainEvent) error
	
	// PublishAll publishes multiple domain events
	PublishAll(ctx context.Context, events []entity.DomainEvent) error
	
	// Close closes the publisher
	Close() error
}

// OrderEventPublisher defines the interface for publishing order-specific events
type OrderEventPublisher interface {
	// PublishOrderCreated publishes an order created event
	PublishOrderCreated(ctx context.Context, event entity.DomainEvent) error
	
	// PublishOrderUpdated publishes an order updated event
	PublishOrderUpdated(ctx context.Context, event entity.DomainEvent) error
	
	// PublishOrderCancelled publishes an order cancelled event
	PublishOrderCancelled(ctx context.Context, event entity.DomainEvent) error
	
	// PublishOrderCompleted publishes an order completed event
	PublishOrderCompleted(ctx context.Context, event entity.DomainEvent) error
	
	// PublishPaymentProcessed publishes a payment processed event
	PublishPaymentProcessed(ctx context.Context, event entity.DomainEvent) error
	
	// PublishShippingUpdated publishes a shipping updated event
	PublishShippingUpdated(ctx context.Context, event entity.DomainEvent) error
	
	// Close closes the publisher
	Close() error
}
