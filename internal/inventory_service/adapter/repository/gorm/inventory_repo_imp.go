package repository

import (
	"context"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
	"gorm.io/gorm"
)

// GormInventoryRepository implements InventoryRepository interface using GORM
type GormInventoryRepository struct {
	db *gorm.DB
}

// NewGormInventoryRepository creates a new inventory repository instance
func NewGormInventoryRepository(db *gorm.DB) *GormInventoryRepository {
	return &GormInventoryRepository{db: db}
}

// GetInventoryItem retrieves an inventory item by SKU
func (r *GormInventoryRepository) GetInventoryItem(ctx context.Context, sku string) (*entity.InventoryItem, error) {
	var item model.InventoryItem
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&item).Error
	if err != nil {
		return nil, err
	}
	return item.ToEntity(), nil
}

// CreateInventoryItem creates a new inventory item
func (r *GormInventoryRepository) CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error) {
	itemModel := model.NewInventoryItemModel(item)
	err := r.db.WithContext(ctx).Create(itemModel).Error
	if err != nil {
		return nil, err
	}
	return itemModel.ToEntity(), nil
}

// UpdateInventoryItem updates an existing inventory item
func (r *GormInventoryRepository) UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) (*entity.InventoryItem, error) {
	itemModel := model.NewInventoryItemModel(item)
	itemModel.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(itemModel).Error
	if err != nil {
		return nil, err
	}
	return itemModel.ToEntity(), nil
}

// CreateReservation creates a new inventory reservation
func (r *GormInventoryRepository) CreateReservation(ctx context.Context, reservation *entity.InventoryReservation) (*entity.InventoryReservation, error) {
	reservationModel := model.NewInventoryReservationModel(reservation)
	err := r.db.WithContext(ctx).Create(reservationModel).Error
	if err != nil {
		return nil, err
	}
	return reservationModel.ToEntity(), nil
}

// GetReservationByID retrieves a reservation by ID
func (r *GormInventoryRepository) GetReservationByID(ctx context.Context, reservationID string) (*entity.InventoryReservation, error) {
	var reservation model.InventoryReservation
	err := r.db.WithContext(ctx).Where("reservation_id = ?", reservationID).First(&reservation).Error
	if err != nil {
		return nil, err
	}
	return reservation.ToEntity(), nil
}

// GetReservationsByOrderID retrieves all reservations for an order
func (r *GormInventoryRepository) GetReservationsByOrderID(ctx context.Context, orderID string) ([]*entity.InventoryReservation, error) {
	var reservations []model.InventoryReservation
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&reservations).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.InventoryReservation, len(reservations))
	for i, res := range reservations {
		result[i] = res.ToEntity()
	}
	return result, nil
}

// UpdateReservation updates an existing reservation
func (r *GormInventoryRepository) UpdateReservation(ctx context.Context, reservation *entity.InventoryReservation) (*entity.InventoryReservation, error) {
	reservationModel := model.NewInventoryReservationModel(reservation)
	err := r.db.WithContext(ctx).Save(reservationModel).Error
	if err != nil {
		return nil, err
	}
	return reservationModel.ToEntity(), nil
}

// DeleteReservation deletes a reservation
func (r *GormInventoryRepository) DeleteReservation(ctx context.Context, reservationID string) error {
	return r.db.WithContext(ctx).Where("reservation_id = ?", reservationID).Delete(&model.InventoryReservation{}).Error
}

// RecordStockTransaction records a stock transaction
func (r *GormInventoryRepository) RecordStockTransaction(ctx context.Context, transaction *entity.StockTransaction) (*entity.StockTransaction, error) {
	transactionModel := model.NewStockTransactionModel(transaction)
	err := r.db.WithContext(ctx).Create(transactionModel).Error
	if err != nil {
		return nil, err
	}
	return transactionModel.ToEntity(), nil
}

// GetStockTransactions retrieves stock transactions for a SKU
func (r *GormInventoryRepository) GetStockTransactions(ctx context.Context, sku string, limit, offset int) ([]*entity.StockTransaction, int, error) {
	var transactions []model.StockTransaction
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&model.StockTransaction{}).Where("sku = ?", sku).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Where("sku = ?", sku).
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.StockTransaction, len(transactions))
	for i, tx := range transactions {
		result[i] = tx.ToEntity()
	}
	return result, int(total), nil
}

// GetLowStockItems retrieves items with stock below their reorder level
func (r *GormInventoryRepository) GetLowStockItems(ctx context.Context, limit, offset int) ([]*entity.InventoryItem, int, error) {
	var items []model.InventoryItem
	var total int64

	// Get total count
	err := r.db.WithContext(ctx).Model(&model.InventoryItem{}).
		Where("available_qty <= reorder_level").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = r.db.WithContext(ctx).Where("available_qty <= reorder_level").
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.InventoryItem, len(items))
	for i, item := range items {
		result[i] = item.ToEntity()
	}
	return result, int(total), nil
}
