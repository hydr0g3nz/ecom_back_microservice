package publisher

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// OrderEventPublisher defines the interface for publishing order events
type OrderEventPublisher interface {
	// PublishOrderCreated publishes an event when an order is created
	PublishOrderCreated(ctx context.Context, order *entity.Order) error
	
	// PublishOrderUpdated publishes an event when an order is updated
	PublishOrderUpdated(ctx context.Context, order *entity.Order) error
	
	// PublishOrderCancelled publishes an event when an order is cancelled
	PublishOrderCancelled(ctx context.Context, order *entity.Order, reason string) error
	
	// PublishPaymentProcessed publishes an event when a payment is processed
	PublishPaymentProcessed(ctx context.Context, order *entity.Order, payment *entity.Payment) error
	
	// PublishRefundProcessed publishes an event when a refund is processed
	PublishRefundProcessed(ctx context.Context, order *entity.Order, refund *entity.Payment, reason string) error
	
	// PublishShippingUpdated publishes an event when shipping information is updated
	PublishShippingUpdated(ctx context.Context, order *entity.Order, shipping *entity.Shipping) error
}