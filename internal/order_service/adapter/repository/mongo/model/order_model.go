// internal/order_service/adapter/repository/mongo/model/order_model.go
package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderItem represents a single item in an order in MongoDB
type OrderItem struct {
	ProductID   string  `bson:"product_id"`
	ProductName string  `bson:"product_name"`
	Quantity    int     `bson:"quantity"`
	Price       float64 `bson:"price"`
	Subtotal    float64 `bson:"subtotal"`
}

// Address represents a shipping or billing address in MongoDB
type Address struct {
	Street     string `bson:"street"`
	City       string `bson:"city"`
	State      string `bson:"state"`
	Country    string `bson:"country"`
	PostalCode string `bson:"postal_code"`
}

// Payment represents payment information for an order in MongoDB
type Payment struct {
	Method        string     `bson:"method"`
	Amount        float64    `bson:"amount"`
	TransactionID string     `bson:"transaction_id"`
	Status        string     `bson:"status"`
	PaidAt        *time.Time `bson:"paid_at,omitempty"`
}

// OrderStatusHistoryItem represents a status change in the order history in MongoDB
type OrderStatusHistoryItem struct {
	Status    string    `bson:"status"`
	Timestamp time.Time `bson:"timestamp"`
	Comment   string    `bson:"comment,omitempty"`
}

// OrderModel represents an order document in MongoDB
type OrderModel struct {
	ID            string                   `bson:"_id"`
	UserID        string                   `bson:"user_id"`
	Items         []OrderItem              `bson:"items"`
	TotalAmount   float64                  `bson:"total_amount"`
	Status        string                   `bson:"status"`
	ShippingInfo  Address                  `bson:"shipping_info"`
	BillingInfo   Address                  `bson:"billing_info"`
	Payment       Payment                  `bson:"payment"`
	Notes         string                   `bson:"notes,omitempty"`
	CreatedAt     time.Time                `bson:"created_at"`
	UpdatedAt     time.Time                `bson:"updated_at"`
	StatusHistory []OrderStatusHistoryItem `bson:"status_history"`
}

// ToEntity converts the MongoDB OrderModel to the domain entity Order
func (om *OrderModel) ToEntity() *entity.Order {
	// Convert OrderItems
	items := make([]entity.OrderItem, len(om.Items))
	for i, item := range om.Items {
		items[i] = entity.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Convert StatusHistory
	statusHistory := make([]entity.OrderStatusHistoryItem, len(om.StatusHistory))
	for i, history := range om.StatusHistory {
		orderStatus, _ := valueobject.ParseOrderStatus(history.Status)
		statusHistory[i] = entity.OrderStatusHistoryItem{
			Status:    orderStatus,
			Timestamp: history.Timestamp,
			Comment:   history.Comment,
		}
	}

	// Parse status
	status, _ := valueobject.ParseOrderStatus(om.Status)

	return &entity.Order{
		ID:          om.ID,
		UserID:      om.UserID,
		Items:       items,
		TotalAmount: om.TotalAmount,
		Status:      status,
		ShippingInfo: entity.Address{
			Street:     om.ShippingInfo.Street,
			City:       om.ShippingInfo.City,
			State:      om.ShippingInfo.State,
			Country:    om.ShippingInfo.Country,
			PostalCode: om.ShippingInfo.PostalCode,
		},
		BillingInfo: entity.Address{
			Street:     om.BillingInfo.Street,
			City:       om.BillingInfo.City,
			State:      om.BillingInfo.State,
			Country:    om.BillingInfo.Country,
			PostalCode: om.BillingInfo.PostalCode,
		},
		Payment: entity.Payment{
			Method:        om.Payment.Method,
			Amount:        om.Payment.Amount,
			TransactionID: om.Payment.TransactionID,
			Status:        om.Payment.Status,
			PaidAt:        om.Payment.PaidAt,
		},
		Notes:         om.Notes,
		CreatedAt:     om.CreatedAt,
		UpdatedAt:     om.UpdatedAt,
		StatusHistory: statusHistory,
	}
}

// FromEntity creates a new MongoDB OrderModel from a domain entity Order
func FromEntity(order *entity.Order) *OrderModel {
	// Convert OrderItems
	items := make([]OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Convert StatusHistory
	statusHistory := make([]OrderStatusHistoryItem, len(order.StatusHistory))
	for i, history := range order.StatusHistory {
		statusHistory[i] = OrderStatusHistoryItem{
			Status:    history.Status.String(),
			Timestamp: history.Timestamp,
			Comment:   history.Comment,
		}
	}

	return &OrderModel{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		ShippingInfo: Address{
			Street:     order.ShippingInfo.Street,
			City:       order.ShippingInfo.City,
			State:      order.ShippingInfo.State,
			Country:    order.ShippingInfo.Country,
			PostalCode: order.ShippingInfo.PostalCode,
		},
		BillingInfo: Address{
			Street:     order.BillingInfo.Street,
			City:       order.BillingInfo.City,
			State:      order.BillingInfo.State,
			Country:    order.BillingInfo.Country,
			PostalCode: order.BillingInfo.PostalCode,
		},
		Payment: Payment{
			Method:        order.Payment.Method,
			Amount:        order.Payment.Amount,
			TransactionID: order.Payment.TransactionID,
			Status:        order.Payment.Status,
			PaidAt:        order.Payment.PaidAt,
		},
		Notes:         order.Notes,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
		StatusHistory: statusHistory,
	}
}
