package entity

import (
	"time"
)

// InventoryItem tracks the main stock information of each ProductID
type InventoryItem struct {
	ProductID    string    `json:"product_id"`
	AvailableQty int       `json:"available_qty"`
	ReservedQty  int       `json:"reserved_qty"`
	SoldQty      int       `json:"sold_qty"`
	ReorderLevel int       `json:"reorder_level"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InventoryReservation tracks each reservation of a product
type InventoryReservation struct {
	ReservationID string    `json:"reservation_id"`
	OrderID       string    `json:"order_id"`
	ProductID     string    `json:"product_id"`
	Qty           int       `json:"qty"`
	Status        string    `json:"status"`
	ReservedAt    time.Time `json:"reserved_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// StockTransaction tracks the stock transactions
type StockTransaction struct {
	TransactionID string    `json:"transaction_id"`
	ProductID     string    `json:"product_id"`
	Type          string    `json:"type"`
	Qty           int       `json:"qty"`
	OccurredAt    time.Time `json:"occurred_at"`
	ReferenceID   *string   `json:"reference_id"`
}
