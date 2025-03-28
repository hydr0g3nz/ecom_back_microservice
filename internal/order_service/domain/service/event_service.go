package service

import (
	"context"
	"encoding/json"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// EventType defines the type of event
type EventType string

const (
	// EventOrderCreated is fired when an order is created
	EventOrderCreated EventType = "ORDER_CREATED"
	
	// EventOrderUpdated is fired when an order is updated
	EventOrderUpdated EventType = "ORDER_UPDATED"
	
	// EventOrderStatusChanged is fired when an order status changes
	EventOrderStatusChanged EventType = "ORDER_STATUS_CHANGED"
	
	// EventOrderPaymentProcessed is fired when a payment is processed
	EventOrderPaymentProcessed EventType = "ORDER_PAYMENT_PROCESSED"
	
	// EventOrderShipped is fired when an order is shipped
	EventOrderShipped EventType = "ORDER_SHIPPED"
	
	// EventOrderDelivered is fired when an order is delivered
	EventOrderDelivered EventType = "ORDER_DELIVERED"
	
	// EventOrderCancelled is fired when an order is cancelled
	EventOrderCancelled EventType = "ORDER_CANCELLED"
)

// Event represents an event in the system
type Event struct {
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
	Version   string          `json:"version"`
}

// OrderEvent represents an event related to an order
type OrderEvent struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
}

// StatusChangeEvent represents a status change event
type StatusChangeEvent struct {
	OrderEvent
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

// PaymentProcessedEvent represents a payment processed event
type PaymentProcessedEvent struct {
	OrderEvent
	PaymentID   string  `json:"payment_id"`
	Amount      float64 `json:"amount"`
	PaymentType string  `json:"payment_type"`
}

// EventProducer defines the interface for producing events
type EventProducer interface {
	// Publish publishes an event to the message broker
	Publish(ctx context.Context, topic string, event Event) error
}

// EventConsumer defines the interface for consuming events
type EventConsumer interface {
	// Subscribe subscribes to events from a topic
	Subscribe(ctx context.Context, topic string, handler func(event Event) error) error
}

// EventService manages the event production and consumption
type EventService struct {
	producer EventProducer
	consumer EventConsumer
}

// NewEventService creates a new event service
func NewEventService(producer EventProducer, consumer EventConsumer) *EventService {
	return &EventService{
		producer: producer,
		consumer: consumer,
	}
}

// PublishOrderCreated publishes an order created event
func (s *EventService) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	
	event := Event{
		Type:      EventOrderCreated,
		Data:      data,
		Timestamp: order.CreatedAt.Unix(),
		Version:   "1.0",
	}
	
	return s.producer.Publish(ctx, "orders", event)
}

// PublishOrderStatusChanged publishes an order status changed event
func (s *EventService) PublishOrderStatusChanged(ctx context.Context, order *entity.Order, oldStatus string) error {
	statusEvent := StatusChangeEvent{
		OrderEvent: OrderEvent{
			OrderID: order.ID,
			UserID:  order.UserID,
		},
		OldStatus: oldStatus,
		NewStatus: string(order.Status),
	}
	
	data, err := json.Marshal(statusEvent)
	if err != nil {
		return err
	}
	
	event := Event{
		Type:      EventOrderStatusChanged,
		Data:      data,
		Timestamp: order.UpdatedAt.Unix(),
		Version:   "1.0",
	}
	
	return s.producer.Publish(ctx, "orders.status", event)
}
