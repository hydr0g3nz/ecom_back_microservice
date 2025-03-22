package mapper

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
)

// ToShippingDTO converts a shipping entity to a shipping DTO
func ToShippingDTO(shipping *entity.Shipping) dto.ShippingDTO {
	shippingDTO := dto.ShippingDTO{
		ID:             shipping.ID.String(),
		OrderID:        shipping.OrderID.String(),
		Carrier:        shipping.Carrier,
		TrackingNumber: shipping.TrackingNumber,
		Status:         shipping.Status.String(),
		ShippingMethod: shipping.ShippingMethod,
		ShippingCost:   shipping.ShippingCost,
		Notes:          shipping.Notes,
		CreatedAt:      shipping.CreatedAt.String(),
		UpdatedAt:      shipping.UpdatedAt.String(),
	}

	if shipping.EstimatedDelivery != nil {
		estimatedDelivery := shipping.EstimatedDelivery.String()
		shippingDTO.EstimatedDelivery = &estimatedDelivery
	}

	if shipping.ShippedAt != nil {
		shippedAt := shipping.ShippedAt.String()
		shippingDTO.ShippedAt = &shippedAt
	}

	if shipping.DeliveredAt != nil {
		deliveredAt := shipping.DeliveredAt.String()
		shippingDTO.DeliveredAt = &deliveredAt
	}

	return shippingDTO
}

// CreateShippingFromDTO creates a shipping entity from an update shipping input DTO
func CreateShippingFromDTO(
	input dto.UpdateShippingInput,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
) (*entity.Shipping, error) {
	// Parse status
	status, err := valueobject.ParseShippingStatus(input.Status)
	if err != nil {
		return nil, err
	}

	// Parse estimated delivery time if provided
	var estimatedDelivery *valueobject.Timestamp
	if input.EstimatedDelivery != nil {
		t, err := time.Parse(time.RFC3339, *input.EstimatedDelivery)
		if err != nil {
			return nil, err
		}
		timestamp := valueobject.NewTimestamp(t)
		estimatedDelivery = &timestamp
	}

	// Create the shipping entity
	return entity.NewShipping(
		idGenerator.NewID(),
		valueobject.ID(input.OrderID),
		input.Carrier,
		input.TrackingNumber,
		status,
		estimatedDelivery,
		input.ShippingMethod,
		input.ShippingCost,
		input.Notes,
		timeProvider,
	)
}
