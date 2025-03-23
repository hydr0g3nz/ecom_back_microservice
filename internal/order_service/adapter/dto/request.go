package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
)

// OrderItemRequest represents an order item in the request
type OrderItemRequest struct {
	ProductID    string  `json:"product_id" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	SKU          string  `json:"sku" validate:"required"`
	Quantity     int     `json:"quantity" validate:"required,min=1"`
	Price        float64 `json:"price" validate:"required,min=0"`
	CurrencyCode string  `json:"currency_code" validate:"required"`
}

// AddressRequest represents an address in the request
type AddressRequest struct {
	FirstName    string `json:"first_name" validate:"required"`
	LastName     string `json:"last_name" validate:"required"`
	AddressLine1 string `json:"address_line1" validate:"required"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	PostalCode   string `json:"postal_code" validate:"required"`
	Country      string `json:"country" validate:"required"`
	Phone        string `json:"phone" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id" validate:"required"`
	Items           []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
	ShippingAddress AddressRequest     `json:"shipping_address" validate:"required,dive"`
	BillingAddress  AddressRequest     `json:"billing_address" validate:"required,dive"`
	Notes           string             `json:"notes"`
	PromotionCodes  []string           `json:"promotion_codes"`
}

// ToUsecaseInput converts CreateOrderRequest to usecase.CreateOrderInput
func (r *CreateOrderRequest) ToUsecaseInput() usecase.CreateOrderInput {
	items := make([]entity.OrderItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = entity.OrderItem{
			ProductID:    item.ProductID,
			Name:         item.Name,
			SKU:          item.SKU,
			Quantity:     item.Quantity,
			Price:        item.Price,
			CurrencyCode: item.CurrencyCode,
		}
	}

	return usecase.CreateOrderInput{
		UserID: r.UserID,
		Items:  items,
		ShippingAddress: entity.Address{
			FirstName:    r.ShippingAddress.FirstName,
			LastName:     r.ShippingAddress.LastName,
			AddressLine1: r.ShippingAddress.AddressLine1,
			AddressLine2: r.ShippingAddress.AddressLine2,
			City:         r.ShippingAddress.City,
			State:        r.ShippingAddress.State,
			PostalCode:   r.ShippingAddress.PostalCode,
			Country:      r.ShippingAddress.Country,
			Phone:        r.ShippingAddress.Phone,
			Email:        r.ShippingAddress.Email,
		},
		BillingAddress: entity.Address{
			FirstName:    r.BillingAddress.FirstName,
			LastName:     r.BillingAddress.LastName,
			AddressLine1: r.BillingAddress.AddressLine1,
			AddressLine2: r.BillingAddress.AddressLine2,
			City:         r.BillingAddress.City,
			State:        r.BillingAddress.State,
			PostalCode:   r.BillingAddress.PostalCode,
			Country:      r.BillingAddress.Country,
			Phone:        r.BillingAddress.Phone,
			Email:        r.BillingAddress.Email,
		},
		Notes:          r.Notes,
		PromotionCodes: r.PromotionCodes,
	}
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	Notes           *string         `json:"notes,omitempty"`
	ShippingAddress *AddressRequest `json:"shipping_address,omitempty"`
	BillingAddress  *AddressRequest `json:"billing_address,omitempty"`
}

