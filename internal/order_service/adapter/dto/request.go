// internal/order_service/adapter/dto/request.go
package dto

import (
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// AddressRequest represents an address in a request
type AddressRequest struct {
	Street     string `json:"street" validate:"required"`
	City       string `json:"city" validate:"required"`
	State      string `json:"state" validate:"required"`
	Country    string `json:"country" validate:"required"`
	PostalCode string `json:"postal_code" validate:"required"`
}

// OrderItemRequest represents an order item in a request
type OrderItemRequest struct {
	ProductID   string  `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// PaymentRequest represents payment information in a request
type PaymentRequest struct {
	Method string  `json:"method" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// OrderRequest represents an order creation or update request
type OrderRequest struct {
	UserID       string             `json:"user_id" validate:"required"`
	Items        []OrderItemRequest `json:"items" validate:"required,dive"`
	ShippingInfo AddressRequest     `json:"shipping_info" validate:"required"`
	BillingInfo  AddressRequest     `json:"billing_info" validate:"required"`
	Payment      PaymentRequest     `json:"payment" validate:"required"`
	Notes        string             `json:"notes,omitempty"`
}

// ToEntity converts OrderRequest to an order entity
func (o OrderRequest) ToEntity() entity.Order {
	// Convert OrderItemRequests to OrderItems
	items := make([]entity.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = entity.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Price * float64(item.Quantity),
		}
	}

	// Create and return the Order entity
	order := entity.Order{
		UserID: o.UserID,
		Items:  items,
		ShippingInfo: entity.Address{
			Street:     o.ShippingInfo.Street,
			City:       o.ShippingInfo.City,
			State:      o.ShippingInfo.State,
			Country:    o.ShippingInfo.Country,
			PostalCode: o.ShippingInfo.PostalCode,
		},
		BillingInfo: entity.Address{
			Street:     o.BillingInfo.Street,
			City:       o.BillingInfo.City,
			State:      o.BillingInfo.State,
			Country:    o.BillingInfo.Country,
			PostalCode: o.BillingInfo.PostalCode,
		},
		Payment: entity.Payment{
			Method: o.Payment.Method,
			Amount: o.Payment.Amount,
		},
		Notes: o.Notes,
	}

	// Calculate total
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Subtotal
	}
	order.TotalAmount = totalAmount

	return order
}

// UpdateOrderStatusRequest represents a request to update an order's status
type UpdateOrderStatusRequest struct {
	Status  string `json:"status" validate:"required"`
	Comment string `json:"comment,omitempty"`
}

// CancelOrderRequest represents a request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason,omitempty"`
}
