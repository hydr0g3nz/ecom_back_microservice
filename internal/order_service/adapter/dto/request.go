package dto

import (
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id" validate:"required"`
	Items           []OrderItemRequest `json:"items" validate:"required,dive,required"`
	ShippingAddress string             `json:"shipping_address" validate:"required"`
}

// OrderItemRequest represents an item in an order request
type OrderItemRequest struct {
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	Price     float64 `json:"price" validate:"required,min=0"`
}

// Validate checks if the create order request is valid
func (r *CreateOrderRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user ID is required")
	}

	if len(r.Items) == 0 {
		return errors.New("at least one item is required")
	}

	for _, item := range r.Items {
		if item.ProductID == "" {
			return errors.New("product ID is required")
		}

		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}

		if item.Price < 0 {
			return errors.New("price cannot be negative")
		}
	}

	if r.ShippingAddress == "" {
		return errors.New("shipping address is required")
	}

	return nil
}

// ToOrderItems converts OrderItemRequest to entity.OrderItem
func (r *CreateOrderRequest) ToOrderItems() []entity.OrderItem {
	items := make([]entity.OrderItem, len(r.Items))
	
	for i, item := range r.Items {
		items[i] = entity.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			TotalPrice: item.Price * float64(item.Quantity),
		}
	}
	
	return items
}

// UpdateOrderStatusRequest represents the request to update an order's status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// Validate checks if the update order status request is valid
func (r *UpdateOrderStatusRequest) Validate() error {
	if r.Status == "" {
		return errors.New("status is required")
	}

	status := valueobject.OrderStatus(r.Status)
	if !status.IsValid() {
		return errors.New("invalid order status")
	}

	return nil
}

// OrderPaymentRequest represents the request to add payment information to an order
type OrderPaymentRequest struct {
	OrderID   string  `json:"order_id" validate:"required"`
	PaymentID string  `json:"payment_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,min=0"`
	Method    string  `json:"method" validate:"required"`
}

// Validate checks if the order payment request is valid
func (r *OrderPaymentRequest) Validate() error {
	if r.OrderID == "" {
		return errors.New("order ID is required")
	}

	if r.PaymentID == "" {
		return errors.New("payment ID is required")
	}

	if r.Amount < 0 {
		return errors.New("amount cannot be negative")
	}

	if r.Method == "" {
		return errors.New("payment method is required")
	}

	return nil
}

// OrderFilterRequest represents the request to filter orders
type OrderFilterRequest struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// Validate checks if the order filter request is valid
func (r *OrderFilterRequest) Validate() error {
	if r.Status != "" {
		status := valueobject.OrderStatus(r.Status)
		if !status.IsValid() {
			return errors.New("invalid order status")
		}
	}

	if r.Page < 0 {
		return errors.New("page cannot be negative")
	}

	if r.Limit < 0 {
		return errors.New("limit cannot be negative")
	}

	return nil
}
