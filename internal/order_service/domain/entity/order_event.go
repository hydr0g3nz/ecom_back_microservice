package entity

import (
	"encoding/json"
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// EventType defines the type of order event
type EventType string

const (
	EventOrderCreated        EventType = "order_created"
	EventOrderUpdated        EventType = "order_updated"
	EventOrderCancelled      EventType = "order_cancelled"
	EventOrderCompleted      EventType = "order_completed"
	EventPaymentProcessed    EventType = "payment_processed"
	EventShippingUpdated     EventType = "shipping_updated"
	EventItemAdded           EventType = "item_added"
	EventItemRemoved         EventType = "item_removed"
	EventItemQuantityUpdated EventType = "item_quantity_updated"
	EventDiscountApplied     EventType = "discount_applied"
	EventStatusChanged       EventType = "status_changed"
)

// IsValidEventType checks if the event type is valid
func IsValidEventType(eventType EventType) bool {
	validTypes := []EventType{
		EventOrderCreated,
		EventOrderUpdated,
		EventOrderCancelled,
		EventOrderCompleted,
		EventPaymentProcessed,
		EventShippingUpdated,
		EventItemAdded,
		EventItemRemoved,
		EventItemQuantityUpdated,
		EventDiscountApplied,
		EventStatusChanged,
	}

	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

// OrderEvent represents an event in the order's lifecycle
type OrderEvent struct {
	ID        valueobject.ID          `json:"id"`
	OrderID   valueobject.ID          `json:"order_id"`
	Type      EventType               `json:"type"`
	Data      json.RawMessage         `json:"data"` // Serialized event data
	Version   int                     `json:"version"`
	Timestamp valueobject.Timestamp   `json:"timestamp"`
	UserID    valueobject.ID          `json:"user_id"`
}

// ValidateOrderEvent validates the order event
func ValidateOrderEvent(event OrderEvent) error {
	if event.OrderID.String() == "" {
		return errors.New("order ID is required")
	}
	if !IsValidEventType(event.Type) {
		return errors.New("invalid event type")
	}
	if len(event.Data) == 0 {
		return errors.New("event data is required")
	}
	if event.Version < 1 {
		return errors.New("event version must be positive")
	}
	if event.Timestamp.IsZero() {
		return errors.New("event timestamp is required")
	}
	return nil
}

// NewOrderEvent creates a new order event
func NewOrderEvent(
	id valueobject.ID,
	orderID valueobject.ID,
	eventType EventType,
	data json.RawMessage,
	version int,
	timestamp valueobject.Timestamp,
	userID valueobject.ID,
) (*OrderEvent, error) {
	event := &OrderEvent{
		ID:        id,
		OrderID:   orderID,
		Type:      eventType,
		Data:      data,
		Version:   version,
		Timestamp: timestamp,
		UserID:    userID,
	}

	// Validate the event
	if err := ValidateOrderEvent(*event); err != nil {
		return nil, err
	}

	return event, nil
}

// Generic event data structures

// StatusChangedData represents data for a status change event
type StatusChangedData struct {
	PreviousStatus valueobject.OrderStatus `json:"previous_status"`
	NewStatus      valueobject.OrderStatus `json:"new_status"`
	Reason         string                  `json:"reason"`
}

// ItemData represents data for item-related events
type ItemData struct {
	Item     OrderItem `json:"item"`
	Quantity int       `json:"quantity,omitempty"` // For quantity updates
}

// DiscountData represents data for discount-related events
type DiscountData struct {
	Discount Discount `json:"discount"`
}

// OrderCreatedData contains full order data for order creation event
type OrderCreatedData struct {
	Order Order `json:"order"`
}

// PaymentProcessedData contains payment data for payment events
type PaymentProcessedData struct {
	Payment Payment `json:"payment"`
}

// ShippingUpdatedData contains shipping data for shipping events
type ShippingUpdatedData struct {
	Shipping Shipping `json:"shipping"`
}

// DomainEvent represents a generic domain event
type DomainEvent interface {
	EventType() EventType
	AggregateID() valueobject.ID
	AggregateType() string
	Timestamp() valueobject.Timestamp
	Version() int
	Data() interface{}
}

// orderDomainEvent implements the DomainEvent interface
type orderDomainEvent struct {
	event     *OrderEvent
	eventData interface{}
}

// NewDomainEvent creates a new domain event from an OrderEvent
func NewDomainEvent(event *OrderEvent, eventData interface{}) DomainEvent {
	return &orderDomainEvent{
		event:     event,
		eventData: eventData,
	}
}

// EventType returns the event type
func (e *orderDomainEvent) EventType() EventType {
	return e.event.Type
}

// AggregateID returns the aggregate ID (order ID)
func (e *orderDomainEvent) AggregateID() valueobject.ID {
	return e.event.OrderID
}

// AggregateType returns the aggregate type
func (e *orderDomainEvent) AggregateType() string {
	return "Order"
}

// Timestamp returns the event timestamp
func (e *orderDomainEvent) Timestamp() valueobject.Timestamp {
	return e.event.Timestamp
}

// Version returns the event version
func (e *orderDomainEvent) Version() int {
	return e.event.Version
}

// Data returns the event data
func (e *orderDomainEvent) Data() interface{} {
	return e.eventData
}
