package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// Topic names
const (
	TopicInventory = "inventory"
	TopicPayments  = "payments"
)

// KafkaConsumer handles consuming events from Kafka
type KafkaConsumer struct {
	inventoryReader *kafka.Reader
	paymentReader   *kafka.Reader
	orderUsecase    usecase.OrderUsecase
	paymentUsecase  usecase.PaymentUsecase
	logger          logger.Logger
}

// NewKafkaConsumer creates a new instance of KafkaConsumer
func NewKafkaConsumer(
	brokers []string,
	groupID string,
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	logger logger.Logger,
) *KafkaConsumer {
	inventoryReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicInventory,
		GroupID:     groupID + "-inventory",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     time.Second,
		StartOffset: kafka.FirstOffset,
	})

	paymentReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       TopicPayments,
		GroupID:     groupID + "-payments",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     time.Second,
		StartOffset: kafka.FirstOffset,
	})

	return &KafkaConsumer{
		inventoryReader: inventoryReader,
		paymentReader:   paymentReader,
		orderUsecase:    orderUsecase,
		paymentUsecase:  paymentUsecase,
		logger:          logger,
	}
}

// Start starts consuming messages from Kafka topics
func (c *KafkaConsumer) Start(ctx context.Context) {
	// Start inventory consumer
	go func() {
		c.logger.Info("Starting inventory events consumer")
		c.consumeInventoryEvents(ctx)
	}()

	// Start payment consumer
	go func() {
		c.logger.Info("Starting payment events consumer")
		c.consumePaymentEvents(ctx)
	}()
}

// Close closes all Kafka readers
func (c *KafkaConsumer) Close() error {
	if err := c.inventoryReader.Close(); err != nil {
		return err
	}

	if err := c.paymentReader.Close(); err != nil {
		return err
	}

	return nil
}

// consumeInventoryEvents consumes events from the inventory topic
func (c *KafkaConsumer) consumeInventoryEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutting down inventory events consumer")
			return
		default:
			m, err := c.inventoryReader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Error reading message from inventory topic", "error", err)
				continue
			}

			// Process the message
			err = c.handleInventoryEvent(ctx, m)
			if err != nil {
				c.logger.Error("Error handling inventory event", "error", err)
			}
		}
	}
}

// consumePaymentEvents consumes events from the payment topic
func (c *KafkaConsumer) consumePaymentEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutting down payment events consumer")
			return
		default:
			m, err := c.paymentReader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Error reading message from payment topic", "error", err)
				continue
			}

			// Process the message
			err = c.handlePaymentEvent(ctx, m)
			if err != nil {
				c.logger.Error("Error handling payment event", "error", err)
			}
		}
	}
}

// handleInventoryEvent handles an event from the inventory topic
func (c *KafkaConsumer) handleInventoryEvent(ctx context.Context, msg kafka.Message) error {
	var event map[string]interface{}
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return fmt.Errorf("error unmarshaling inventory event: %w", err)
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type in inventory event")
	}

	c.logger.Info("Received inventory event", "type", eventType)

	switch eventType {
	case "stock_unavailable":
		return c.handleStockUnavailable(ctx, event)
	case "stock_reserved":
		return c.handleStockReserved(ctx, event)
	// Add more event types as needed
	default:
		return fmt.Errorf("unknown inventory event type: %s", eventType)
	}
}

// handlePaymentEvent handles an event from the payment topic
func (c *KafkaConsumer) handlePaymentEvent(ctx context.Context, msg kafka.Message) error {
	var event map[string]interface{}
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return fmt.Errorf("error unmarshaling payment event: %w", err)
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type in payment event")
	}

	c.logger.Info("Received payment event", "type", eventType)

	switch eventType {
	case "payment_successful":
		return c.handlePaymentSuccessful(ctx, event)
	case "payment_failed":
		return c.handlePaymentFailed(ctx, event)
	// Add more event types as needed
	default:
		return fmt.Errorf("unknown payment event type: %s", eventType)
	}
}

// handleStockUnavailable handles a stock_unavailable event
func (c *KafkaConsumer) handleStockUnavailable(ctx context.Context, event map[string]interface{}) error {
	orderID, ok := event["order_id"].(string)
	if !ok {
		return fmt.Errorf("missing order_id in stock_unavailable event")
	}

	reason := "Order cancelled due to insufficient stock"
	if eventReason, ok := event["reason"].(string); ok {
		reason = eventReason
	}

	// Cancel the order using the order usecase
	return c.orderUsecase.CancelOrder(ctx, orderID, reason)
}

// handleStockReserved handles a stock_reserved event
func (c *KafkaConsumer) handleStockReserved(ctx context.Context, event map[string]interface{}) error {
	// This is just a placeholder for demonstration
	// In a real implementation, you might update the order status or trigger another action
	c.logger.Info("Stock reserved for order", "event", event)
	return nil
}

// handlePaymentSuccessful handles a payment_successful event
func (c *KafkaConsumer) handlePaymentSuccessful(ctx context.Context, event map[string]interface{}) error {
	// Extract payment details from the event
	orderID, ok := event["order_id"].(string)
	if !ok {
		return fmt.Errorf("missing order_id in payment_successful event")
	}

	amount, ok := event["amount"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid amount in payment_successful event")
	}

	// Build payment input from event data
	paymentInput := usecase.ProcessPaymentInput{
		OrderID:         orderID,
		Amount:          amount,
		Currency:        getStringOrDefault(event, "currency", "USD"),
		Method:          getStringOrDefault(event, "method", "credit_card"),
		TransactionID:   getStringOrDefault(event, "transaction_id", ""),
		GatewayResponse: getStringOrDefault(event, "gateway_response", ""),
	}

	// Process the payment using the payment usecase
	_, err := c.paymentUsecase.ProcessPayment(ctx, paymentInput)
	return err
}

// handlePaymentFailed handles a payment_failed event
func (c *KafkaConsumer) handlePaymentFailed(ctx context.Context, event map[string]interface{}) error {
	orderID, ok := event["order_id"].(string)
	if !ok {
		return fmt.Errorf("missing order_id in payment_failed event")
	}

	reason := "Order cancelled due to payment failure"
	if eventReason, ok := event["reason"].(string); ok {
		reason = eventReason
	}

	// Cancel the order using the order usecase
	return c.orderUsecase.CancelOrder(ctx, orderID, reason)
}

// getStringOrDefault gets a string value from a map or returns a default value
func getStringOrDefault(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}
