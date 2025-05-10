// internal/order_service/domain/entity/order.go
package entity

import (
	"fmt"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderItem represents a single item in an order
type OrderItem struct {
	ProductID   string  `json:"product_id" bson:"product_id"`
	ProductName string  `json:"product_name" bson:"product_name"`
	Quantity    int     `json:"quantity" bson:"quantity"`
	Price       float64 `json:"price" bson:"price"`
	Subtotal    float64 `json:"subtotal" bson:"subtotal"`
}

// Address represents a shipping or billing address
type Address struct {
	Street     string `json:"street" bson:"street"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	Country    string `json:"country" bson:"country"`
	PostalCode string `json:"postal_code" bson:"postal_code"`
}

// Payment represents payment information for an order
type Payment struct {
	Method        string     `json:"method" bson:"method"`
	Amount        float64    `json:"amount" bson:"amount"`
	TransactionID string     `json:"transaction_id" bson:"transaction_id"`
	Status        string     `json:"status" bson:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
}

// Order represents an order entity
type Order struct {
	ID            string                   `json:"id" bson:"_id"`
	UserID        string                   `json:"user_id" bson:"user_id"`
	Items         []OrderItem              `json:"items" bson:"items"`
	TotalAmount   float64                  `json:"total_amount" bson:"total_amount"`
	Status        valueobject.OrderStatus  `json:"status" bson:"status"`
	ShippingInfo  Address                  `json:"shipping_info" bson:"shipping_info"`
	BillingInfo   Address                  `json:"billing_info" bson:"billing_info"`
	Payment       Payment                  `json:"payment" bson:"payment"`
	Notes         string                   `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt     time.Time                `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at" bson:"updated_at"`
	StatusHistory []OrderStatusHistoryItem `json:"status_history" bson:"status_history"`
}

// OrderStatusHistoryItem represents a status change in the order history
type OrderStatusHistoryItem struct {
	Status    valueobject.OrderStatus `json:"status" bson:"status"`
	Timestamp time.Time               `json:"timestamp" bson:"timestamp"`
	Comment   string                  `json:"comment,omitempty" bson:"comment,omitempty"`
}

// AddStatusHistoryItem adds a new status change to the order history
func (o *Order) AddStatusHistoryItem(status valueobject.OrderStatus, comment string) {
	historyItem := OrderStatusHistoryItem{
		Status:    status,
		Timestamp: time.Now(),
		Comment:   comment,
	}
	o.StatusHistory = append(o.StatusHistory, historyItem)
	o.Status = status
	o.UpdatedAt = time.Now()
}

// CalculateTotalAmount calculates the total amount for the order
func (o *Order) CalculateTotalAmount() {
	var total float64
	for _, item := range o.Items {
		total += item.Subtotal
	}
	o.TotalAmount = total
}

// ValidateOrder validates if the order has all required fields
func (o *Order) ValidateOrder() error {
	if o.UserID == "" {
		return ErrInvalidOrderData
	}
	if len(o.Items) == 0 {
		return ErrInvalidOrderData
	}
	for _, item := range o.Items {
		if item.ProductID == "" || item.Quantity <= 0 {
			return ErrInvalidOrderData
		}
	}
	return nil
}

// CanTransitionToStatus checks if the order can transition to the given status
func (o *Order) CanTransitionToStatus(newStatus valueobject.OrderStatus) bool {
	// Define valid status transitions based on current status
	fmt.Println("Current status:", o.Status)
	fmt.Println("New status:", newStatus)
	switch o.Status {
	case valueobject.OrderStatusPending:
		return newStatus == valueobject.OrderStatusProcessing ||
			newStatus == valueobject.OrderStatusCancelled

	case valueobject.OrderStatusProcessing:
		return newStatus == valueobject.OrderStatusShipped ||
			newStatus == valueobject.OrderStatusCancelled ||
			newStatus == valueobject.OrderStatusFailed

	case valueobject.OrderStatusShipped:
		return newStatus == valueobject.OrderStatusDelivered ||
			newStatus == valueobject.OrderStatusReturned

	case valueobject.OrderStatusDelivered:
		return newStatus == valueobject.OrderStatusReturned ||
			newStatus == valueobject.OrderStatusCompleted

	case valueobject.OrderStatusFailed:
		return newStatus == valueobject.OrderStatusProcessing

	case valueobject.OrderStatusCancelled,
		valueobject.OrderStatusReturned,
		valueobject.OrderStatusCompleted:
		return false // Terminal states
	}

	return false
}
