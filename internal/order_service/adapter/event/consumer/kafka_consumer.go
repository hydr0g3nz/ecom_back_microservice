// internal/order_service/adapter/event/consumer/kafka_consumer.go
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// EventPayload defines the structure of the event payload
type EventPayload struct {
	EventID     string      `json:"event_id"`
	EventType   string      `json:"event_type"`
	OccurredAt  time.Time   `json:"occurred_at"`
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      string      `json:"status"`
	Data        interface{} `json:"data,omitempty"`
}

// InventoryReservedPayload defines the payload for inventory reserved event
type InventoryReservedPayload struct {
	OrderID     string  `json:"order_id"`
	Success     bool    `json:"success"`
	Message     string  `json:"message,omitempty"`
	ProductID   string  `json:"product_id,omitempty"`
	Quantity    int     `json:"quantity,omitempty"`
	TotalAmount float64 `json:"total_amount,omitempty"`
}

// PaymentProcessedPayload defines the payload for payment processed event
type PaymentProcessedPayload struct {
	OrderID       string  `json:"order_id"`
	TransactionID string  `json:"transaction_id"`
	Success       bool    `json:"success"`
	Message       string  `json:"message,omitempty"`
	Amount        float64 `json:"amount"`
}

// KafkaConsumer implements a Kafka consumer for order-related events
type KafkaConsumer struct {
	consumer     *kafka.Consumer
	orderUsecase usecase.OrderUsecase
	logger       logger.Logger
	topics       struct {
		inventoryEvents string
		paymentEvents   string
	}
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// NewKafkaConsumer creates a new KafkaConsumer
func NewKafkaConsumer(
	brokers string,
	groupID string,
	orderUsecase usecase.OrderUsecase,
	logger logger.Logger,
) (*KafkaConsumer, error) {
	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       brokers,
		"group.id":                groupID,
		"auto.offset.reset":       "earliest",
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 5000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	kc := &KafkaConsumer{
		consumer:     consumer,
		orderUsecase: orderUsecase,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}

	// Set default topics
	kc.topics.inventoryEvents = "inventory-events-result"
	kc.topics.paymentEvents = "payment-events-result"

	return kc, nil
}

// Start starts the Kafka consumer
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	// Subscribe to inventory events
	if err := kc.SubscribeToInventoryEvents(ctx); err != nil {
		return err
	}

	// Subscribe to payment events
	if err := kc.SubscribeToPaymentEvents(ctx); err != nil {
		return err
	}

	return nil
}

// Close closes the Kafka consumer
func (kc *KafkaConsumer) Close() error {
	// Signal all consumers to stop
	close(kc.stopChan)

	// Wait for all consumers to finish
	kc.wg.Wait()

	// Close the Kafka consumer
	return kc.consumer.Close()
}

// SubscribeToInventoryEvents subscribes to inventory-related events
func (kc *KafkaConsumer) SubscribeToInventoryEvents(ctx context.Context) error {
	// Subscribe to inventory events topic
	err := kc.consumer.Subscribe(kc.topics.inventoryEvents, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to inventory events topic: %w", err)
	}

	// Start goroutine to consume messages
	kc.wg.Add(1)
	go func() {
		defer kc.wg.Done()
		kc.consumeInventoryEvents(ctx)
	}()

	kc.logger.Info("Subscribed to inventory events", "topic", kc.topics.inventoryEvents)
	return nil
}

// SubscribeToPaymentEvents subscribes to payment-related events
func (kc *KafkaConsumer) SubscribeToPaymentEvents(ctx context.Context) error {
	// Subscribe to payment events topic
	err := kc.consumer.Subscribe(kc.topics.paymentEvents, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to payment events topic: %w", err)
	}

	// Start goroutine to consume messages
	kc.wg.Add(1)
	go func() {
		defer kc.wg.Done()
		kc.consumePaymentEvents(ctx)
	}()

	kc.logger.Info("Subscribed to payment events", "topic", kc.topics.paymentEvents)
	return nil
}

