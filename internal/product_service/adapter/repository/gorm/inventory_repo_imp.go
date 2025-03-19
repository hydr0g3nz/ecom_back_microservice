package repository

import (
	"context"
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormInventoryRepository implements InventoryRepository interface using GORM
type GormInventoryRepository struct {
	db *gorm.DB
}

// NewGormInventoryRepository creates a new instance of GormInventoryRepository
func NewGormInventoryRepository(db *gorm.DB) *GormInventoryRepository {
	return &GormInventoryRepository{db: db}
}

// Create stores a new inventory record
func (r *GormInventoryRepository) Create(ctx context.Context, inventory entity.Inventory) (*entity.Inventory, error) {
	inventoryModel := model.NewInventoryModel(&inventory)
	err := r.db.WithContext(ctx).Create(inventoryModel).Error
	if err != nil {
		return nil, err
	}
	return inventoryModel.ToEntity(), nil
}

// GetByProductID retrieves inventory by product ID
func (r *GormInventoryRepository) GetByProductID(ctx context.Context, productID string) (*entity.Inventory, error) {
	var inventoryModel model.Inventory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inventoryModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrInventoryNotFound
		}
		return nil, err
	}
	return inventoryModel.ToEntity(), nil
}

// Update updates an existing inventory
func (r *GormInventoryRepository) Update(ctx context.Context, inventory entity.Inventory) (*entity.Inventory, error) {
	inventoryModel := model.NewInventoryModel(&inventory)
	err := r.db.WithContext(ctx).Save(inventoryModel).Error
	if err != nil {
		return nil, err
	}
	return inventoryModel.ToEntity(), nil
}

// UpdateQuantity updates just the quantity of a product
func (r *GormInventoryRepository) UpdateQuantity(ctx context.Context, productID string, quantity int) error {
	result := r.db.WithContext(ctx).Model(&model.Inventory{}).
		Where("product_id = ?", productID).
		Update("quantity", quantity)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrInventoryNotFound
	}

	return nil
}

// ReserveStock reserves stock for a product (for order processing)
func (r *GormInventoryRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventoryModel model.Inventory

		// Use correct FOR UPDATE clause
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ?", productID).
			First(&inventoryModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return entity.ErrInventoryNotFound
			}
			return err
		}

		availableStock := inventoryModel.Quantity - inventoryModel.Reserved
		if availableStock < quantity {
			return entity.ErrInsufficientStock
		}

		if err := tx.Model(&inventoryModel).
			Update("reserved", inventoryModel.Reserved+quantity).Error; err != nil {
			return err
		}

		return nil
	})
}

// ReleaseReservedStock releases reserved stock back to available
func (r *GormInventoryRepository) ReleaseReservedStock(ctx context.Context, productID string, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventoryModel model.Inventory

		// Use correct FOR UPDATE clause
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ?", productID).
			First(&inventoryModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return entity.ErrInventoryNotFound
			}
			return err
		}

		if inventoryModel.Reserved < quantity {
			quantity = inventoryModel.Reserved
		}

		if err := tx.Model(&inventoryModel).
			Update("reserved", inventoryModel.Reserved-quantity).Error; err != nil {
			return err
		}

		return nil
	})
}
