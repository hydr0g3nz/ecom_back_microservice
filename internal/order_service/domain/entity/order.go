package entity

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderItem represents an individual product in an order
type OrderItem struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	TotalPrice   float64 `json:"total_price"`
	CurrencyCode string  `json:"currency_code"`
}

// Order represents an order in the system
type Order struct {
	ID              string                  `json:"id"`
	UserID          string                  `json:"user_id"`
	Items           []OrderItem             `json:"items"`
	TotalAmount     float64                 `json:"total_amount"`
	Status          valueobject.OrderStatus `json:"status"`
	ShippingAddress Address                 `json:"shipping_address"`
	BillingAddress  Address                 `json:"billing_address"`
	PaymentID       string                  `json:"payment_id"`
	ShippingID      string                  `json:"shipping_id"`
	Notes           string                  `json:"notes"`
	PromotionCodes  []string                `json:"promotion_codes"`
	Discounts       []Discount              `json:"discounts"`
	TaxAmount       float64                 `json:"tax_amount"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
	CancelledAt     *time.Time              `json:"cancelled_at"`
	Version         int                     `json:"version"`
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

// Discount represents a discount applied to an order
type Discount struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // percentage, fixed, shipping
	Amount      float64 `json:"amount"`
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
func (o *Order) AddItem(item OrderItem) {
	// Check if the item already exists
	for i, existingItem := range o.Items {
		if existingItem.ProductID == item.ProductID {
			// Update the quantity and total price
			o.Items[i].Quantity += item.Quantity
			o.Items[i].TotalPrice = o.Items[i].Price * float64(o.Items[i].Quantity)
			o.UpdatedAt = time.Now()
			o.TotalAmount = o.CalculateTotal()
			return
		}
	}

	// Add new item
	o.Items = append(o.Items, item)
	o.UpdatedAt = time.Now()
	o.TotalAmount = o.CalculateTotal()
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(productID string) bool {
	for i, item := range o.Items {
		if item.ProductID == productID {
			// Remove the item
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.UpdatedAt = time.Now()
			o.TotalAmount = o.CalculateTotal()
			return true
		}
	}
	return false
}

// UpdateItemQuantity updates the quantity of an item
func (o *Order) UpdateItemQuantity(productID string, quantity int) bool {
	if quantity < 1 {
		return o.RemoveItem(productID)
	}

	for i, item := range o.Items {
		if item.ProductID == productID {
			o.Items[i].Quantity = quantity
			o.Items[i].TotalPrice = o.Items[i].Price * float64(quantity)
			o.UpdatedAt = time.Now()
			o.TotalAmount = o.CalculateTotal()
			return true
		}
	}
	return false
}

// ApplyDiscount applies a discount to the order
func (o *Order) ApplyDiscount(discount Discount) {
	// Check if the discount already exists
	for i, existingDiscount := range o.Discounts {
		if existingDiscount.Code == discount.Code {
			o.Discounts[i] = discount
			o.UpdatedAt = time.Now()
			o.TotalAmount = o.CalculateTotal()
			return
		}
	}

	// Add new discount
	o.Discounts = append(o.Discounts, discount)
	o.PromotionCodes = append(o.PromotionCodes, discount.Code)
	o.UpdatedAt = time.Now()
	o.TotalAmount = o.CalculateTotal()
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
func (o *Order) UpdateStatus(status valueobject.OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
	o.Version++

	// Update timestamps based on status
	now := time.Now()
	if status == valueobject.Completed {
		o.CompletedAt = &now
	} else if status == valueobject.Cancelled {
		o.CancelledAt = &now
	}
}
