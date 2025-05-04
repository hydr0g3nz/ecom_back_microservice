// internal/inventory_service/adapter/dto/inventory_dto.go
package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
)

// CreateInventoryItemRequest represents the request body for creating an inventory item
type CreateInventoryItemRequest struct {
	SKU          string `json:"sku" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description"`
	AvailableQty int    `json:"available_qty" validate:"min=0"`
	ReservedQty  int    `json:"reserved_qty" validate:"min=0"`
	SoldQty      int    `json:"sold_qty" validate:"min=0"`
	ReorderLevel int    `json:"reorder_level" validate:"min=0"`
}

// ToEntity converts the request DTO to an InventoryItem entity
func (d *CreateInventoryItemRequest) ToEntity() entity.InventoryItem {
	return entity.InventoryItem{
		SKU:          d.SKU,
		AvailableQty: d.AvailableQty,
		ReservedQty:  d.ReservedQty,
		SoldQty:      d.SoldQty,
		ReorderLevel: d.ReorderLevel,
		UpdatedAt:    time.Now(), // Will be overwritten by usecase
	}
}

// UpdateInventoryItemRequest represents the request body for updating an inventory item
type UpdateInventoryItemRequest struct {
	Name         string `json:"name"` // Allow zero value for optional fields
	Description  string `json:"description"`
	AvailableQty int    `json:"available_qty" validate:"min=0"`
	ReservedQty  int    `json:"reserved_qty" validate:"min=0"`
	SoldQty      int    `json:"sold_qty" validate:"min=0"`
	ReorderLevel int    `json:"reorder_level" validate:"min=0"`
	// Note: SKU is expected from path parameter for update
}

// ToEntity converts the request DTO to an InventoryItem entity (for update)
// It requires the existing SKU to be set
func (d *UpdateInventoryItemRequest) ToEntity(sku string) entity.InventoryItem {
	return entity.InventoryItem{
		SKU:          sku, // Use SKU from path param
		AvailableQty: d.AvailableQty,
		ReservedQty:  d.ReservedQty,
		SoldQty:      d.SoldQty,
		ReorderLevel: d.ReorderLevel,
		// CreatedAt should not be updated here
		UpdatedAt: time.Now(), // Will be overwritten by usecase
	}
}

// AddStockRequest represents the request body for adding stock
type AddStockRequest struct {
	Quantity    int    `json:"quantity" validate:"required,min=1"`
	ReferenceID string `json:"reference_id"`
}

// ReserveStockRequest represents the request body for reserving stock
type ReserveStockRequest struct {
	OrderID string         `json:"order_id" validate:"required"`
	Items   map[string]int `json:"items" validate:"required"` // map[sku]quantity
}

// GetTransactionHistoryRequest represents the query parameters for transaction history
type GetTransactionHistoryRequest struct {
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"pageSize" validate:"min=1"`
}

// GetLowStockItemsRequest represents the query parameters for low stock items
type GetLowStockItemsRequest struct {
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"pageSize" validate:"min=1"`
}

// InventoryItemsWithTotal represents a response structure for paginated lists of inventory items
type InventoryItemsWithTotal struct {
	Items []*entity.InventoryItem `json:"items"`
	Total int                     `json:"total"`
}

// StockTransactionsWithTotal represents a response structure for paginated lists of transactions
type StockTransactionsWithTotal struct {
	Transactions []*entity.StockTransaction `json:"transactions"`
	Total        int                        `json:"total"`
}
