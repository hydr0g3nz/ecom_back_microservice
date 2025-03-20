package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/segmentio/kafka-go"
)

// OrderEventPublisher defines the interface for publishing order events
type OrderEventPublisher interface {
	PublishOrderCreated(ctx context.Context, order *entity.Order) error
	PublishOrderUpdated(ctx context.Context, order *entity.Order) error
	PublishOrderCancelled(ctx context.Context, order *entity.Order, reason string) error
	PublishPaymentProcessed(ctx context.Context, order *entity.Order, payment *entity.Payment) error
	PublishShippingUpdated(ctx context.Context, order *entity.Order, shipping *entity.Shipping) error
}

// KafkaOrderEventPublisher implements the OrderEventPublisher interface using Kafka
type KafkaOrderEventPublisher struct {
	writer *kafka.Writer
}

// NewKafkaOrderEventPublisher creates a new instance of KafkaOrderEventPublisher
func NewKafkaOrderEventPublisher(brokers []string) *KafkaOrderEventPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaOrderEventPublisher{
		writer: writer,
	}
}

// PublishOrderCreated publishes an order created event
func (p *KafkaOrderEventPublisher) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	// Create event payload
	event := map[string]interface{}{
		"event_type": "order_created",
		"order_id":   order.ID,
		"user_id":    order.UserID,
		"items":      order.Items,
		"total":      order.TotalAmount,
		"status":     order.Status,
		"created_at": order.CreatedAt,
	}

	return p.publishEvent(ctx, "orders", order.ID, event)
}

// PublishOrderUpdated publishes an order updated event
func (p *KafkaOrderEventPublisher) PublishOrderUpdated(ctx context.Context, order *entity.Order) error {
	// Create event payload
	event := map[string]interface{}{
		"event_type": "order_updated",
		"order_id":   order.ID,
		"user_id":    order.UserID,
		"status":     order.Status,
		"updated_at": order.UpdatedAt,
		"version":    order.Version,
	}

	return p.publishEvent(ctx, "orders", order.ID, event)
}

// PublishOrderCancelled publishes an order cancelled event
func (p *KafkaOrderEventPublisher) PublishOrderCancelled(ctx context.Context, order *entity.Order, reason string) error {
	// Create event payload
	event := map[string]interface{}{
		"event_type":   "order_cancelled",
		"order_id":     order.ID,
		"user_id":      order.UserID,
		"reason":       reason,
		"cancelled_at": order.CancelledAt,
		"version":      order.Version,
	}

	return p.publishEvent(ctx, "orders", order.ID, event)
}

// PublishPaymentProcessed publishes a payment processed event
func (p *KafkaOrderEventPublisher) PublishPaymentProcessed(ctx context.Context, order *entity.Order, payment *entity.Payment) error {
	// Create event payload
	event := map[string]interface{}{
		"event_type":     "payment_processed",
		"order_id":       order.ID,
		"payment_id":     payment.ID,
		"user_id":        order.UserID,
		"amount":         payment.Amount,
		"status":         payment.Status,
		"method":         payment.Method,
		"processed_at":   payment.CompletedAt,
		"transaction_id": payment.TransactionID,
	}

	return p.publishEvent(ctx, "payments", payment.ID, event)
}

// PublishShippingUpdated publishes a shipping updated event
func (p *KafkaOrderEventPublisher) PublishShippingUpdated(ctx context.Context, order *entity.Order, shipping *entity.Shipping) error {
	// Create event payload
	event := map[string]interface{}{
		"event_type":      "shipping_updated",
		"order_id":        order.ID,
		"shipping_id":     shipping.ID,
		"user_id":         order.UserID,
		"status":          shipping.Status,
		"carrier":         shipping.Carrier,
		"tracking_number": shipping.TrackingNumber,
		"updated_at":      shipping.UpdatedAt,
	}

	return p.publishEvent(ctx, "shipping", shipping.ID, event)
}

// publishEvent publishes an event to Kafka
func (p *KafkaOrderEventPublisher) publishEvent(ctx context.Context, topic string, key string, payload interface{}) error {
	// Serialize the payload
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializing event: %w", err)
	}

	// Write message to Kafka
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})

	if err != nil {
		return fmt.Errorf("error publishing event to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka writer
func (p *KafkaOrderEventPublisher) Close() error {
	return p.writer.Close()
}
