package entity

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// Shipping represents shipping information for an order
type Shipping struct {
	ID                string                     `json:"id"`
	OrderID           string                     `json:"order_id"`
	Carrier           string                     `json:"carrier"`
	TrackingNumber    string                     `json:"tracking_number"`
	Status            valueobject.ShippingStatus `json:"status"`
	EstimatedDelivery *time.Time                 `json:"estimated_delivery"`
	ShippedAt         *time.Time                 `json:"shipped_at"`
	DeliveredAt       *time.Time                 `json:"delivered_at"`
	ShippingMethod    string                     `json:"shipping_method"`
	ShippingCost      float64                    `json:"shipping_cost"`
	Notes             string                     `json:"notes"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

// UpdateStatus updates the shipping status and related timestamps
func (s *Shipping) UpdateStatus(status valueobject.ShippingStatus) {
	s.Status = status
	s.UpdatedAt = time.Now()

	now := time.Now()
	if status == valueobject.ShippingStatusShipped {
		s.ShippedAt = &now
	} else if status == valueobject.ShippingStatusDelivered {
		s.DeliveredAt = &now
	}
}
