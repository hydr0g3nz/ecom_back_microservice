package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

// InventoryUsecase defines the interface for inventory operations
type InventoryUsecase interface {
	// GetInventoryItem retrieves inventory information for a product SKU
	GetInventoryItem(ctx context.Context, sku string) (*entity.InventoryItem, error)

	// CreateInventoryItem creates a new inventory item
	CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error)

	// UpdateInventoryItem updates an existing inventory item
	UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error)

	// AddStock adds stock to an inventory item
	AddStock(ctx context.Context, sku string, quantity int, referenceID string) (*entity.InventoryItem, error)

	// ReserveStock reserves stock for an order
	ReserveStock(ctx context.Context, orderID string, items map[string]int) ([]*entity.InventoryReservation, error)

	// CompleteReservation marks a reservation as completed and deducts stock
	CompleteReservation(ctx context.Context, orderID string) error

	// CancelReservation cancels a reservation and releases stock
	CancelReservation(ctx context.Context, orderID string) error

	// GetReservationsByOrderID gets all reservations for an order
	GetReservationsByOrderID(ctx context.Context, orderID string) ([]*entity.InventoryReservation, error)

	// GetStockTransactionHistory gets the transaction history for a SKU
	GetStockTransactionHistory(ctx context.Context, sku string, page, pageSize int) ([]*entity.StockTransaction, int, error)

	// GetLowStockItems gets items with stock below reorder level
	GetLowStockItems(ctx context.Context, page, pageSize int) ([]*entity.InventoryItem, int, error)
}

// inventoryUsecase implements the InventoryUsecase interface
type inventoryUsecase struct {
	repo       repository.InventoryRepository
	eventPub   service.EventPublisherService
	errBuilder *utils.ErrorBuilder
}

// NewInventoryUsecase creates a new instance of InventoryUsecase
func NewInventoryUsecase(
	repo repository.InventoryRepository,
	eventPub service.EventPublisherService,
) InventoryUsecase {
	return &inventoryUsecase{
		repo:       repo,
		eventPub:   eventPub,
		errBuilder: utils.NewErrorBuilder("InventoryUsecase"),
	}
}

// GetInventoryItem retrieves inventory information for a product SKU
func (iu *inventoryUsecase) GetInventoryItem(ctx context.Context, sku string) (*entity.InventoryItem, error) {
	item, err := iu.repo.GetInventoryItem(ctx, sku)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}
	return item, nil
}

// CreateInventoryItem creates a new inventory item
func (iu *inventoryUsecase) CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error) {
	// Set current time
	item.UpdatedAt = time.Now()

	// Create inventory item
	newItem, err := iu.repo.CreateInventoryItem(ctx, item)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}

	// Create initial stock transaction
	if item.AvailableQty > 0 {
		transaction := &entity.StockTransaction{
			TransactionID: uuid.New().String(),
			SKU:           item.SKU,
			Type:          valueobject.StockTypeReleased.String(),
			Qty:           item.AvailableQty,
			OccurredAt:    time.Now(),
		}

		_, err = iu.repo.RecordStockTransaction(ctx, transaction)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error recording initial stock transaction: %v\n", err)
		}
	}

	// Publish stock updated event
	if err := iu.eventPub.PublishStockUpdated(ctx, newItem); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing stock updated event: %v\n", err)
	}

	// Check if stock is below reorder level
	if newItem.AvailableQty <= newItem.ReorderLevel {
		if err := iu.eventPub.PublishStockLow(ctx, newItem); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock low event: %v\n", err)
		}
	}

	return newItem, nil
}

// UpdateInventoryItem updates an existing inventory item
func (iu *inventoryUsecase) UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error) {
	// Ensure the item exists
	existingItem, err := iu.repo.GetInventoryItem(ctx, item.SKU)
	if err != nil {
		return nil, iu.errBuilder.Err(entity.ErrInventoryNotFound)
	}

	// Check if stock level is changing
	stockChanged := existingItem.AvailableQty != item.AvailableQty

	// Set updated time
	item.UpdatedAt = time.Now()

	// Update inventory item
	updatedItem, err := iu.repo.UpdateInventoryItem(ctx, item)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}

	// Record stock transaction if stock level changed
	if stockChanged {
		// Calculate difference
		diff := item.AvailableQty - existingItem.AvailableQty

		var transactionType string
		if diff > 0 {
			transactionType = valueobject.StockTypeReleased.String()
		} else {
			transactionType = valueobject.StockTypeDeducted.String()
			diff = -diff // Make positive for recording
		}

		transaction := &entity.StockTransaction{
			TransactionID: uuid.New().String(),
			SKU:           item.SKU,
			Type:          transactionType,
			Qty:           diff,
			OccurredAt:    time.Now(),
		}

		_, err = iu.repo.RecordStockTransaction(ctx, transaction)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error recording stock transaction: %v\n", err)
		}

		// Publish appropriate event
		if transactionType == valueobject.StockTypeDeducted.String() {
			if err := iu.eventPub.PublishStockDeducted(ctx, transaction); err != nil {
				fmt.Printf("Error publishing stock deducted event: %v\n", err)
			}
		}
	}

	// Publish stock updated event
	if err := iu.eventPub.PublishStockUpdated(ctx, updatedItem); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing stock updated event: %v\n", err)
	}

	// Check if stock is below reorder level
	if updatedItem.AvailableQty <= updatedItem.ReorderLevel {
		if err := iu.eventPub.PublishStockLow(ctx, updatedItem); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock low event: %v\n", err)
		}
	}

	return updatedItem, nil
}

