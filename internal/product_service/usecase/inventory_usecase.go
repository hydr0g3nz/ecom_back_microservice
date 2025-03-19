package usecase

import (
	"context"
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

type InventoryUsecase interface {
	// GetInventory retrieves inventory for a product
	GetInventory(ctx context.Context, productID string) (*entity.Inventory, error)

	// UpdateInventory updates the inventory for a product
	UpdateInventory(ctx context.Context, productID string, quantity int) error

	// ReserveStock reserves stock for a product (e.g., when added to cart)
	ReserveStock(ctx context.Context, productID string, quantity int) error

	// ConfirmReservation confirms a reservation (e.g., after successful payment)
	ConfirmReservation(ctx context.Context, productID string, quantity int) error

	// CancelReservation cancels a reservation (e.g., when removing from cart)
	CancelReservation(ctx context.Context, productID string, quantity int) error

	// IsInStock checks if a product is in stock
	IsInStock(ctx context.Context, productID string, quantity int) (bool, error)
}

// inventoryUsecase implements the InventoryUsecase interface
type inventoryUsecase struct {
	inventoryRepo repository.InventoryRepository
	productRepo   repository.ProductRepository
	errBuilder    *utils.ErrorBuilder
}

// NewInventoryUsecase creates a new instance of InventoryUsecase
func NewInventoryUsecase(
	ir repository.InventoryRepository,
	pr repository.ProductRepository,
) InventoryUsecase {
	return &inventoryUsecase{
		inventoryRepo: ir,
		productRepo:   pr,
		errBuilder:    utils.NewErrorBuilder("InventoryUsecase"),
	}
}

// GetInventory retrieves inventory for a product
func (iu *inventoryUsecase) GetInventory(ctx context.Context, productID string) (*entity.Inventory, error) {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	inventory, err := iu.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, iu.errBuilder.Err(err)
	}
	return inventory, nil
}

// UpdateInventory updates the inventory for a product
func (iu *inventoryUsecase) UpdateInventory(ctx context.Context, productID string, quantity int) error {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Get current inventory to check if it exists
	inventory, err := iu.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	// Ensure quantity is not negative
	if quantity < 0 {
		return iu.errBuilder.Err(errors.New("quantity cannot be negative"))
	}

	// Update inventory with new quantity
	inventory.Quantity = quantity
	_, err = iu.inventoryRepo.Update(ctx, *inventory)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	return nil
}

// ReserveStock reserves stock for a product
func (iu *inventoryUsecase) ReserveStock(ctx context.Context, productID string, quantity int) error {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Check if there's enough stock
	inventory, err := iu.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	// Check if we have enough available stock (total - reserved)
	availableStock := inventory.Quantity - inventory.Reserved
	if availableStock < quantity {
		return iu.errBuilder.Err(entity.ErrInsufficientStock)
	}

	// Reserve the stock
	err = iu.inventoryRepo.ReserveStock(ctx, productID, quantity)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	return nil
}

// ConfirmReservation confirms a reservation and decreases the actual stock
func (iu *inventoryUsecase) ConfirmReservation(ctx context.Context, productID string, quantity int) error {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Get current inventory
	inventory, err := iu.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	// Check if there's enough reserved stock
	if inventory.Reserved < quantity {
		return iu.errBuilder.Err(errors.New("not enough reserved stock to confirm"))
	}

	// Update inventory: decrease both quantity and reserved amount
	inventory.Quantity -= quantity
	inventory.Reserved -= quantity
	_, err = iu.inventoryRepo.Update(ctx, *inventory)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	return nil
}

// CancelReservation cancels a reservation and releases the stock
func (iu *inventoryUsecase) CancelReservation(ctx context.Context, productID string, quantity int) error {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Release the reserved stock
	err = iu.inventoryRepo.ReleaseReservedStock(ctx, productID, quantity)
	if err != nil {
		return iu.errBuilder.Err(err)
	}

	return nil
}

// IsInStock checks if a product is in stock
func (iu *inventoryUsecase) IsInStock(ctx context.Context, productID string, quantity int) (bool, error) {
	// Ensure the product exists
	_, err := iu.productRepo.GetByID(ctx, productID)
	if err != nil {
		return false, iu.errBuilder.Err(entity.ErrProductNotFound)
	}

	// Get inventory
	inventory, err := iu.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return false, iu.errBuilder.Err(err)
	}

	// Check if available stock is sufficient
	availableStock := inventory.Quantity - inventory.Reserved
	return availableStock >= quantity, nil
}
