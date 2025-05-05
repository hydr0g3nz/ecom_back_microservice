package event

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/consumer"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/event/producer"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// KafkaEventPublisherService implements EventPublisherService using Kafka.
type KafkaEventPublisherService struct {
	producer *producer.KafkaProducer
	logger   logger.Logger
}

// KafkaEventSubscriberService implements EventSubscriberService using Kafka
type KafkaEventSubscriberService struct {
	consumer *consumer.KafkaConsumer
	logger   logger.Logger
}

// NewKafkaEventPublisherService creates a new KafkaEventPublisherService.
func NewKafkaEventPublisherService(producer *producer.KafkaProducer, logger logger.Logger) service.EventPublisherService {
	return &KafkaEventPublisherService{
		producer: producer,
		logger:   logger,
	}
}

// NewKafkaEventSubscriberService creates a new KafkaEventSubscriberService.
func NewKafkaEventSubscriberService(consumer *consumer.KafkaConsumer, logger logger.Logger) service.EventSubscriberService {
	return &KafkaEventSubscriberService{
		consumer: consumer,
		logger:   logger,
	}
}

// PublishOrderCreated publishes an event that a new order has been created.
func (kes *KafkaEventPublisherService) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCreated(ctx, order)
}

// PublishOrderUpdated publishes an event that an order has been updated.
func (kes *KafkaEventPublisherService) PublishOrderUpdated(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderUpdated(ctx, order)
}

// PublishOrderCancelled publishes an event that an order has been cancelled.
func (kes *KafkaEventPublisherService) PublishOrderCancelled(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCancelled(ctx, order)
}

// PublishOrderCompleted publishes an event that an order has been completed.
func (kes *KafkaEventPublisherService) PublishOrderCompleted(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishOrderCompleted(ctx, order)
}

// // PublishReserveInventory publishes a request to reserve inventory for an order.
// func (kes *KafkaEventPublisherService) PublishReserveInventory(ctx context.Context, order *entity.Order) error {
// 	return kes.producer.PublishReserveInventory(ctx, order)
// }

// // PublishReleaseInventory publishes a request to release reserved inventory.
// func (kes *KafkaEventPublisherService) PublishReleaseInventory(ctx context.Context, order *entity.Order) error {
// 	return kes.producer.PublishReleaseInventory(ctx, order)
// }

// PublishPaymentRequest publishes a request to process payment for an order.
func (kes *KafkaEventPublisherService) PublishPaymentRequest(ctx context.Context, order *entity.Order) error {
	return kes.producer.PublishPaymentRequest(ctx, order)
}

// Close closes the Kafka producer.
func (kes *KafkaEventPublisherService) Close() error {
	if err := kes.producer.Close(); err != nil {
		kes.logger.Error("Failed to close Kafka producer", "error", err)
		return err
	}
	return nil
}

// SubscribeToInventoryEvents subscribes to inventory-related events.
func (kes *KafkaEventSubscriberService) SubscribeToInventoryEvents(ctx context.Context) error {
	return kes.consumer.SubscribeToInventoryEvents(ctx)
}

// SubscribeToPaymentEvents subscribes to payment-related events.
func (kes *KafkaEventSubscriberService) SubscribeToPaymentEvents(ctx context.Context) error {
	return kes.consumer.SubscribeToPaymentEvents(ctx)
}

// Close closes the Kafka consumer.
func (kes *KafkaEventSubscriberService) Close() error {
	if err := kes.consumer.Close(); err != nil {
		kes.logger.Error("Failed to close Kafka consumer", "error", err)
		return err
	}
	return nil
}
