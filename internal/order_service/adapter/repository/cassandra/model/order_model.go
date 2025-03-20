package model

import (
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderModel represents an order in Cassandra
type OrderModel struct {
	ID              gocql.UUID `json:"id"`
	UserID          string     `json:"user_id"`
	Items           []byte     `json:"items"` // Serialized OrderItems
	TotalAmount     float64    `json:"total_amount"`
	Status          string     `json:"status"`
	ShippingAddress []byte     `json:"shipping_address"` // Serialized Address
	BillingAddress  []byte     `json:"billing_address"`  // Serialized Address
	PaymentID       string     `json:"payment_id"`
	ShippingID      string     `json:"shipping_id"`
	Notes           string     `json:"notes"`
	PromotionCodes  []string   `json:"promotion_codes"`
	Discounts       []byte     `json:"discounts"` // Serialized Discounts
	TaxAmount       float64    `json:"tax_amount"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CancelledAt     *time.Time `json:"cancelled_at"`
	Version         int        `json:"version"`
}

// ToEntity converts a Cassandra OrderModel to a domain Order entity
func (om *OrderModel) ToEntity() (*entity.Order, error) {
	var items []entity.OrderItem
	if err := json.Unmarshal(om.Items, &items); err != nil {
		return nil, err
	}

	var shippingAddress entity.Address
	if err := json.Unmarshal(om.ShippingAddress, &shippingAddress); err != nil {
		return nil, err
	}

	var billingAddress entity.Address
	if err := json.Unmarshal(om.BillingAddress, &billingAddress); err != nil {
		return nil, err
	}

	var discounts []entity.Discount
	if err := json.Unmarshal(om.Discounts, &discounts); err != nil {
		return nil, err
	}

	status, err := valueobject.ParseOrderStatus(om.Status)
	if err != nil {
		return nil, err
	}

	return &entity.Order{
		ID:              om.ID.String(),
		UserID:          om.UserID,
		Items:           items,
		TotalAmount:     om.TotalAmount,
		Status:          status,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
		PaymentID:       om.PaymentID,
		ShippingID:      om.ShippingID,
		Notes:           om.Notes,
		PromotionCodes:  om.PromotionCodes,
		Discounts:       discounts,
		TaxAmount:       om.TaxAmount,
		CreatedAt:       om.CreatedAt,
		UpdatedAt:       om.UpdatedAt,
		CompletedAt:     om.CompletedAt,
		CancelledAt:     om.CancelledAt,
		Version:         om.Version,
	}, nil
}

// FromEntity converts a domain Order entity to a Cassandra OrderModel
func FromOrderEntity(order *entity.Order) (*OrderModel, error) {
	id, err := gocql.ParseUUID(order.ID)
	if err != nil {
		return nil, err
	}

	items, err := json.Marshal(order.Items)
	if err != nil {
		return nil, err
	}

	shippingAddress, err := json.Marshal(order.ShippingAddress)
	if err != nil {
		return nil, err
	}

	billingAddress, err := json.Marshal(order.BillingAddress)
	if err != nil {
		return nil, err
	}

	discounts, err := json.Marshal(order.Discounts)
	if err != nil {
		return nil, err
	}

	return &OrderModel{
		ID:              id,
		UserID:          order.UserID,
		Items:           items,
		TotalAmount:     order.TotalAmount,
		Status:          order.Status.String(),
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
		PaymentID:       order.PaymentID,
		ShippingID:      order.ShippingID,
		Notes:           order.Notes,
		PromotionCodes:  order.PromotionCodes,
		Discounts:       discounts,
		TaxAmount:       order.TaxAmount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		CompletedAt:     order.CompletedAt,
		CancelledAt:     order.CancelledAt,
		Version:         order.Version,
	}, nil
}