// consumeInventoryEvents consumes messages from the inventory events topic
func (kc *KafkaConsumer) consumeInventoryEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("Context cancelled, stopping inventory events consumer")
			return
		case <-kc.stopChan:
			kc.logger.Info("Stopping inventory events consumer")
			return
		default:
			// Poll for messages
			msg, err := kc.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Timeout or no message
				if !err.(kafka.Error).IsTimeout() {
					kc.logger.Error("Failed to read message", "error", err)
				}
				continue
			}

			// Process message
			kc.processInventoryEvent(ctx, msg)
		}
	}
}

// consumePaymentEvents consumes messages from the payment events topic
func (kc *KafkaConsumer) consumePaymentEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("Context cancelled, stopping payment events consumer")
			return
		case <-kc.stopChan:
			kc.logger.Info("Stopping payment events consumer")
			return
		default:
			// Poll for messages
			msg, err := kc.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Timeout or no message
				if !err.(kafka.Error).IsTimeout() {
					kc.logger.Error("Failed to read message", "error", err)
				}
				continue
			}

			// Process message
			kc.processPaymentEvent(ctx, msg)
		}
	}
}

// processInventoryEvent processes a message from the inventory events topic
func (kc *KafkaConsumer) processInventoryEvent(ctx context.Context, msg *kafka.Message) {
	// Parse message payload
	var payload EventPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		kc.logger.Error("Failed to unmarshal inventory event payload", "error", err)
		return
	}

	kc.logger.Info("Received inventory event",
		"event_id", payload.EventID,
		"event_type", payload.EventType,
		"order_id", payload.OrderID)

	// Process event based on type
	switch payload.EventType {
	case service.EventTypeInventoryReserved:
		// Parse inventory reserved data
		jsonData, err := json.Marshal(payload.Data)
		if err != nil {
			kc.logger.Error("Failed to marshal inventory data", "error", err)
			return
		}

		var inventoryData InventoryReservedPayload
		if err := json.Unmarshal(jsonData, &inventoryData); err != nil {
			kc.logger.Error("Failed to unmarshal inventory data", "error", err)
			return
		}

		// Update order status based on inventory reservation result
		_, err = kc.orderUsecase.ProcessInventoryReserved(
			ctx,
			payload.OrderID,
			inventoryData.Success,
			inventoryData.Message,
		)
		if err != nil {
			kc.logger.Error("Failed to process inventory reserved event", "error", err, "order_id", payload.OrderID)
			return
		}

		kc.logger.Info("Processed inventory reserved event", "order_id", payload.OrderID, "success", inventoryData.Success)

	default:
		kc.logger.Warn("Unknown inventory event type", "event_type", payload.EventType)
	}
}

// processPaymentEvent processes a message from the payment events topic
func (kc *KafkaConsumer) processPaymentEvent(ctx context.Context, msg *kafka.Message) {
	// Parse message payload
	var payload EventPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		kc.logger.Error("Failed to unmarshal payment event payload", "error", err)
		return
	}

	kc.logger.Info("Received payment event",
		"event_id", payload.EventID,
		"event_type", payload.EventType,
		"order_id", payload.OrderID)

	// Process event based on type
	switch payload.EventType {
	case service.EventTypePaymentProcessed:
		// Parse payment processed data
		jsonData, err := json.Marshal(payload.Data)
		if err != nil {
			kc.logger.Error("Failed to marshal payment data", "error", err)
			return
		}

		var paymentData PaymentProcessedPayload
		if err := json.Unmarshal(jsonData, &paymentData); err != nil {
			kc.logger.Error("Failed to unmarshal payment data", "error", err)
			return
		}

		// Update order status based on payment processing result
		_, err = kc.orderUsecase.ProcessPaymentCompleted(
			ctx,
			payload.OrderID,
			paymentData.TransactionID,
			paymentData.Success,
		)
		if err != nil {
			kc.logger.Error("Failed to process payment completed event", "error", err, "order_id", payload.OrderID)
			return
		}

		kc.logger.Info("Processed payment completed event", "order_id", payload.OrderID, "success", paymentData.Success)

	default:
		kc.logger.Warn("Unknown payment event type", "event_type", payload.EventType)
	}
}