// AddStock adds stock to an inventory item
func (iu *inventoryUsecase) AddStock(ctx context.Context, sku string, quantity int, referenceID string) (*entity.InventoryItem, error) {
	if quantity <= 0 {
		return nil, iu.errBuilder.Err(entity.ErrInvalidProductData)
	}

	// Get the current inventory item
	item, err := iu.repo.GetInventoryItem(ctx, sku)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}

	// Update available quantity
	item.AvailableQty += quantity
	item.UpdatedAt = time.Now()

	// Update inventory item
	updatedItem, err := iu.repo.UpdateInventoryItem(ctx, item)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}

	// Record stock transaction
	refPtr := &referenceID
	if referenceID == "" {
		refPtr = nil
	}

	transaction := &entity.StockTransaction{
		TransactionID: uuid.New().String(),
		SKU:           sku,
		Type:          valueobject.StockTypeReleased.String(),
		Qty:           quantity,
		OccurredAt:    time.Now(),
		ReferenceID:   refPtr,
	}

	_, err = iu.repo.RecordStockTransaction(ctx, transaction)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error recording stock transaction: %v\n", err)
	}

	// Publish stock updated event
	if err := iu.eventPub.PublishStockUpdated(ctx, updatedItem); err != nil {
		// Log error but continue
		fmt.Printf("Error publishing stock updated event: %v\n", err)
	}

	return updatedItem, nil
}

// ReserveStock reserves stock for an order
func (iu *inventoryUsecase) ReserveStock(ctx context.Context, orderID string, items map[string]int) ([]*entity.InventoryReservation, error) {
	reservations := make([]*entity.InventoryReservation, 0, len(items))

	// First check if all items have sufficient stock
	for sku, qty := range items {
		// Get inventory item
		inventoryItem, err := iu.repo.GetInventoryItem(ctx, sku)
		if err != nil {
			return nil, iu.errBuilder.Err(err)
		}

		// Check if there's enough available stock
		if inventoryItem.AvailableQty < qty {
			// Publish reservation failed event
			iu.eventPub.PublishStockReservationFailed(ctx, orderID, sku, "Insufficient stock")
			return nil, iu.errBuilder.Err(entity.ErrInsufficientStock)
		}
	}

	// Now we know we have enough stock for all items, proceed with reservations
	for sku, qty := range items {
		// Get inventory item
		inventoryItem, err := iu.repo.GetInventoryItem(ctx, sku)
		if err != nil {
			// Should not happen as we already checked
			return nil, iu.errBuilder.Err(err)
		}

		// Create reservation with 30-minute expiry
		reservation := &entity.InventoryReservation{
			ReservationID: uuid.New().String(),
			OrderID:       orderID,
			SKU:           sku,
			Qty:           qty,
			Status:        valueobject.ReserveStatusReserved.String(),
			ReservedAt:    time.Now(),
			ExpiresAt:     time.Now().Add(30 * time.Minute),
		}

		// Save the reservation
		createdReservation, err := iu.repo.CreateReservation(ctx, reservation)
		if err != nil {
			return nil, iu.errBuilder.Err(err)
		}

		// Update inventory item
		inventoryItem.AvailableQty -= qty
		inventoryItem.ReservedQty += qty
		inventoryItem.UpdatedAt = time.Now()

		_, err = iu.repo.UpdateInventoryItem(ctx, inventoryItem)
		if err != nil {
			// TODO: In a real system, implement compensating transaction to roll back the reservation
			return nil, iu.errBuilder.Err(err)
		}

		// Record stock transaction
		refID := createdReservation.ReservationID
		transaction := &entity.StockTransaction{
			TransactionID: uuid.New().String(),
			SKU:           sku,
			Type:          valueobject.StockTypeReserved.String(),
			Qty:           qty,
			OccurredAt:    time.Now(),
			ReferenceID:   &refID,
		}

		_, err = iu.repo.RecordStockTransaction(ctx, transaction)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error recording stock transaction: %v\n", err)
		}

		// Publish stock reserved event
		if err := iu.eventPub.PublishStockReserved(ctx, createdReservation); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock reserved event: %v\n", err)
		}

		reservations = append(reservations, createdReservation)

		// Check if stock is below reorder level after reservation
		if inventoryItem.AvailableQty <= inventoryItem.ReorderLevel {
			if err := iu.eventPub.PublishStockLow(ctx, inventoryItem); err != nil {
				// Log error but continue
				fmt.Printf("Error publishing stock low event: %v\n", err)
			}
		}
	}

	return reservations, nil
}

