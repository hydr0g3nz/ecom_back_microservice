package entity

import (
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderItem represents an individual product in an order
type OrderItem struct {
	ID           valueobject.ID `json:"id"`
	ProductID    valueobject.ID `json:"product_id"`
	Name         string         `json:"name"`
	SKU          string         `json:"sku"`
	Quantity     int            `json:"quantity"`
	Price        float64        `json:"price"`
	TotalPrice   float64        `json:"total_price"`
	CurrencyCode string         `json:"currency_code"`
}

// ValidateOrderItem validates the order item
func ValidateOrderItem(item OrderItem) error {
	if item.ProductID.String() == "" {
		return errors.New("product ID is required")
	}
	if item.Name == "" {
		return errors.New("item name is required")
	}
	if item.Quantity <= 0 {
		return errors.New("item quantity must be positive")
	}
	if item.Price < 0 {
		return errors.New("item price cannot be negative")
	}
	if item.CurrencyCode == "" {
		return errors.New("currency code is required")
	}
	return nil
}

// NewOrderItem creates a new order item
func NewOrderItem(
	id valueobject.ID,
	productID valueobject.ID,
	name string,
	sku string,
	quantity int,
	price float64,
	currencyCode string,
) (OrderItem, error) {
	item := OrderItem{
		ID:           id,
		ProductID:    productID,
		Name:         name,
		SKU:          sku,
		Quantity:     quantity,
		Price:        price,
		TotalPrice:   price * float64(quantity),
		CurrencyCode: currencyCode,
	}

	if err := ValidateOrderItem(item); err != nil {
		return OrderItem{}, err
	}

	return item, nil
}

// Order represents an order in the system
type Order struct {
	ID              valueobject.ID              `json:"id"`
	UserID          valueobject.ID              `json:"user_id"`
	Items           []OrderItem                 `json:"items"`
	TotalAmount     float64                     `json:"total_amount"`
	Status          valueobject.OrderStatus     `json:"status"`
	ShippingAddress Address                     `json:"shipping_address"`
	BillingAddress  Address                     `json:"billing_address"`
	PaymentID       valueobject.ID              `json:"payment_id"`
	ShippingID      valueobject.ID              `json:"shipping_id"`
	Notes           string                      `json:"notes"`
	PromotionCodes  []string                    `json:"promotion_codes"`
	Discounts       []Discount                  `json:"discounts"`
	TaxAmount       float64                     `json:"tax_amount"`
	CreatedAt       valueobject.Timestamp       `json:"created_at"`
	UpdatedAt       valueobject.Timestamp       `json:"updated_at"`
	CompletedAt     *valueobject.Timestamp      `json:"completed_at"`
	CancelledAt     *valueobject.Timestamp      `json:"cancelled_at"`
	Version         int                         `json:"version"`
}

// ValidateOrder validates the order
func ValidateOrder(order Order) error {
	if order.UserID.String() == "" {
		return errors.New("user ID is required")
	}
	if len(order.Items) == 0 {
		return errors.New("order must have at least one item")
	}
	for _, item := range order.Items {
		if err := ValidateOrderItem(item); err != nil {
			return err
		}
	}
	if err := ValidateAddress(order.ShippingAddress); err != nil {
		return errors.New("invalid shipping address: " + err.Error())
	}
	if err := ValidateAddress(order.BillingAddress); err != nil {
		return errors.New("invalid billing address: " + err.Error())
	}
	if !order.Status.IsValid() {
		return errors.New("invalid order status")
	}
	if order.TaxAmount < 0 {
		return errors.New("tax amount cannot be negative")
	}
	return nil
}

// NewOrder creates a new order
func NewOrder(
	id valueobject.ID,
	userID valueobject.ID,
	items []OrderItem,
	shippingAddress Address,
	billingAddress Address,
	notes string,
	promotionCodes []string,
	timeProvider valueobject.TimeProvider,
) (*Order, error) {
	// Default status for new orders
	status := valueobject.Created

	// Create the order
	order := &Order{
		ID:              id,
		UserID:          userID,
		Items:           items,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
		Notes:           notes,
		PromotionCodes:  promotionCodes,
		Status:          status,
		CreatedAt:       timeProvider.Now(),
		UpdatedAt:       timeProvider.Now(),
		Version:         1,
	}

	// Calculate the total amount
	order.TotalAmount = order.CalculateTotal()

	// Validate the order
	if err := ValidateOrder(*order); err != nil {
		return nil, err
	}

	return order, nil
}

// Address represents a shipping or billing address
type Address struct {
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

// ValidateAddress validates the address
func ValidateAddress(address Address) error {
	if address.FirstName == "" {
		return errors.New("first name is required")
	}
	if address.LastName == "" {
		return errors.New("last name is required")
	}
	if address.AddressLine1 == "" {
		return errors.New("address line 1 is required")
	}
	if address.City == "" {
		return errors.New("city is required")
	}
	if address.PostalCode == "" {
		return errors.New("postal code is required")
	}
	if address.Country == "" {
		return errors.New("country is required")
	}
	return nil
}

// NewAddress creates a new address
func NewAddress(
	firstName string,
	lastName string,
	addressLine1 string,
	addressLine2 string,
	city string,
	state string,
	postalCode string,
	country string,
	phone string,
	email string,
) (Address, error) {
	address := Address{
		FirstName:    firstName,
		LastName:     lastName,
		AddressLine1: addressLine1,
		AddressLine2: addressLine2,
		City:         city,
		State:        state,
		PostalCode:   postalCode,
		Country:      country,
		Phone:        phone,
		Email:        email,
	}

	if err := ValidateAddress(address); err != nil {
		return Address{}, err
	}

	return address, nil
}

// Discount represents a discount applied to an order
type Discount struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // percentage, fixed, shipping
	Amount      float64 `json:"amount"`
}

// ValidateDiscount validates the discount
func ValidateDiscount(discount Discount) error {
	if discount.Code == "" {
		return errors.New("discount code is required")
	}
	if discount.Type != "percentage" && discount.Type != "fixed" && discount.Type != "shipping" {
		return errors.New("invalid discount type")
	}
	if discount.Amount <= 0 {
		return errors.New("discount amount must be positive")
	}
	return nil
}

// NewDiscount creates a new discount
func NewDiscount(code string, description string, discountType string, amount float64) (Discount, error) {
	discount := Discount{
		Code:        code,
		Description: description,
		Type:        discountType,
		Amount:      amount,
	}

	if err := ValidateDiscount(discount); err != nil {
		return Discount{}, err
	}

	return discount, nil
}

// CalculateTotal calculates the total price of the order
func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.TotalPrice
	}

	// Apply discounts
	for _, discount := range o.Discounts {
		if discount.Type == "percentage" {
			total -= total * (discount.Amount / 100.0)
		} else if discount.Type == "fixed" {
			total -= discount.Amount
		}
	}

	// Add tax
	total += o.TaxAmount

	// Ensure non-negative total
	if total < 0 {
		total = 0
	}

	return total
}

