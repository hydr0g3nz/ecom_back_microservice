package query

import (
	"context"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// ListOrdersUsecase defines the interface for listing orders
type ListOrdersUsecase interface {
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error)
	ListByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error)
	ListByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.Order, int, error)
	Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error)
}

// listOrdersUsecase implements the ListOrdersUsecase interface
type listOrdersUsecase struct {
	orderReadRepo repository.OrderReadRepository
}

// NewListOrdersUsecase creates a new instance of listOrdersUsecase
func NewListOrdersUsecase(orderReadRepo repository.OrderReadRepository) ListOrdersUsecase {
	return &listOrdersUsecase{
		orderReadRepo: orderReadRepo,
	}
}

// ListByUser retrieves orders for a specific user
func (uc *listOrdersUsecase) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error) {
	return uc.orderReadRepo.GetByUserID(ctx, userID, page, pageSize)
}

// ListByStatus retrieves orders with a specific status
func (uc *listOrdersUsecase) ListByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error) {
	return uc.orderReadRepo.FindByStatus(ctx, status, page, pageSize)
}

// ListByDateRange retrieves orders within a date range
func (uc *listOrdersUsecase) ListByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.Order, int, error) {
	return uc.orderReadRepo.FindByDateRange(ctx, startDate, endDate, page, pageSize)
}

// Search searches for orders based on various criteria
func (uc *listOrdersUsecase) Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error) {
	return uc.orderReadRepo.Search(ctx, criteria, page, pageSize)
}
