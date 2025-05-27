package service

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
)

// EventPublisherService defines the interface for publishing inventory eventspackage event

// EventPublisher defines the methods for publishing payment-related domain events.
// This interface is used by the usecase layer to decouple it from the specific
// event messaging implementation (e.g., Kafka, RabbitMQ).
type EventPublisher interface {
	PublishPaymentCreated(ctx context.Context, evt *entity.Payment) error
	PublishPaymentUpdated(ctx context.Context, evt *entity.Payment) error
	PublishPaymentCompleted(ctx context.Context, evt *entity.Payment) error
	PublishPaymentFailed(ctx context.Context, evt *entity.PaymentFailed) error
	PublishRefundInitiated(ctx context.Context, evt *entity.Payment) error
	PublishRefundCompleted(ctx context.Context, evt *entity.Payment) error
	// Add a Close method for graceful shutdown
	Close() error
}
