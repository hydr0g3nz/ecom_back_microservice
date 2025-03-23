package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	TotalPrice   float64 `json:"total_price"`
	CurrencyCode string  `json:"currency_code"`
}

// AddressResponse represents an address in the response
type AddressResponse struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
}

// DiscountResponse represents a discount in the response
type DiscountResponse struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
}

// OrderResponse represents an order in the response
type OrderResponse struct {
	ID              string              `json:"id"`
	UserID          string              `json:"user_id"`
	Items           []OrderItemResponse `json:"items"`
	TotalAmount     float64             `json:"total_amount"`
	Status          string              `json:"status"`
	ShippingAddress AddressResponse     `json:"shipping_address"`
	BillingAddress  AddressResponse     `json:"billing_address"`
	PaymentID       string              `json:"payment_id"`
	ShippingID      string              `json:"shipping_id"`
	Notes           string              `json:"notes"`
	PromotionCodes  []string            `json:"promotion_codes"`
	Discounts       []DiscountResponse  `json:"discounts"`
	TaxAmount       float64             `json:"tax_amount"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	CompletedAt     *time.Time          `json:"completed_at,omitempty"`
	CancelledAt     *time.Time          `json:"cancelled_at,omitempty"`
	Version         int                 `json:"version"`
}

// FromEntity converts an entity.Order to an OrderResponse
func OrderResponseFromEntity(order *entity.Order) OrderResponse {
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			Name:         item.Name,
			SKU:          item.SKU,
			Quantity:     item.Quantity,
			Price:        item.Price,
			TotalPrice:   item.TotalPrice,
			CurrencyCode: item.CurrencyCode,
		}
	}

	discounts := make([]DiscountResponse, len(order.Discounts))
	for i, discount := range order.Discounts {
		discounts[i] = DiscountResponse{
			Code:        discount.Code,
			Description: discount.Description,
			Type:        discount.Type,
			Amount:      discount.Amount,
		}
	}

	return OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		ShippingAddress: AddressResponse{
			FirstName:    order.ShippingAddress.FirstName,
			LastName:     order.ShippingAddress.LastName,
			AddressLine1: order.ShippingAddress.AddressLine1,
			AddressLine2: order.ShippingAddress.AddressLine2,
			City:         order.ShippingAddress.City,
			State:        order.ShippingAddress.State,
			PostalCode:   order.ShippingAddress.PostalCode,
			Country:      order.ShippingAddress.Country,
			Phone:        order.ShippingAddress.Phone,
			Email:        order.ShippingAddress.Email,
		},
		BillingAddress: AddressResponse{
			FirstName:    order.BillingAddress.FirstName,
			LastName:     order.BillingAddress.LastName,
			AddressLine1: order.BillingAddress.AddressLine1,
			AddressLine2: order.BillingAddress.AddressLine2,
			City:         order.BillingAddress.City,
			State:        order.BillingAddress.State,
			PostalCode:   order.BillingAddress.PostalCode,
			Country:      order.BillingAddress.Country,
			Phone:        order.BillingAddress.Phone,
			Email:        order.BillingAddress.Email,
		},
		PaymentID:      order.PaymentID,
		ShippingID:     order.ShippingID,
		Notes:          order.Notes,
		PromotionCodes: order.PromotionCodes,
		Discounts:      discounts,
		TaxAmount:      order.TaxAmount,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		CompletedAt:    order.CompletedAt,
		CancelledAt:    order.CancelledAt,
		Version:        order.Version,
	}
}

// PaymentResponse represents a payment in the response
type PaymentResponse struct {
	ID              string     `json:"id"`
	OrderID         string     `json:"order_id"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	Method          string     `json:"method"`
	Status          string     `json:"status"`
	TransactionID   string     `json:"transaction_id"`
	GatewayResponse string     `json:"gateway_response"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	FailedAt        *time.Time `json:"failed_at,omitempty"`
}

// FromEntity converts an entity.Payment to a PaymentResponse
func PaymentResponseFromEntity(payment *entity.Payment) PaymentResponse {
	return PaymentResponse{
		ID:              payment.ID,
		OrderID:         payment.OrderID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Method:          payment.Method,
		Status:          string(payment.Status),
		TransactionID:   payment.TransactionID,
		GatewayResponse: payment.GatewayResponse,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
		CompletedAt:     payment.CompletedAt,
		FailedAt:        payment.FailedAt,
	}
}

// ShippingResponse represents shipping information in the response
type ShippingResponse struct {
	ID                string     `json:"id"`
	OrderID           string     `json:"order_id"`
	Carrier           string     `json:"carrier"`
	TrackingNumber    string     `json:"tracking_number"`
	Status            string     `json:"status"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	ShippedAt         *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	ShippingMethod    string     `json:"shipping_method"`
	ShippingCost      float64    `json:"shipping_cost"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// FromEntity converts an entity.Shipping to a ShippingResponse
func ShippingResponseFromEntity(shipping *entity.Shipping) ShippingResponse {
	return ShippingResponse{
		ID:                shipping.ID,
		OrderID:           shipping.OrderID,
		Carrier:           shipping.Carrier,
		TrackingNumber:    shipping.TrackingNumber,
		Status:            string(shipping.Status),
		EstimatedDelivery: shipping.EstimatedDelivery,
		ShippedAt:         shipping.ShippedAt,
		DeliveredAt:       shipping.DeliveredAt,
		ShippingMethod:    shipping.ShippingMethod,
		ShippingCost:      shipping.ShippingCost,
		Notes:             shipping.Notes,
		CreatedAt:         shipping.CreatedAt,
		UpdatedAt:         shipping.UpdatedAt,
	}
}

// OrderEventResponse represents an event in the order's lifecycle
type OrderEventResponse struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Type      string    `json:"type"`
	Data      []byte    `json:"data"`
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
}

// FromEntity converts an entity.OrderEvent to an OrderEventResponse
func OrderEventResponseFromEntity(event *entity.OrderEvent) OrderEventResponse {
	return OrderEventResponse{
		ID:        event.ID,
		OrderID:   event.OrderID,
		Type:      string(event.Type),
		Data:      event.Data,
		Version:   event.Version,
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
	}
}

// PaginatedOrdersResponse represents a paginated list of orders
type PaginatedOrdersResponse struct {
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
	Orders     []OrderResponse `json:"orders"`
}

// CreatePaginatedResponse creates a paginated response from a list of orders
func CreatePaginatedResponse(orders []*entity.Order, total, page, pageSize int) PaginatedOrdersResponse {
	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = OrderResponseFromEntity(order)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return PaginatedOrdersResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Orders:     orderResponses,
	}
}

// OrderHistoryResponse represents the event history for an order
type OrderHistoryResponse struct {
	OrderID string               `json:"order_id"`
	Events  []OrderEventResponse `json:"events"`
}
