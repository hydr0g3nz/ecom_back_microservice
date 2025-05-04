package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
)

// InventoryItem is the GORM model for inventory items
type InventoryItem struct {
	SKU          string    `gorm:"primaryKey"`
	AvailableQty int       `gorm:"not null"`
	ReservedQty  int       `gorm:"not null"`
	SoldQty      int       `gorm:"not null"`
	ReorderLevel int       `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// ToEntity converts a GORM model to a domain entity
func (m *InventoryItem) ToEntity() *entity.InventoryItem {
	return &entity.InventoryItem{
		SKU:          m.SKU,
		AvailableQty: m.AvailableQty,
		ReservedQty:  m.ReservedQty,
		SoldQty:      m.SoldQty,
		ReorderLevel: m.ReorderLevel,
		UpdatedAt:    m.UpdatedAt,
	}
}

// NewInventoryItemModel creates a new GORM model from a domain entity
func NewInventoryItemModel(item *entity.InventoryItem) *InventoryItem {
	return &InventoryItem{
		SKU:          item.SKU,
		AvailableQty: item.AvailableQty,
		ReservedQty:  item.ReservedQty,
		SoldQty:      item.SoldQty,
		ReorderLevel: item.ReorderLevel,
		UpdatedAt:    item.UpdatedAt,
	}
}

// InventoryReservation is the GORM model for inventory reservations
type InventoryReservation struct {
	ReservationID string    `gorm:"primaryKey"`
	OrderID       string    `gorm:"index;not null"`
	SKU           string    `gorm:"index;not null"`
	Qty           int       `gorm:"not null"`
	Status        string    `gorm:"not null"`
	ReservedAt    time.Time `gorm:"not null"`
	ExpiresAt     time.Time `gorm:"not null"`
}

// ToEntity converts a GORM model to a domain entity
func (m *InventoryReservation) ToEntity() *entity.InventoryReservation {
	return &entity.InventoryReservation{
		ReservationID: m.ReservationID,
		OrderID:       m.OrderID,
		SKU:           m.SKU,
		Qty:           m.Qty,
		Status:        m.Status,
		ReservedAt:    m.ReservedAt,
		ExpiresAt:     m.ExpiresAt,
	}
}

// NewInventoryReservationModel creates a new GORM model from a domain entity
func NewInventoryReservationModel(reservation *entity.InventoryReservation) *InventoryReservation {
	return &InventoryReservation{
		ReservationID: reservation.ReservationID,
		OrderID:       reservation.OrderID,
		SKU:           reservation.SKU,
		Qty:           reservation.Qty,
		Status:        reservation.Status,
		ReservedAt:    reservation.ReservedAt,
		ExpiresAt:     reservation.ExpiresAt,
	}
}

// StockTransaction is the GORM model for stock transactions
type StockTransaction struct {
	TransactionID string    `gorm:"primaryKey"`
	SKU           string    `gorm:"index;not null"`
	Type          string    `gorm:"not null"`
	Qty           int       `gorm:"not null"`
	OccurredAt    time.Time `gorm:"not null;index"`
	ReferenceID   *string
}

// ToEntity converts a GORM model to a domain entity
func (m *StockTransaction) ToEntity() *entity.StockTransaction {
	return &entity.StockTransaction{
		TransactionID: m.TransactionID,
		SKU:           m.SKU,
		Type:          m.Type,
		Qty:           m.Qty,
		OccurredAt:    m.OccurredAt,
		ReferenceID:   m.ReferenceID,
	}
}

// NewStockTransactionModel creates a new GORM model from a domain entity
func NewStockTransactionModel(transaction *entity.StockTransaction) *StockTransaction {
	return &StockTransaction{
		TransactionID: transaction.TransactionID,
		SKU:           transaction.SKU,
		Type:          transaction.Type,
		Qty:           transaction.Qty,
		OccurredAt:    transaction.OccurredAt,
		ReferenceID:   transaction.ReferenceID,
	}
}
