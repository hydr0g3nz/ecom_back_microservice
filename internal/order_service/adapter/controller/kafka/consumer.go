package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/segmentio/kafka-go"
)

// KafkaConsumer handles consuming events from Kafka
type KafkaConsumer struct {
	inventoryReader       *kafka.Reader
	paymentReader         *kafka.Reader
	cancelOrderUsecase    command.CancelOrderUsecase
	processPaymentUsecase command.ProcessPaymentUsecase
	updateShippingUsecase command.UpdateShippingUsecase
}

// NewKafkaConsumer creates a new instance of KafkaConsumer
func NewKafkaConsumer(
	brokers []string,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
) *KafkaConsumer {
	inventoryReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "inventory",
		GroupID:     "order-service-inventory",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     time.Second,
		StartOffset: kafka.FirstOffset,
	})

	paymentReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       "payments",
		GroupID:     "order-service-payments",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     time.Second,
		StartOffset: kafka.FirstOffset,
	})

	return &KafkaConsumer{
		inventoryReader:       inventoryReader,
		paymentReader:         paymentReader,
		cancelOrderUsecase:    cancelOrderUsecase,
		processPaymentUsecase: processPaymentUsecase,
		updateShippingUsecase: updateShippingUsecase,
	}
}

// Start starts consuming messages from Kafka topics
func (c *KafkaConsumer) Start(ctx context.Context) {
	// Start inventory consumer
	go func() {
		c.consumeInventoryEvents(ctx)
	}()

	// Start payment consumer
	go func() {
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
			return
		default:
			m, err := c.inventoryReader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from inventory topic: %v", err)
				continue
			}

			// Process the message
			err = c.handleInventoryEvent(ctx, m)
			if err != nil {
				log.Printf("Error handling inventory event: %v", err)
			}
		}
	}
}

// consumePaymentEvents consumes events from the payment topic
func (c *KafkaConsumer) consumePaymentEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := c.paymentReader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message from payment topic: %v", err)
				continue
			}

			// Process the message
			err = c.handlePaymentEvent(ctx, m)
			if err != nil {
				log.Printf("Error handling payment event: %v", err)
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

	switch eventType {
	case "stock_unavailable":
		return c.handleStockUnavailable(ctx, event)
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

	// Cancel the order
	return c.cancelOrderUsecase.Execute(ctx, orderID, reason)
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
	paymentInput := command.ProcessPaymentInput{
		OrderID:         orderID,
		Amount:          amount,
		Currency:        getStringOrDefault(event, "currency", "USD"),
		Method:          getStringOrDefault(event, "method", "credit_card"),
		TransactionID:   getStringOrDefault(event, "transaction_id", ""),
		GatewayResponse: getStringOrDefault(event, "gateway_response", ""),
	}

	// Process the payment
	_, err := c.processPaymentUsecase.Execute(ctx, paymentInput)
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

	// Cancel the order
	return c.cancelOrderUsecase.Execute(ctx, orderID, reason)
}

// getStringOrDefault gets a string value from a map or returns a default value
func getStringOrDefault(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}
