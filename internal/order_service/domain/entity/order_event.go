package entity

import (
	"encoding/json"
	"time"

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

// OrderEvent represents an event in the order's lifecycle
type OrderEvent struct {
	ID        string          `json:"id"`
	OrderID   string          `json:"order_id"`
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"` // Serialized event data
	Version   int             `json:"version"`
	Timestamp time.Time       `json:"timestamp"`
	UserID    string          `json:"user_id"`
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
