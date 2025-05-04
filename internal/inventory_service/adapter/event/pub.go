// NewKafkaEventPublisher creates a new Kafka event publisher
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/service"
	"github.com/segmentio/kafka-go"
)

func NewKafkaEventPublisher(config *KafkaConfig) (*KafkaEventPublisher, error) {
	// Writer for inventory events
	w := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.InventoryTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 50 * time.Millisecond,
		// Configure additional settings as needed
	}

	// Writer for order-related events (separate topic)
	ow := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.OrderTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 50 * time.Millisecond,
	}

	return &KafkaEventPublisher{
		writer:       w,
		orderWriter:  ow,
		kafkaConfig:  config,
		serviceState: "ready",
	}, nil
}

// StockEventPayload represents the common payload structure for stock events
type StockEventPayload struct {
	EventType     string                 `json:"event_type"`
	Timestamp     time.Time              `json:"timestamp"`
	SKU           string                 `json:"sku,omitempty"`
	OrderID       string                 `json:"order_id,omitempty"`
	Quantity      int                    `json:"quantity,omitempty"`
	ReorderLevel  int                    `json:"reorder_level,omitempty"`
	AvailableQty  int                    `json:"available_qty,omitempty"`
	ReservationID string                 `json:"reservation_id,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// serializeAndPublish serializes an event payload and publishes it to Kafka
func (k *KafkaEventPublisher) serializeAndPublish(ctx context.Context, payload StockEventPayload, orderRelated bool) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize event payload: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(payload.SKU),
		Value: payloadBytes,
		Time:  time.Now(),
	}

	var writer *kafka.Writer
	if orderRelated {
		writer = k.orderWriter
	} else {
		writer = k.writer
	}

	if err := writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return nil
}

// PublishStockUpdated publishes an event that stock has been updated
func (k *KafkaEventPublisher) PublishStockUpdated(ctx context.Context, item *entity.InventoryItem) error {
	payload := StockEventPayload{
		EventType:    service.EventTypeStockUpdated,
		Timestamp:    time.Now(),
		SKU:          item.SKU,
		AvailableQty: item.AvailableQty,
		ReorderLevel: item.ReorderLevel,
		Data: map[string]interface{}{
			"inventory_item": item,
		},
	}

	return k.serializeAndPublish(ctx, payload, false)
}

// PublishStockReserved publishes an event that stock has been reserved for an order
func (k *KafkaEventPublisher) PublishStockReserved(ctx context.Context, reservation *entity.InventoryReservation) error {
	payload := StockEventPayload{
		EventType:     service.EventTypeStockReserved,
		Timestamp:     time.Now(),
		SKU:           reservation.SKU,
		OrderID:       reservation.OrderID,
		Quantity:      reservation.Qty,
		ReservationID: reservation.ReservationID,
		Data: map[string]interface{}{
			"reservation": reservation,
		},
	}

	// Publish to both inventory topic and order topic
	if err := k.serializeAndPublish(ctx, payload, false); err != nil {
		return err
	}

	// Also publish to order topic
	orderPayload := StockEventPayload{
		EventType:     service.EventTypeOrderReservationCreated,
		Timestamp:     time.Now(),
		SKU:           reservation.SKU,
		OrderID:       reservation.OrderID,
		Quantity:      reservation.Qty,
		ReservationID: reservation.ReservationID,
		Data: map[string]interface{}{
			"reservation": reservation,
		},
	}

	return k.serializeAndPublish(ctx, orderPayload, true)
}

// PublishStockReservationFailed publishes an event that stock reservation has failed
func (k *KafkaEventPublisher) PublishStockReservationFailed(ctx context.Context, orderID string, sku string, reason string) error {
	payload := StockEventPayload{
		EventType: service.EventTypeStockReservationFailed,
		Timestamp: time.Now(),
		SKU:       sku,
		OrderID:   orderID,
		Reason:    reason,
	}

	// Publish to order topic as this is relevant to order processing
	return k.serializeAndPublish(ctx, payload, true)
}

// PublishStockReleased publishes an event that reserved stock has been released
func (k *KafkaEventPublisher) PublishStockReleased(ctx context.Context, reservation *entity.InventoryReservation) error {
	payload := StockEventPayload{
		EventType:     service.EventTypeStockReleased,
		Timestamp:     time.Now(),
		SKU:           reservation.SKU,
		OrderID:       reservation.OrderID,
		Quantity:      reservation.Qty,
		ReservationID: reservation.ReservationID,
		Data: map[string]interface{}{
			"reservation": reservation,
		},
	}

	// Publish to inventory topic
	if err := k.serializeAndPublish(ctx, payload, false); err != nil {
		return err
	}

	// Also publish order cancellation event
	orderPayload := StockEventPayload{
		EventType:     service.EventTypeOrderReservationCanceled,
		Timestamp:     time.Now(),
		SKU:           reservation.SKU,
		OrderID:       reservation.OrderID,
		Quantity:      reservation.Qty,
		ReservationID: reservation.ReservationID,
		Data: map[string]interface{}{
			"reservation": reservation,
		},
	}

	return k.serializeAndPublish(ctx, orderPayload, true)
}

// PublishStockDeducted publishes an event that stock has been deducted
func (k *KafkaEventPublisher) PublishStockDeducted(ctx context.Context, transaction *entity.StockTransaction) error {
	payload := StockEventPayload{
		EventType: service.EventTypeStockDeducted,
		Timestamp: time.Now(),
		SKU:       transaction.SKU,
		Quantity:  transaction.Qty,
		Data: map[string]interface{}{
			"transaction": transaction,
		},
	}

	// If there's a reference ID (usually order ID), include it
	if transaction.ReferenceID != nil {
		payload.OrderID = *transaction.ReferenceID
	}

	return k.serializeAndPublish(ctx, payload, false)
}

// PublishStockLow publishes an event that stock is below reorder level
func (k *KafkaEventPublisher) PublishStockLow(ctx context.Context, item *entity.InventoryItem) error {
	payload := StockEventPayload{
		EventType:    service.EventTypeStockLow,
		Timestamp:    time.Now(),
		SKU:          item.SKU,
		AvailableQty: item.AvailableQty,
		ReorderLevel: item.ReorderLevel,
		Data: map[string]interface{}{
			"inventory_item": item,
		},
	}

	return k.serializeAndPublish(ctx, payload, false)
}

// Close closes the Kafka writer connections
func (k *KafkaEventPublisher) Close() error {
	if err := k.writer.Close(); err != nil {
		return fmt.Errorf("failed to close inventory topic writer: %w", err)
	}
	if err := k.orderWriter.Close(); err != nil {
		return fmt.Errorf("failed to close order topic writer: %w", err)
	}
	k.serviceState = "closed"
	return nil
}
