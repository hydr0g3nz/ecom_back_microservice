package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
)

// InventoryRepository defines the interface for inventory data persistence operations
type InventoryRepository interface {
	// GetInventoryItem retrieves an inventory item by SKU
	GetInventoryItem(ctx context.Context, sku string) (*entity.InventoryItem, error)

	// CreateInventoryItem creates a new inventory item
	CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error)

	// UpdateInventoryItem updates an existing inventory item
	UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error)

	// CreateReservation creates a new inventory reservation
	CreateReservation(ctx context.Context, reservation *entity.InventoryReservation) (*entity.InventoryReservation, error)

	// GetReservationByID retrieves a reservation by ID
	GetReservationByID(ctx context.Context, reservationID string) (*entity.InventoryReservation, error)

	// GetReservationsByOrderID retrieves all reservations for an order
	GetReservationsByOrderID(ctx context.Context, orderID string) ([]*entity.InventoryReservation, error)

	// UpdateReservation updates an existing reservation
	UpdateReservation(ctx context.Context, reservation *entity.InventoryReservation) (*entity.InventoryReservation, error)

	// DeleteReservation deletes a reservation
	DeleteReservation(ctx context.Context, reservationID string) error

	// RecordStockTransaction records a stock transaction
	RecordStockTransaction(ctx context.Context, transaction *entity.StockTransaction) (*entity.StockTransaction, error)

	// GetStockTransactions retrieves stock transactions for a SKU
	GetStockTransactions(ctx context.Context, sku string, limit, offset int) ([]*entity.StockTransaction, int, error)

	// GetLowStockItems retrieves items with stock below their reorder level
	GetLowStockItems(ctx context.Context, limit, offset int) ([]*entity.InventoryItem, int, error)
}
