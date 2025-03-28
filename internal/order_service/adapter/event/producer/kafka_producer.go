// internal/order_service/adapter/event/producer/kafka_producer.go
package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// EventPayload defines the structure of the event payload
type EventPayload struct {
	EventID     string          `json:"event_id"`
	EventType   string          `json:"event_type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	OrderID     string          `json:"order_id"`
	UserID      string          `json:"user_id"`
	TotalAmount float64         `json:"total_amount"`
	Status      string          `json:"status"`
	Items       []OrderItemData `json:"items,omitempty"`
	Data        interface{}     `json:"data,omitempty"`
}

// OrderItemData represents order item data in events
type OrderItemData struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// KafkaProducer implements EventService interface for producing events
type KafkaProducer struct {
	producer *kafka.Producer
	logger   logger.Logger
	topics   struct {
		orderEvents     string
		inventoryEvents string
		paymentEvents   string
	}
}

// NewKafkaProducer creates a new KafkaProducer
func NewKafkaProducer(brokers string, logger logger.Logger) (*KafkaProducer, error) {
	// Create Kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"acks":              "all",
		"retries":           3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	kp := &KafkaProducer{
		producer: producer,
		logger:   logger,
	}

	// Set default topics
	kp.topics.orderEvents = "order-events"
	kp.topics.inventoryEvents = "inventory-events"
	kp.topics.paymentEvents = "payment-events"

	// Start go routine to handle delivery reports
	go kp.handleDeliveryReports()

	return kp, nil
}

// handleDeliveryReports handles Kafka delivery reports
func (kp *KafkaProducer) handleDeliveryReports() {
	for e := range kp.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				kp.logger.Error("Failed to deliver message",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"error", ev.TopicPartition.Error)
			} else {
				kp.logger.Debug("Message delivered",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"offset", ev.TopicPartition.Offset)
			}
		}
	}
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() error {
	kp.producer.Flush(15 * 1000) // 15 seconds timeout
	kp.producer.Close()
	return nil
}

// produceEvent produces a Kafka event with the given payload to the specified topic
func (kp *KafkaProducer) produceEvent(topic, key string, payload interface{}) error {
	// Serialize payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: jsonPayload,
		Headers: []kafka.Header{
			{
				Key:   "content-type",
				Value: []byte("application/json"),
			},
		},
		Timestamp: time.Now(),
	}

	// Produce message to Kafka
	err = kp.producer.Produce(message, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// PublishOrderCreated publishes an event that a new order has been created
func (kp *KafkaProducer) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	// Create order items data
	items := make([]OrderItemData, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemData{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeOrderCreated,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		Items:       items,
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.orderEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish order created event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published order created event", "order_id", order.ID)
	return nil
}

// PublishOrderUpdated publishes an event that an order has been updated
func (kp *KafkaProducer) PublishOrderUpdated(ctx context.Context, order *entity.Order) error {
	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeOrderUpdated,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.orderEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish order updated event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published order updated event", "order_id", order.ID)
	return nil
}

// PublishOrderCancelled publishes an event that an order has been cancelled
func (kp *KafkaProducer) PublishOrderCancelled(ctx context.Context, order *entity.Order) error {
	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeOrderCancelled,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.orderEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish order cancelled event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published order cancelled event", "order_id", order.ID)
	return nil
}

// PublishOrderCompleted publishes an event that an order has been completed
func (kp *KafkaProducer) PublishOrderCompleted(ctx context.Context, order *entity.Order) error {
	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeOrderCompleted,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.orderEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish order completed event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published order completed event", "order_id", order.ID)
	return nil
}

// PublishReserveInventory publishes a request to reserve inventory for an order
func (kp *KafkaProducer) PublishReserveInventory(ctx context.Context, order *entity.Order) error {
	// Create order items data
	items := make([]OrderItemData, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemData{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeReserveInventory,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		Items:       items,
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.inventoryEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish reserve inventory event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published reserve inventory event", "order_id", order.ID)
	return nil
}

// PublishReleaseInventory publishes a request to release reserved inventory
func (kp *KafkaProducer) PublishReleaseInventory(ctx context.Context, order *entity.Order) error {
	// Create order items data
	items := make([]OrderItemData, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemData{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypeReleaseInventory,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		Items:       items,
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.inventoryEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish release inventory event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published release inventory event", "order_id", order.ID)
	return nil
}

// PublishPaymentRequest publishes a request to process payment for an order
func (kp *KafkaProducer) PublishPaymentRequest(ctx context.Context, order *entity.Order) error {
	// Create event payload
	payload := EventPayload{
		EventID:     fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType:   service.EventTypePaymentRequested,
		OccurredAt:  time.Now(),
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		Data: map[string]interface{}{
			"payment_method": order.Payment.Method,
			"amount":         order.Payment.Amount,
		},
	}

	// Produce event to Kafka
	err := kp.produceEvent(kp.topics.paymentEvents, order.ID, payload)
	if err != nil {
		kp.logger.Error("Failed to publish payment request event", "error", err, "order_id", order.ID)
		return err
	}

	kp.logger.Info("Published payment request event", "order_id", order.ID)
	return nil
}

// SubscribeToInventoryEvents subscribes to inventory-related events
// This is just a placeholder - the actual implementation will be in the consumer package
func (kp *KafkaProducer) SubscribeToInventoryEvents(ctx context.Context) error {
	kp.logger.Info("Subscribing to inventory events is handled by the Kafka consumer")
	return nil
}

// SubscribeToPaymentEvents subscribes to payment-related events
// This is just a placeholder - the actual implementation will be in the consumer package
func (kp *KafkaProducer) SubscribeToPaymentEvents(ctx context.Context) error {
	kp.logger.Info("Subscribing to payment events is handled by the Kafka consumer")
	return nil
}
