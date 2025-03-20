package valueobject

import (
	"errors"
	"strings"
)

type ShippingStatus string

const (
	ShippingStatusPending        ShippingStatus = "pending"
	ShippingStatusProcessing     ShippingStatus = "processing"
	ShippingStatusReadyToShip    ShippingStatus = "ready_to_ship"
	ShippingStatusShipped        ShippingStatus = "shipped"
	ShippingStatusInTransit      ShippingStatus = "in_transit"
	ShippingStatusOutForDelivery ShippingStatus = "out_for_delivery"
	ShippingStatusDelivered      ShippingStatus = "delivered"
	ShippingStatusFailed         ShippingStatus = "failed"
	ShippingStatusReturned       ShippingStatus = "returned"
)

func (s ShippingStatus) String() string {
	return string(s)
}

func (s ShippingStatus) IsValid() bool {
	statuses := [...]ShippingStatus{
		ShippingStatusPending, ShippingStatusProcessing, ShippingStatusReadyToShip,
		ShippingStatusShipped, ShippingStatusInTransit, ShippingStatusOutForDelivery,
		ShippingStatusDelivered, ShippingStatusFailed, ShippingStatusReturned,
	}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParseShippingStatus(status string) (ShippingStatus, error) {
	status = strings.ToLower(status)
	if !ShippingStatus(status).IsValid() {
		return "", errors.New("invalid shipping status")
	}
	return ShippingStatus(status), nil
}
