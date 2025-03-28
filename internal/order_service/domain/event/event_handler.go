package event

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// Handler defines the interface for handling domain events
type Handler interface {
	// Handle handles a domain event
	Handle(ctx context.Context, event entity.DomainEvent) error
	
	// EventType returns the event type this handler can handle
	EventType() entity.EventType
}

// InventoryEventHandler defines the interface for handling inventory events
type InventoryEventHandler interface {
	// HandleStockReservationFailed handles a stock reservation failure event
	HandleStockReservationFailed(ctx context.Context, eventData []byte) error
	
	// HandleStockReleased handles a stock release event
	HandleStockReleased(ctx context.Context, eventData []byte) error
}

// PaymentEventHandler defines the interface for handling payment events
type PaymentEventHandler interface {
	// HandlePaymentSuccess handles a payment success event
	HandlePaymentSuccess(ctx context.Context, eventData []byte) error
	
	// HandlePaymentFailure handles a payment failure event
	HandlePaymentFailure(ctx context.Context, eventData []byte) error
}

// ShippingEventHandler defines the interface for handling shipping events
type ShippingEventHandler interface {
	// HandleShippingCreated handles a shipping created event
	HandleShippingCreated(ctx context.Context, eventData []byte) error
	
	// HandleShippingUpdated handles a shipping updated event
	HandleShippingUpdated(ctx context.Context, eventData []byte) error
	
	// HandleShippingDelivered handles a shipping delivered event
	HandleShippingDelivered(ctx context.Context, eventData []byte) error
}

// OrderEventHandler defines the interface for handling order events
type OrderEventHandler interface {
	// HandleOrderCreated handles an order created event
	HandleOrderCreated(ctx context.Context, eventData []byte) error
	
	// HandleOrderUpdated handles an order updated event
	HandleOrderUpdated(ctx context.Context, eventData []byte) error
	
	// HandleOrderCancelled handles an order cancelled event
	HandleOrderCancelled(ctx context.Context, eventData []byte) error
	
	// HandleOrderCompleted handles an order completed event
	HandleOrderCompleted(ctx context.Context, eventData []byte) error
}
