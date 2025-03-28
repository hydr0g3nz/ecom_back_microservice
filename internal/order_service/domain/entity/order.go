package entity

import (
	"time"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// Order represents the order entity in the domain
type Order struct {
	ID          string                    `json:"id"`
	UserID      string                    `json:"user_id"`
	Items       []OrderItem               `json:"items"`
	TotalAmount float64                   `json:"total_amount"`
	Status      valueobject.OrderStatus   `json:"status"`
	ShippingAddress string                `json:"shipping_address"`
	PaymentID   string                    `json:"payment_id"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	TotalPrice float64 `json:"total_price"`
}

// NewOrder creates a new order
func NewOrder(userID string, items []OrderItem, shippingAddress string) *Order {
	totalAmount := calculateTotalAmount(items)
	
	return &Order{
		UserID:         userID,
		Items:          items,
		TotalAmount:    totalAmount,
		Status:         valueobject.OrderStatusCreated,
		ShippingAddress: shippingAddress,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// CalculateTotalAmount calculates the total amount of the order
func calculateTotalAmount(items []OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += item.TotalPrice
	}
	return total
}

// UpdateStatus updates the status of the order
func (o *Order) UpdateStatus(status valueobject.OrderStatus) error {
	if !status.IsValid() {
		return ErrInvalidOrderStatus
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// AddPaymentID adds a payment ID to the order
func (o *Order) AddPaymentID(paymentID string) {
	o.PaymentID = paymentID
	o.UpdatedAt = time.Now()
}
