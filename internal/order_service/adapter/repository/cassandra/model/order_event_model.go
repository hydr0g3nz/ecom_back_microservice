package model

import (
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// OrderEventModel represents an order event in Cassandra
type OrderEventModel struct {
	ID        gocql.UUID `json:"id"`
	OrderID   gocql.UUID `json:"order_id"`
	Type      string     `json:"type"`
	Data      []byte     `json:"data"` // Serialized event data
	Version   int        `json:"version"`
	Timestamp time.Time  `json:"timestamp"`
	UserID    string     `json:"user_id"`
}

// ToEntity converts a Cassandra OrderEventModel to a domain OrderEvent entity
func (oem *OrderEventModel) ToEntity() (*entity.OrderEvent, error) {
	return &entity.OrderEvent{
		ID:        oem.ID.String(),
		OrderID:   oem.OrderID.String(),
		Type:      entity.EventType(oem.Type),
		Data:      json.RawMessage(oem.Data),
		Version:   oem.Version,
		Timestamp: oem.Timestamp,
		UserID:    oem.UserID,
	}, nil
}

// FromEntity converts a domain OrderEvent entity to a Cassandra OrderEventModel
func FromOrderEventEntity(event *entity.OrderEvent) (*OrderEventModel, error) {
	id, err := gocql.ParseUUID(event.ID)
	if err != nil {
		return nil, err
	}

	orderID, err := gocql.ParseUUID(event.OrderID)
	if err != nil {
		return nil, err
	}

	return &OrderEventModel{
		ID:        id,
		OrderID:   orderID,
		Type:      string(event.Type),
		Data:      event.Data,
		Version:   event.Version,
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
	}, nil
}
