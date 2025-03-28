package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// OrderResponse represents the order response
type OrderResponse struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	Items           []OrderItemResponse  `json:"items"`
	TotalAmount     float64              `json:"total_amount"`
	Status          string               `json:"status"`
	ShippingAddress string               `json:"shipping_address"`
	PaymentID       string               `json:"payment_id,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// OrderItemResponse represents an item in an order response
type OrderItemResponse struct {
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	TotalPrice float64 `json:"total_price"`
}

// FromEntity converts a domain entity to a response DTO
func (r *OrderResponse) FromEntity(order *entity.Order) {
	r.ID = order.ID
	r.UserID = order.UserID
	r.TotalAmount = order.TotalAmount
	r.Status = string(order.Status)
	r.ShippingAddress = order.ShippingAddress
	r.PaymentID = order.PaymentID
	r.CreatedAt = order.CreatedAt
	r.UpdatedAt = order.UpdatedAt

	r.Items = make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		r.Items[i] = OrderItemResponse{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			TotalPrice: item.TotalPrice,
		}
	}
}

// NewOrderResponse creates a new OrderResponse from an entity
func NewOrderResponse(order *entity.Order) *OrderResponse {
	response := &OrderResponse{}
	response.FromEntity(order)
	return response
}

// OrdersResponse represents a list of orders response with pagination
type OrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// OrderSummaryResponse represents a summary of orders
type OrderSummaryResponse struct {
	TotalOrders     int     `json:"total_orders"`
	TotalAmount     float64 `json:"total_amount"`
	PendingOrders   int     `json:"pending_orders"`
	ProcessingOrders int    `json:"processing_orders"`
	ShippedOrders   int     `json:"shipped_orders"`
	DeliveredOrders int     `json:"delivered_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	OrderID     string  `json:"order_id"`
	PaymentID   string  `json:"payment_id"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	Method      string  `json:"method"`
	ProcessedAt time.Time `json:"processed_at"`
}

// FromEntities converts a list of domain entities to a paginated response
func NewOrdersResponse(orders []*entity.Order, totalCount, page, pageSize int) *OrdersResponse {
	response := &OrdersResponse{
		Orders:     make([]OrderResponse, len(orders)),
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (totalCount + pageSize - 1) / pageSize, // Ceiling division
	}

	for i, order := range orders {
		response.Orders[i] = *NewOrderResponse(order)
	}

	return response
}