// ToUsecaseInput converts UpdateOrderRequest to usecase.UpdateOrderInput
func (r *UpdateOrderRequest) ToUsecaseInput() usecase.UpdateOrderInput {
	var shippingAddress *entity.Address
	var billingAddress *entity.Address

	if r.ShippingAddress != nil {
		address := entity.Address{
			FirstName:    r.ShippingAddress.FirstName,
			LastName:     r.ShippingAddress.LastName,
			AddressLine1: r.ShippingAddress.AddressLine1,
			AddressLine2: r.ShippingAddress.AddressLine2,
			City:         r.ShippingAddress.City,
			State:        r.ShippingAddress.State,
			PostalCode:   r.ShippingAddress.PostalCode,
			Country:      r.ShippingAddress.Country,
			Phone:        r.ShippingAddress.Phone,
			Email:        r.ShippingAddress.Email,
		}
		shippingAddress = &address
	}

	if r.BillingAddress != nil {
		address := entity.Address{
			FirstName:    r.BillingAddress.FirstName,
			LastName:     r.BillingAddress.LastName,
			AddressLine1: r.BillingAddress.AddressLine1,
			AddressLine2: r.BillingAddress.AddressLine2,
			City:         r.BillingAddress.City,
			State:        r.BillingAddress.State,
			PostalCode:   r.BillingAddress.PostalCode,
			Country:      r.BillingAddress.Country,
			Phone:        r.BillingAddress.Phone,
			Email:        r.BillingAddress.Email,
		}
		billingAddress = &address
	}

	return usecase.UpdateOrderInput{
		Notes:           r.Notes,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
	}
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	Amount          float64 `json:"amount" validate:"required,min=0"`
	Currency        string  `json:"currency" validate:"required"`
	Method          string  `json:"method" validate:"required"`
	TransactionID   string  `json:"transaction_id"`
	GatewayResponse string  `json:"gateway_response"`
}

// ToUsecaseInput converts ProcessPaymentRequest to usecase.ProcessPaymentInput
func (r *ProcessPaymentRequest) ToUsecaseInput(orderID string) usecase.ProcessPaymentInput {
	return usecase.ProcessPaymentInput{
		OrderID:         orderID,
		Amount:          r.Amount,
		Currency:        r.Currency,
		Method:          r.Method,
		TransactionID:   r.TransactionID,
		GatewayResponse: r.GatewayResponse,
	}
}

// UpdateShippingRequest represents the request to update shipping information
type UpdateShippingRequest struct {
	Carrier           string     `json:"carrier" validate:"required"`
	TrackingNumber    string     `json:"tracking_number"`
	Status            string     `json:"status" validate:"required"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	ShippingMethod    string     `json:"shipping_method" validate:"required"`
	ShippingCost      float64    `json:"shipping_cost" validate:"min=0"`
	Notes             string     `json:"notes,omitempty"`
}

// ToUsecaseInput converts UpdateShippingRequest to usecase.UpdateShippingInput
func (r *UpdateShippingRequest) ToUsecaseInput(orderID string) (usecase.UpdateShippingInput, error) {
	status, err := valueobject.ParseShippingStatus(r.Status)
	if err != nil {
		return usecase.UpdateShippingInput{}, err
	}

	return usecase.UpdateShippingInput{
		OrderID:           orderID,
		Carrier:           r.Carrier,
		TrackingNumber:    r.TrackingNumber,
		Status:            status,
		EstimatedDelivery: r.EstimatedDelivery,
		ShippingMethod:    r.ShippingMethod,
		ShippingCost:      r.ShippingCost,
		Notes:             r.Notes,
	}, nil
}

// SearchOrdersRequest represents the request to search orders
type SearchOrdersRequest struct {
	UserID    string     `json:"user_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	ProductID string     `json:"product_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	MinAmount float64    `json:"min_amount,omitempty"`
	MaxAmount float64    `json:"max_amount,omitempty"`
	Page      int        `json:"page,omitempty"`
	PageSize  int        `json:"page_size,omitempty"`
}

// ToCriteria converts SearchOrdersRequest to a criteria map for searching
func (r *SearchOrdersRequest) ToCriteria() map[string]interface{} {
	criteria := make(map[string]interface{})

	if r.UserID != "" {
		criteria["user_id"] = r.UserID
	}

	if r.Status != "" {
		criteria["status"] = r.Status
	}

	if r.ProductID != "" {
		criteria["product_id"] = r.ProductID
	}

	if r.StartDate != nil {
		criteria["start_date"] = r.StartDate
	}

	if r.EndDate != nil {
		criteria["end_date"] = r.EndDate
	}

	if r.MinAmount > 0 {
		criteria["min_amount"] = r.MinAmount
	}

	if r.MaxAmount > 0 {
		criteria["max_amount"] = r.MaxAmount
	}

	return criteria
}
