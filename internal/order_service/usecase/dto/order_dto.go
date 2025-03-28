package dto

// OrderItemDTO represents an order item data transfer object
type OrderItemDTO struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	TotalPrice   float64 `json:"total_price"`
	CurrencyCode string  `json:"currency_code"`
}

// AddressDTO represents an address data transfer object
type AddressDTO struct {
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

// DiscountDTO represents a discount data transfer object
type DiscountDTO struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // percentage, fixed, shipping
	Amount      float64 `json:"amount"`
}

// OrderDTO represents an order data transfer object
type OrderDTO struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	Items           []OrderItemDTO `json:"items"`
	TotalAmount     float64       `json:"total_amount"`
	Status          string        `json:"status"`
	ShippingAddress AddressDTO    `json:"shipping_address"`
	BillingAddress  AddressDTO    `json:"billing_address"`
	PaymentID       string        `json:"payment_id"`
	ShippingID      string        `json:"shipping_id"`
	Notes           string        `json:"notes"`
	PromotionCodes  []string      `json:"promotion_codes"`
	Discounts       []DiscountDTO `json:"discounts"`
	TaxAmount       float64       `json:"tax_amount"`
	CreatedAt       string        `json:"created_at"`
	UpdatedAt       string        `json:"updated_at"`
	CompletedAt     *string       `json:"completed_at,omitempty"`
	CancelledAt     *string       `json:"cancelled_at,omitempty"`
	Version         int           `json:"version"`
}

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	UserID          string          `json:"user_id"`
	Items           []OrderItemDTO  `json:"items"`
	ShippingAddress AddressDTO      `json:"shipping_address"`
	BillingAddress  AddressDTO      `json:"billing_address"`
	Notes           string          `json:"notes"`
	PromotionCodes  []string        `json:"promotion_codes"`
}

// UpdateOrderInput represents the input for updating an order
type UpdateOrderInput struct {
	Notes           *string        `json:"notes,omitempty"`
	ShippingAddress *AddressDTO    `json:"shipping_address,omitempty"`
	BillingAddress  *AddressDTO    `json:"billing_address,omitempty"`
}

// ListOrdersOutput represents the output for listing orders
type ListOrdersOutput struct {
	Orders     []OrderDTO `json:"orders"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// OrderFilterOptions represents options for filtering orders
type OrderFilterOptions struct {
	UserID     string  `json:"user_id,omitempty"`
	Status     string  `json:"status,omitempty"`
	ProductID  string  `json:"product_id,omitempty"`
	StartDate  string  `json:"start_date,omitempty"`
	EndDate    string  `json:"end_date,omitempty"`
	MinAmount  float64 `json:"min_amount,omitempty"`
	MaxAmount  float64 `json:"max_amount,omitempty"`
	Page       int     `json:"page,omitempty"`
	PageSize   int     `json:"page_size,omitempty"`
}
