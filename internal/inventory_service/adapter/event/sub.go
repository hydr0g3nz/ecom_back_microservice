package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/usecase"
	"github.com/segmentio/kafka-go"
)

// KafkaConfig holds the configuration for Kafka connection
type KafkaConfig struct {
	Brokers         []string
	InventoryTopic  string
	OrderTopic      string
	ConsumerGroupID string
}

// KafkaEventPublisher implements the EventPublisherService interface using Kafka
type KafkaEventPublisher struct {
	writer       *kafka.Writer
	orderWriter  *kafka.Writer
	kafkaConfig  *KafkaConfig
	serviceState string // Can be used for health checks
}

// KafkaEventSubscriber implements the EventSubscriberService interface using Kafka
type KafkaEventSubscriber struct {
	orderReader         *kafka.Reader
	reservationReader   *kafka.Reader
	inventoryUsecase    usecase.ReservationProcessorUsecase
	kafkaConfig         *KafkaConfig
	orderMessageHandler func(ctx context.Context, msg []byte) error
	serviceState        string // Can be used for health checks
}

// OrderEventPayload represents the order event structure
type OrderEventPayload struct {
	EventType string          `json:"event_type"`
	OrderID   string          `json:"order_id"`
	Items     []OrderItemData `json:"items"`
	Timestamp time.Time       `json:"timestamp"`
}

// OrderItemData represents the order item data
type OrderItemData struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// ReservationEventPayload represents the event payload for reservation/release operations
type ReservationEventPayload struct {
	EventType     string    `json:"event_type"`
	OrderID       string    `json:"order_id"`
	ReservationID string    `json:"reservation_id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewKafkaEventSubscriber creates a new Kafka event subscriber
func NewKafkaEventSubscriber(
	config *KafkaConfig,
	inventoryUsecase usecase.ReservationProcessorUsecase,
) (*KafkaEventSubscriber, error) {

	// Reader for order events
	orderReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       config.OrderTopic,
		GroupID:     config.ConsumerGroupID + "-orders",
		MinBytes:    10e3,             // 10KB
		MaxBytes:    10e6,             // 10MB
		StartOffset: kafka.LastOffset, // Start from the newest message
	})

	// Reader for inventory reservation/release events (if needed as a separate topic)
	reservationReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       "inventory_reservations", // Adjust as needed
		GroupID:     config.ConsumerGroupID + "-reservations",
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	return &KafkaEventSubscriber{
		orderReader:       orderReader,
		reservationReader: reservationReader,
		inventoryUsecase:  inventoryUsecase,
		kafkaConfig:       config,
		serviceState:      "ready",
	}, nil
}

// SubscribeToOrderEvents subscribes to order-related events
func (k *KafkaEventSubscriber) SubscribeToOrderEvents(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping order event subscription")
				return
			default:
				// Read messages from order topic
				msg, err := k.orderReader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message from order topic: %v", err)
					continue
				}

				// Process the message
				if err := k.processOrderMessage(ctx, msg.Value); err != nil {
					log.Printf("Error processing order message: %v", err)
				}
			}
		}
	}()

	return nil
}

// processOrderMessage processes messages from the order topic
func (k *KafkaEventSubscriber) processOrderMessage(ctx context.Context, msg []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(msg, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Extract event type
	eventType, ok := payload["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid event_type in message")
	}

	// Route to appropriate handler based on event type
	switch eventType {
	case "order.created":
		return k.HandleOrderCreated(ctx, msg)
	case "order.cancelled":
		return k.HandleOrderCancelled(ctx, msg)
	case "order.reservation.request":
		return k.HandleReservationRequest(ctx, msg)
	case "order.release.request":
		return k.HandleReleaseRequest(ctx, msg)
	default:
		log.Printf("Ignoring unknown event type: %s", eventType)
		return nil
	}
}

// HandleOrderCreated handles the event when an order is created
func (k *KafkaEventSubscriber) HandleOrderCreated(ctx context.Context, orderData []byte) error {
	return k.inventoryUsecase.ProcessReservationRequest(ctx, orderData)
}

// HandleOrderCancelled handles the event when an order is cancelled
func (k *KafkaEventSubscriber) HandleOrderCancelled(ctx context.Context, orderData []byte) error {
	return k.inventoryUsecase.ProcessReleaseRequest(ctx, orderData)
}

// HandleReservationRequest handles reservation requests
func (k *KafkaEventSubscriber) HandleReservationRequest(ctx context.Context, reservationData []byte) error {
	return k.inventoryUsecase.ProcessReservationRequest(ctx, reservationData)
}

// HandleReleaseRequest handles release requests
func (k *KafkaEventSubscriber) HandleReleaseRequest(ctx context.Context, releaseData []byte) error {
	return k.inventoryUsecase.ProcessReleaseRequest(ctx, releaseData)
}

// Close closes the Kafka reader connections
func (k *KafkaEventSubscriber) Close() error {
	if err := k.orderReader.Close(); err != nil {
		return fmt.Errorf("failed to close order reader: %w", err)
	}
	if err := k.reservationReader.Close(); err != nil {
		return fmt.Errorf("failed to close reservation reader: %w", err)
	}
	k.serviceState = "closed"
	return nil
}
