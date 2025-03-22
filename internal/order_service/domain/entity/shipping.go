package entity

import (
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// Shipping represents shipping information for an order
type Shipping struct {
	ID                valueobject.ID                 `json:"id"`
	OrderID           valueobject.ID                 `json:"order_id"`
	Carrier           string                         `json:"carrier"`
	TrackingNumber    string                         `json:"tracking_number"`
	Status            valueobject.ShippingStatus     `json:"status"`
	EstimatedDelivery *valueobject.Timestamp         `json:"estimated_delivery"`
	ShippedAt         *valueobject.Timestamp         `json:"shipped_at"`
	DeliveredAt       *valueobject.Timestamp         `json:"delivered_at"`
	ShippingMethod    string                         `json:"shipping_method"`
	ShippingCost      float64                        `json:"shipping_cost"`
	Notes             string                         `json:"notes"`
	CreatedAt         valueobject.Timestamp          `json:"created_at"`
	UpdatedAt         valueobject.Timestamp          `json:"updated_at"`
}

// ValidateShipping validates the shipping information
func ValidateShipping(shipping Shipping) error {
	if shipping.OrderID.String() == "" {
		return errors.New("order ID is required")
	}
	if shipping.ShippingMethod == "" {
		return errors.New("shipping method is required")
	}
	if shipping.ShippingCost < 0 {
		return errors.New("shipping cost cannot be negative")
	}
	if !shipping.Status.IsValid() {
		return errors.New("invalid shipping status")
	}
	return nil
}

// NewShipping creates a new shipping record
func NewShipping(
	id valueobject.ID,
	orderID valueobject.ID,
	carrier string,
	trackingNumber string,
	status valueobject.ShippingStatus,
	estimatedDelivery *valueobject.Timestamp,
	shippingMethod string,
	shippingCost float64,
	notes string,
	timeProvider valueobject.TimeProvider,
) (*Shipping, error) {
	shipping := &Shipping{
		ID:                id,
		OrderID:           orderID,
		Carrier:           carrier,
		TrackingNumber:    trackingNumber,
		Status:            status,
		EstimatedDelivery: estimatedDelivery,
		ShippingMethod:    shippingMethod,
		ShippingCost:      shippingCost,
		Notes:             notes,
		CreatedAt:         timeProvider.Now(),
		UpdatedAt:         timeProvider.Now(),
	}

	// Set timestamps based on status
	if status == valueobject.ShippingStatusShipped {
		shippedAt := timeProvider.Now()
		shipping.ShippedAt = &shippedAt
	} else if status == valueobject.ShippingStatusDelivered {
		deliveredAt := timeProvider.Now()
		shipping.DeliveredAt = &deliveredAt
	}

	// Validate the shipping
	if err := ValidateShipping(*shipping); err != nil {
		return nil, err
	}

	return shipping, nil
}

// UpdateStatus updates the shipping status and related timestamps
func (s *Shipping) UpdateStatus(status valueobject.ShippingStatus, timeProvider valueobject.TimeProvider) error {
	if !status.IsValid() {
		return errors.New("invalid shipping status")
	}

	s.Status = status
	s.UpdatedAt = timeProvider.Now()

	// Update timestamps based on status
	if status == valueobject.ShippingStatusShipped {
		shippedAt := timeProvider.Now()
		s.ShippedAt = &shippedAt
	} else if status == valueobject.ShippingStatusDelivered {
		deliveredAt := timeProvider.Now()
		s.DeliveredAt = &deliveredAt
	}

	return nil
}

// UpdateTrackingInfo updates the tracking information
func (s *Shipping) UpdateTrackingInfo(carrier string, trackingNumber string, timeProvider valueobject.TimeProvider) {
	s.Carrier = carrier
	s.TrackingNumber = trackingNumber
	s.UpdatedAt = timeProvider.Now()
}

// UpdateEstimatedDelivery updates the estimated delivery time
func (s *Shipping) UpdateEstimatedDelivery(estimatedDelivery valueobject.Timestamp, timeProvider valueobject.TimeProvider) {
	s.EstimatedDelivery = &estimatedDelivery
	s.UpdatedAt = timeProvider.Now()
}
