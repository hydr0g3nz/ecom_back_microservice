package usecase

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

// OrderHistoryUsecase defines the interface for retrieving order history
type OrderHistoryUsecase interface {
	GetEvents(ctx context.Context, orderID string) ([]*entity.OrderEvent, error)
	RebuildFromEvents(ctx context.Context, orderID string) (*entity.Order, error)
}

// orderHistoryUsecase implements the OrderHistoryUsecase interface
type orderHistoryUsecase struct {
	orderEventRepo repository.OrderEventRepository
	errBuilder     *utils.ErrorBuilder
}

// NewOrderHistoryUsecase creates a new instance of orderHistoryUsecase
func NewOrderHistoryUsecase(orderEventRepo repository.OrderEventRepository) OrderHistoryUsecase {
	return &orderHistoryUsecase{
		orderEventRepo: orderEventRepo,
		errBuilder:     utils.NewErrorBuilder("OrderHistoryUsecase"),
	}
}

// GetEvents retrieves all events for an order
func (uc *orderHistoryUsecase) GetEvents(ctx context.Context, orderID string) ([]*entity.OrderEvent, error) {
	events, err := uc.orderEventRepo.GetEventsByOrderID(ctx, orderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return events, nil
}

// RebuildFromEvents reconstructs an order from its event history
func (uc *orderHistoryUsecase) RebuildFromEvents(ctx context.Context, orderID string) (*entity.Order, error) {
	order, err := uc.orderEventRepo.RebuildOrderFromEvents(ctx, orderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return order, nil
}
