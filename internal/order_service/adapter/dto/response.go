// internal/order_service/adapter/dto/response.go
package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// AddressResponse represents an address in a response
type AddressResponse struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

// OrderItemResponse represents an order item in a response
type OrderItemResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// PaymentResponse represents payment information in a response
type PaymentResponse struct {
	Method        string     `json:"method"`
	Amount        float64    `json:"amount"`
	TransactionID string     `json:"transaction_id,omitempty"`
	Status        string     `json:"status,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}

// OrderStatusHistoryResponse represents a status change in the order history
type OrderStatusHistoryResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Comment   string    `json:"comment,omitempty"`
}

// OrderResponse represents an order response
type OrderResponse struct {
	ID            string                       `json:"id"`
	UserID        string                       `json:"user_id"`
	Items         []OrderItemResponse          `json:"items"`
	TotalAmount   float64                      `json:"total_amount"`
	Status        string                       `json:"status"`
	ShippingInfo  AddressResponse              `json:"shipping_info"`
	BillingInfo   AddressResponse              `json:"billing_info"`
	Payment       PaymentResponse              `json:"payment"`
	Notes         string                       `json:"notes,omitempty"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
	StatusHistory []OrderStatusHistoryResponse `json:"status_history"`
}

// OrderResponseFromEntity converts an order entity to OrderResponse
func OrderResponseFromEntity(order *entity.Order) OrderResponse {
	// Convert OrderItems
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Convert StatusHistory
	statusHistory := make([]OrderStatusHistoryResponse, len(order.StatusHistory))
	for i, history := range order.StatusHistory {
		statusHistory[i] = OrderStatusHistoryResponse{
			Status:    history.Status.String(),
			Timestamp: history.Timestamp,
			Comment:   history.Comment,
		}
	}

	// Create and return the OrderResponse
	return OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		ShippingInfo: AddressResponse{
			Street:     order.ShippingInfo.Street,
			City:       order.ShippingInfo.City,
			State:      order.ShippingInfo.State,
			Country:    order.ShippingInfo.Country,
			PostalCode: order.ShippingInfo.PostalCode,
		},
		BillingInfo: AddressResponse{
			Street:     order.BillingInfo.Street,
			City:       order.BillingInfo.City,
			State:      order.BillingInfo.State,
			Country:    order.BillingInfo.Country,
			PostalCode: order.BillingInfo.PostalCode,
		},
		Payment: PaymentResponse{
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

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(total, page, pageSize int, data interface{}) PaginatedResponse {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return PaginatedResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Data:       data,
	}
}
