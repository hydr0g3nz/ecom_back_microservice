// internal/order_service/adapter/event/event_service_impl.go
package event

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/consumer"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/producer"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// KafkaEventService combines both producer and consumer for a complete event service
type KafkaEventService struct {
	producer *producer.KafkaProducer
	consumer *consumer.KafkaConsumer
	logger   logger.Logger
}

// NewKafkaEventService creates a new KafkaEventService
func NewKafkaEventService(
	producer *producer.KafkaProducer,
	consumer *consumer.KafkaConsumer,
	logger logger.Logger,
) service.EventService {
	return &KafkaEventService{
		producer: producer,
		consumer: consumer,
		logger:   logger,
	}
}

// PublishOrderCreated publishes an event that a new order has been created
func (kes *KafkaEventService) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCreated(ctx, order)
}

// PublishOrderUpdated publishes an event that an order has been updated
func (kes *KafkaEventService) PublishOrderUpdated(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderUpdated(ctx, order)
}

// PublishOrderCancelled publishes an event that an order has been cancelled
func (kes *KafkaEventService) PublishOrderCancelled(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCancelled(ctx, order)
}

// PublishOrderCompleted publishes an event that an order has been completed
func (kes *KafkaEventService) PublishOrderCompleted(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCompleted(ctx, order)
}

// PublishReserveInventory publishes a request to reserve inventory for an order
func (kes *KafkaEventService) PublishReserveInventory(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishReserveInventory(ctx, order)
}

// PublishReleaseInventory publishes a request to release reserved inventory
func (kes *KafkaEventService) PublishReleaseInventory(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishReleaseInventory(ctx, order)
}

// PublishPaymentRequest publishes a request to process payment for an order
func (kes *KafkaEventService) PublishPaymentRequest(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishPaymentRequest(ctx, order)
}

// SubscribeToInventoryEvents subscribes to inventory-related events
func (kes *KafkaEventService) SubscribeToInventoryEvents(ctx context.Context) error {
	return kes.consumer.SubscribeToInventoryEvents(ctx)
}

// SubscribeToPaymentEvents subscribes to payment-related events
func (kes *KafkaEventService) SubscribeToPaymentEvents(ctx context.Context) error {
	return kes.consumer.SubscribeToPaymentEvents(ctx)
}

// Close closes all event connections
func (kes *KafkaEventService) Close() error {
	// Close consumer first
	if err := kes.consumer.Close(); err != nil {
		kes.logger.Error("Failed to close Kafka consumer", "error", err)
	}

	// Then close producer
	if err := kes.producer.Close(); err != nil {
		kes.logger.Error("Failed to close Kafka producer", "error", err)
	}

	return nil
}
