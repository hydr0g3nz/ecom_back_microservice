package entity

import "time"

// Inventory tracks product stock information
type Inventory struct {
	ID        string     `json:"id"`
	ProductID string     `json:"product_id"`
	Quantity  int        `json:"quantity"`
	Reserved  int        `json:"reserved"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
