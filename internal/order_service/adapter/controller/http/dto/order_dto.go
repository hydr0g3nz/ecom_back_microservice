package dto

import (
	"errors"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// CreateOrderRequest represents the request for creating an order
type CreateOrderRequest struct {
	UserID      string                   `json:"user_id"`
	Items       []OrderItemRequest       `json:"items"`
	ShippingAddress ShippingAddressRequest `json:"shipping_address"`
	BillingAddress  *ShippingAddressRequest `json:"billing_address,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// OrderItemRequest represents an item in an order request
type OrderItemRequest struct {
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	Name       string  `json:"name"`
	SKU        string  `json:"sku,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ShippingAddressRequest represents a shipping address in a request
type ShippingAddressRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	AddressLine1  string `json:"address_line_1"`
	AddressLine2  string `json:"address_line_2,omitempty"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Email         string `json:"email,omitempty"`
}

// UpdateOrderRequest represents the request for updating an order
type UpdateOrderRequest struct {
	ShippingAddress *ShippingAddressRequest `json:"shipping_address,omitempty"`
	BillingAddress  *ShippingAddressRequest `json:"billing_address,omitempty"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// CancelOrderRequest represents the request for cancelling an order
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

// ProcessPaymentRequest represents the request for processing a payment
type ProcessPaymentRequest struct {
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	Method           string                 `json:"method"`
	PaymentMethodID  string                 `json:"payment_method_id,omitempty"`
	PaymentIntentID  string                 `json:"payment_intent_id,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RefundPaymentRequest represents the request for refunding a payment
type RefundPaymentRequest struct {
	Amount  float64 `json:"amount"`
	Reason  string  `json:"reason"`
}

// UpdateShippingRequest represents the request for updating shipping
type UpdateShippingRequest struct {
	Carrier         string    `json:"carrier,omitempty"`
	TrackingNumber  string    `json:"tracking_number,omitempty"`
	Status          string    `json:"status,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// OrderResponse represents the response for an order
type OrderResponse struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Status          string                 `json:"status"`
	Items           []OrderItemResponse    `json:"items"`
	Subtotal        float64                `json:"subtotal"`
	Tax             float64                `json:"tax"`
	ShippingCost    float64                `json:"shipping_cost"`
	Total           float64                `json:"total"`
	PaymentID       string                 `json:"payment_id,omitempty"`
	ShippingAddress ShippingAddressResponse `json:"shipping_address"`
	BillingAddress  *ShippingAddressResponse `json:"billing_address,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// OrderItemResponse represents an item in an order response
type OrderItemResponse struct {
	ProductID  string                 `json:"product_id"`
	Quantity   int                    `json:"quantity"`
	UnitPrice  float64                `json:"unit_price"`
	Total      float64                `json:"total"`
	Name       string                 `json:"name"`
	SKU        string                 `json:"sku,omitempty"`
	ImageURL   string                 `json:"image_url,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ShippingAddressResponse represents a shipping address in a response
type ShippingAddressResponse struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	AddressLine1  string `json:"address_line_1"`
	AddressLine2  string `json:"address_line_2,omitempty"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Email         string `json:"email,omitempty"`
}

// PaymentResponse represents the response for a payment
type PaymentResponse struct {
	ID              string                 `json:"id"`
	OrderID         string                 `json:"order_id"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Method          string                 `json:"method"`
	Status          string                 `json:"status"`
	Description     string                 `json:"description,omitempty"`
	RefundedAmount  float64                `json:"refunded_amount,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
}

// ShippingResponse represents the response for shipping information
type ShippingResponse struct {
	ID                string                 `json:"id"`
	OrderID           string                 `json:"order_id"`
	Carrier           string                 `json:"carrier,omitempty"`
	TrackingNumber    string                 `json:"tracking_number,omitempty"`
	Status            string                 `json:"status"`
	EstimatedDelivery *time.Time             `json:"estimated_delivery,omitempty"`
	ShippedAt         *time.Time             `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time             `json:"delivered_at,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// OrderEventResponse represents the response for an order event
type OrderEventResponse struct {
	ID        string                 `json:"id"`
	OrderID   string                 `json:"order_id"`
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// Validate validates the CreateOrderRequest
func (r *CreateOrderRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	
	if len(r.Items) == 0 {
		return errors.New("at least one item is required")
	}
	
	for i, item := range r.Items {
		if item.ProductID == "" {
			return errors.New("product_id is required for all items")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be positive for all items")
		}
		if item.UnitPrice < 0 {
			return errors.New("unit_price must be non-negative for all items")
		}
		if item.Name == "" {
			return errors.New("name is required for all items")
		}
	}
	
	if r.ShippingAddress.FirstName == "" || r.ShippingAddress.LastName == "" {
		return errors.New("first_name and last_name are required in shipping address")
	}
	
	if r.ShippingAddress.AddressLine1 == "" {
		return errors.New("address_line_1 is required in shipping address")
	}
	
	if r.ShippingAddress.City == "" {
		return errors.New("city is required in shipping address")
	}
	
	if r.ShippingAddress.PostalCode == "" {
		return errors.New("postal_code is required in shipping address")
	}
	
	if r.ShippingAddress.Country == "" {
		return errors.New("country is required in shipping address")
	}
	
	return nil
}

// ToEntity converts a CreateOrderRequest to an Order entity
func (r *CreateOrderRequest) ToEntity() *entity.Order {
	order := &entity.Order{
		UserID: r.UserID,
		Items:  make([]*entity.OrderItem, len(r.Items)),
		ShippingAddress: entity.Address{
			FirstName:    r.ShippingAddress.FirstName,
			LastName:     r.ShippingAddress.LastName,
			AddressLine1: r.ShippingAddress.AddressLine1,
			AddressLine2: r.ShippingAddress.AddressLine2,
			City:         r.ShippingAddress.City,
			State:        r.ShippingAddress.State,
			PostalCode:   r.ShippingAddress.PostalCode,
			Country:      r.ShippingAddress.Country,
			PhoneNumber:  r.ShippingAddress.PhoneNumber,
			Email:        r.ShippingAddress.Email,
		},
		Metadata: r.Metadata,
	}
	
	// Add billing address if provided
	if r.BillingAddress != nil {
		billingAddress := entity.Address{
			FirstName:    r.BillingAddress.FirstName,
			LastName:     r.BillingAddress.LastName,
			AddressLine1: r.BillingAddress.AddressLine1,
			AddressLine2: r.BillingAddress.AddressLine2,
			City:         r.BillingAddress.City,
			State:        r.BillingAddress.State,
			PostalCode:   r.BillingAddress.PostalCode,
			Country:      r.BillingAddress.Country,
			PhoneNumber:  r.BillingAddress.PhoneNumber,
			Email:        r.BillingAddress.Email,
		}
		order.BillingAddress = &billingAddress
	}
	
	// Add items
	for i, item := range r.Items {
		order.Items[i] = &entity.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Name:      item.Name,
			SKU:       item.SKU,
			ImageURL:  item.ImageURL,
			Metadata:  item.Metadata,
		}
	}
	
	return order
}

// ToMap converts an UpdateOrderRequest to a map for updates
func (r *UpdateOrderRequest) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	
	if r.ShippingAddress != nil {
		updates["shipping_address"] = entity.Address{
			FirstName:    r.ShippingAddress.FirstName,
			LastName:     r.ShippingAddress.LastName,
			AddressLine1: r.ShippingAddress.AddressLine1,
			AddressLine2: r.ShippingAddress.AddressLine2,
			City:         r.ShippingAddress.City,
			State:        r.ShippingAddress.State,
			PostalCode:   r.ShippingAddress.PostalCode,
			Country:      r.ShippingAddress.Country,
			PhoneNumber:  r.ShippingAddress.PhoneNumber,
			Email:        r.ShippingAddress.Email,
		}
	}
	
	if r.BillingAddress != nil {
		updates["billing_address"] = entity.Address{
			FirstName:    r.BillingAddress.FirstName,
			LastName:     r.BillingAddress.LastName,
			AddressLine1: r.BillingAddress.AddressLine1,
			AddressLine2: r.BillingAddress.AddressLine2,
			City:         r.BillingAddress.City,
			State:        r.BillingAddress.State,
			PostalCode:   r.BillingAddress.PostalCode,
			Country:      r.BillingAddress.Country,
			PhoneNumber:  r.BillingAddress.PhoneNumber,
			Email:        r.BillingAddress.Email,
		}
	}
	
	if r.Metadata != nil {
		updates["metadata"] = r.Metadata
	}
	
	return updates
}

// Validate validates the ProcessPaymentRequest
func (r *ProcessPaymentRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	
	if r.Currency == "" {
		return errors.New("currency is required")
	}
	
	if r.Method == "" {
		return errors.New("payment method is required")
	}
	
	// For certain payment methods, additional fields might be required
	if r.Method == "card" && r.PaymentMethodID == "" {
		return errors.New("payment_method_id is required for card payments")
	}
	
	return nil
}

// ToEntity converts a ProcessPaymentRequest to a Payment entity
func (r *ProcessPaymentRequest) ToEntity() *entity.Payment {
	return &entity.Payment{
		Amount:          r.Amount,
		Currency:        r.Currency,
		Method:          r.Method,
		PaymentMethodID: r.PaymentMethodID,
		PaymentIntentID: r.PaymentIntentID,
		Description:     r.Description,
		Type:            entity.PaymentTypeCharge,
		Status:          entity.PaymentStatusPending,
		Metadata:        r.Metadata,
	}
}

// Validate validates the RefundPaymentRequest
func (r *RefundPaymentRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	
	if r.Reason == "" {
		return errors.New("reason is required")
	}
	
	return nil
}

// Validate validates the UpdateShippingRequest
func (r *UpdateShippingRequest) Validate() error {
	if r.Status != "" {
		validStatuses := map[string]bool{
			entity.ShippingStatusPending:    true,
			entity.ShippingStatusProcessing: true,
			entity.ShippingStatusShipped:    true,
			entity.ShippingStatusDelivered:  true,
			entity.ShippingStatusReturned:   true,
		}
		
		if !validStatuses[r.Status] {
			return errors.New("invalid status")
		}
		
		// If status is shipped, carrier and tracking number are required
		if r.Status == entity.ShippingStatusShipped && (r.Carrier == "" || r.TrackingNumber == "") {
			return errors.New("carrier and tracking_number are required when status is shipped")
		}
	}
	
	return nil
}

// ToEntity converts an UpdateShippingRequest to a Shipping entity
func (r *UpdateShippingRequest) ToEntity() *entity.Shipping {
	shipping := &entity.Shipping{
		Metadata: r.Metadata,
	}
	
	if r.Carrier != "" {
		shipping.Carrier = r.Carrier
	}
	
	if r.TrackingNumber != "" {
		shipping.TrackingNumber = r.TrackingNumber
	}
	
	if r.Status != "" {
		shipping.Status = r.Status
	}
	
	if r.EstimatedDelivery != nil {
		shipping.EstimatedDelivery = r.EstimatedDelivery
	}
	
	return shipping
}

// OrderToResponse converts an Order entity to an OrderResponse
func OrderToResponse(order *entity.Order) OrderResponse {
	resp := OrderResponse{
		ID:           order.ID,
		UserID:       order.UserID,
		Status:       order.Status,
		Subtotal:     order.Subtotal,
		Tax:          order.Tax,
		ShippingCost: order.ShippingCost,
		Total:        order.Total,
		PaymentID:    order.PaymentID,
		ShippingAddress: ShippingAddressResponse{
			FirstName:    order.ShippingAddress.FirstName,
			LastName:     order.ShippingAddress.LastName,
			AddressLine1: order.ShippingAddress.AddressLine1,
			AddressLine2: order.ShippingAddress.AddressLine2,
			City:         order.ShippingAddress.City,
			State:        order.ShippingAddress.State,
			PostalCode:   order.ShippingAddress.PostalCode,
			Country:      order.ShippingAddress.Country,
			PhoneNumber:  order.ShippingAddress.PhoneNumber,
			Email:        order.ShippingAddress.Email,
		},
		Metadata:  order.Metadata,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
	
	// Add billing address if exists
	if order.BillingAddress != nil {
		billingAddress := ShippingAddressResponse{
			FirstName:    order.BillingAddress.FirstName,
			LastName:     order.BillingAddress.LastName,
			AddressLine1: order.BillingAddress.AddressLine1,
			AddressLine2: order.BillingAddress.AddressLine2,
			City:         order.BillingAddress.City,
			State:        order.BillingAddress.State,
			PostalCode:   order.BillingAddress.PostalCode,
			Country:      order.BillingAddress.Country,
			PhoneNumber:  order.BillingAddress.PhoneNumber,
			Email:        order.BillingAddress.Email,
		}
		resp.BillingAddress = &billingAddress
	}
	
	// Add items
	resp.Items = make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		resp.Items[i] = OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Total:     item.UnitPrice * float64(item.Quantity),
			Name:      item.Name,
			SKU:       item.SKU,
			ImageURL:  item.ImageURL,
			Metadata:  item.Metadata,
		}
	}
	
	return resp
}

// OrderEventToResponse converts an OrderEvent entity to an OrderEventResponse
func OrderEventToResponse(event *entity.OrderEvent) OrderEventResponse {
	return OrderEventResponse{
		ID:        event.ID,
		OrderID:   event.OrderID,
		Type:      event.Type,
		Data:      event.Data,
		CreatedAt: event.CreatedAt,
	}
}

// ShippingToResponse converts a Shipping entity to a ShippingResponse
func ShippingToResponse(shipping *entity.Shipping) ShippingResponse {
	return ShippingResponse{
		ID:                shipping.ID,
		OrderID:           shipping.OrderID,
		Carrier:           shipping.Carrier,
		TrackingNumber:    shipping.TrackingNumber,
		Status:            shipping.Status,
		EstimatedDelivery: shipping.EstimatedDelivery,
		ShippedAt:         shipping.ShippedAt,
		DeliveredAt:       shipping.DeliveredAt,
		Metadata:          shipping.Metadata,
		CreatedAt:         shipping.CreatedAt,
		UpdatedAt:         shipping.UpdatedAt,
	}
}