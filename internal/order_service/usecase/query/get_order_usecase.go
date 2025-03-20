package query

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
)

// GetOrderUsecase defines the interface for retrieving an order
type GetOrderUsecase interface {
	Execute(ctx context.Context, orderID string) (*entity.Order, error)
}

// getOrderUsecase implements the GetOrderUsecase interface
type getOrderUsecase struct {
	orderReadRepo repository.OrderReadRepository
}

// NewGetOrderUsecase creates a new instance of getOrderUsecase
func NewGetOrderUsecase(orderReadRepo repository.OrderReadRepository) GetOrderUsecase {
	return &getOrderUsecase{
		orderReadRepo: orderReadRepo,
	}
}

// Execute retrieves an order by ID
func (uc *getOrderUsecase) Execute(ctx context.Context, orderID string) (*entity.Order, error) {
	return uc.orderReadRepo.GetByID(ctx, orderID)
}

// internal/order_service/usecase/query/list_orders_usecase.go
