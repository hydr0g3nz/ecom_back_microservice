package model

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// ShippingModel represents shipping information in Cassandra
type ShippingModel struct {
	ID                gocql.UUID `json:"id"`
	OrderID           gocql.UUID `json:"order_id"`
	Carrier           string     `json:"carrier"`
	TrackingNumber    string     `json:"tracking_number"`
	Status            string     `json:"status"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	ShippedAt         *time.Time `json:"shipped_at"`
	DeliveredAt       *time.Time `json:"delivered_at"`
	ShippingMethod    string     `json:"shipping_method"`
	ShippingCost      float64    `json:"shipping_cost"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ToEntity converts a Cassandra ShippingModel to a domain Shipping entity
func (sm *ShippingModel) ToEntity() (*entity.Shipping, error) {
	status, err := valueobject.ParseShippingStatus(sm.Status)
	if err != nil {
		return nil, err
	}

	return &entity.Shipping{
		ID:                sm.ID.String(),
		OrderID:           sm.OrderID.String(),
		Carrier:           sm.Carrier,
		TrackingNumber:    sm.TrackingNumber,
		Status:            status,
		EstimatedDelivery: sm.EstimatedDelivery,
		ShippedAt:         sm.ShippedAt,
		DeliveredAt:       sm.DeliveredAt,
		ShippingMethod:    sm.ShippingMethod,
		ShippingCost:      sm.ShippingCost,
		Notes:             sm.Notes,
		CreatedAt:         sm.CreatedAt,
		UpdatedAt:         sm.UpdatedAt,
	}, nil
}

// FromEntity converts a domain Shipping entity to a Cassandra ShippingModel
func FromShippingEntity(shipping *entity.Shipping) (*ShippingModel, error) {
	id, err := gocql.ParseUUID(shipping.ID)
	if err != nil {
		return nil, err
	}

	orderID, err := gocql.ParseUUID(shipping.OrderID)
	if err != nil {
		return nil, err
	}

	return &ShippingModel{
		ID:                id,
		OrderID:           orderID,
		Carrier:           shipping.Carrier,
		TrackingNumber:    shipping.TrackingNumber,
		Status:            shipping.Status.String(),
		EstimatedDelivery: shipping.EstimatedDelivery,
		ShippedAt:         shipping.ShippedAt,
		DeliveredAt:       shipping.DeliveredAt,
		ShippingMethod:    shipping.ShippingMethod,
		ShippingCost:      shipping.ShippingCost,
		Notes:             shipping.Notes,
		CreatedAt:         shipping.CreatedAt,
		UpdatedAt:         shipping.UpdatedAt,
	}, nil
}