// CompleteReservation marks a reservation as completed and deducts stock
func (iu *inventoryUsecase) CompleteReservation(ctx context.Context, orderID string) error {
	// Get all reservations for the order
	reservations, err := iu.repo.GetReservationsByOrderID(ctx, orderID)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	if len(reservations) == 0 {
		return iu.errBuilder.Err(entity.ErrInventoryNotFound)
	}

	for _, reservation := range reservations {
		// Only process if the reservation is still in RESERVED status
		if reservation.Status != valueobject.ReserveStatusReserved.String() {
			continue
		}

		// Get inventory item
		inventoryItem, err := iu.repo.GetInventoryItem(ctx, reservation.SKU)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Update inventory item: move from reserved to sold
		inventoryItem.ReservedQty -= reservation.Qty
		inventoryItem.SoldQty += reservation.Qty
		inventoryItem.UpdatedAt = time.Now()

		_, err = iu.repo.UpdateInventoryItem(ctx, inventoryItem)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Update reservation status
		reservation.Status = valueobject.ReserveStatusCompleted.String()
		_, err = iu.repo.UpdateReservation(ctx, reservation)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Record stock transaction
		refID := reservation.ReservationID
		transaction := &entity.StockTransaction{
			TransactionID: uuid.New().String(),
			SKU:           reservation.SKU,
			Type:          valueobject.StockTypeDeducted.String(),
			Qty:           reservation.Qty,
			OccurredAt:    time.Now(),
			ReferenceID:   &refID,
		}

		_, err = iu.repo.RecordStockTransaction(ctx, transaction)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error recording stock transaction: %v\n", err)
		}

		// Publish stock deducted event
		if err := iu.eventPub.PublishStockDeducted(ctx, transaction); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock deducted event: %v\n", err)
		}
	}

	return nil
}

// CancelReservation cancels a reservation and releases stock
func (iu *inventoryUsecase) CancelReservation(ctx context.Context, orderID string) error {
	// Get all reservations for the order
	reservations, err := iu.repo.GetReservationsByOrderID(ctx, orderID)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	if len(reservations) == 0 {
		return iu.errBuilder.Err(entity.ErrInventoryNotFound)
	}

	for _, reservation := range reservations {
		// Only process if the reservation is still in RESERVED status
		if reservation.Status != valueobject.ReserveStatusReserved.String() {
			continue
		}

		// Get inventory item
		inventoryItem, err := iu.repo.GetInventoryItem(ctx, reservation.SKU)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Update inventory item
		inventoryItem.AvailableQty += reservation.Qty
		inventoryItem.ReservedQty -= reservation.Qty
		inventoryItem.UpdatedAt = time.Now()

		_, err = iu.repo.UpdateInventoryItem(ctx, inventoryItem)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Update reservation status
		reservation.Status = valueobject.ReserveStatusCancelled.String()
		_, err = iu.repo.UpdateReservation(ctx, reservation)
		if err != nil {
			return iu.errBuilder.Err(err)
		}

		// Record stock transaction
		refID := reservation.ReservationID
		transaction := &entity.StockTransaction{
			TransactionID: uuid.New().String(),
			SKU:           reservation.SKU,
			Type:          valueobject.StockTypeReleased.String(),
			Qty:           reservation.Qty,
			OccurredAt:    time.Now(),
			ReferenceID:   &refID,
		}

		_, err = iu.repo.RecordStockTransaction(ctx, transaction)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error recording stock transaction: %v\n", err)
		}

		// Publish stock released event
		if err := iu.eventPub.PublishStockReleased(ctx, reservation); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock released event: %v\n", err)
		}
	}

	return nil
}

// GetReservationsByOrderID gets all reservations for an order
func (iu *inventoryUsecase) GetReservationsByOrderID(ctx context.Context, orderID string) ([]*entity.InventoryReservation, error) {
	reservations, err := iu.repo.GetReservationsByOrderID(ctx, orderID)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}
	return reservations, nil
}

// GetStockTransactionHistory gets the transaction history for a SKU
func (iu *inventoryUsecase) GetStockTransactionHistory(ctx context.Context, sku string, page, pageSize int) ([]*entity.StockTransaction, int, error) {
	offset := (page - 1) * pageSize
	transactions, total, err := iu.repo.GetStockTransactions(ctx, sku, pageSize, offset)
	if err != nil {
		return nil, 0, iu.errBuilder.Err(err)
	}
	return transactions, total, nil
}

// GetLowStockItems gets items with stock below reorder level
func (iu *inventoryUsecase) GetLowStockItems(ctx context.Context, page, pageSize int) ([]*entity.InventoryItem, int, error) {
	offset := (page - 1) * pageSize
	items, total, err := iu.repo.GetLowStockItems(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, iu.errBuilder.Err(err)
	}
	return items, total, nil
}