// AddItem adds an item to the order
func (o *Order) AddItem(item OrderItem, timeProvider valueobject.TimeProvider) error {
	// Validate the item
	if err := ValidateOrderItem(item); err != nil {
		return err
	}

	// Check if the order is in a valid state for modification
	if o.Status == valueobject.Cancelled || o.Status == valueobject.Completed {
		return ErrInvalidOrderStatus
	}

	// Check if the item already exists
	for i, existingItem := range o.Items {
		if existingItem.ProductID == item.ProductID {
			// Update the quantity and total price
			o.Items[i].Quantity += item.Quantity
			o.Items[i].TotalPrice = o.Items[i].Price * float64(o.Items[i].Quantity)
			o.UpdatedAt = timeProvider.Now()
			o.TotalAmount = o.CalculateTotal()
			o.Version++
			return nil
		}
	}

	// Add new item
	o.Items = append(o.Items, item)
	o.UpdatedAt = timeProvider.Now()
	o.TotalAmount = o.CalculateTotal()
	o.Version++
	return nil
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(productID valueobject.ID, timeProvider valueobject.TimeProvider) error {
	// Check if the order is in a valid state for modification
	if o.Status == valueobject.Cancelled || o.Status == valueobject.Completed {
		return ErrInvalidOrderStatus
	}

	for i, item := range o.Items {
		if item.ProductID == productID {
			// Remove the item
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.UpdatedAt = timeProvider.Now()
			o.TotalAmount = o.CalculateTotal()
			o.Version++
			return nil
		}
	}
	return ErrItemNotFound
}

// UpdateItemQuantity updates the quantity of an item
func (o *Order) UpdateItemQuantity(productID valueobject.ID, quantity int, timeProvider valueobject.TimeProvider) error {
	// Check if the order is in a valid state for modification
	if o.Status == valueobject.Cancelled || o.Status == valueobject.Completed {
		return ErrInvalidOrderStatus
	}

	if quantity < 1 {
		return o.RemoveItem(productID, timeProvider)
	}

	for i, item := range o.Items {
		if item.ProductID == productID {
			o.Items[i].Quantity = quantity
			o.Items[i].TotalPrice = o.Items[i].Price * float64(quantity)
			o.UpdatedAt = timeProvider.Now()
			o.TotalAmount = o.CalculateTotal()
			o.Version++
			return nil
		}
	}
	return ErrItemNotFound
}

// ApplyDiscount applies a discount to the order
func (o *Order) ApplyDiscount(discount Discount, timeProvider valueobject.TimeProvider) error {
	// Validate the discount
	if err := ValidateDiscount(discount); err != nil {
		return err
	}

	// Check if the order is in a valid state for modification
	if o.Status == valueobject.Cancelled || o.Status == valueobject.Completed {
		return ErrInvalidOrderStatus
	}

	// Check if the discount already exists
	for i, existingDiscount := range o.Discounts {
		if existingDiscount.Code == discount.Code {
			o.Discounts[i] = discount
			o.UpdatedAt = timeProvider.Now()
			o.TotalAmount = o.CalculateTotal()
			o.Version++
			return nil
		}
	}

	// Add new discount
	o.Discounts = append(o.Discounts, discount)
	o.PromotionCodes = append(o.PromotionCodes, discount.Code)
	o.UpdatedAt = timeProvider.Now()
	o.TotalAmount = o.CalculateTotal()
	o.Version++
	return nil
}

// CanCancel checks if the order can be cancelled
func (o *Order) CanCancel() bool {
	return o.Status == valueobject.Created ||
		o.Status == valueobject.Pending ||
		o.Status == valueobject.PaymentPending
}

// CanShip checks if the order can be shipped
func (o *Order) CanShip() bool {
	return o.Status == valueobject.PaymentCompleted ||
		o.Status == valueobject.Processing
}

// UpdateStatus updates the order status and related timestamps
func (o *Order) UpdateStatus(status valueobject.OrderStatus, timeProvider valueobject.TimeProvider) error {
	// Validate the status transition
	if !valueobject.IsValidTransition(o.Status, status) {
		return ErrInvalidOrderStatus
	}

	o.Status = status
	o.UpdatedAt = timeProvider.Now()
	o.Version++

	// Update timestamps based on status
	now := timeProvider.Now()
	if status == valueobject.Completed {
		o.CompletedAt = &now
	} else if status == valueobject.Cancelled {
		o.CancelledAt = &now
	}

	return nil
}
