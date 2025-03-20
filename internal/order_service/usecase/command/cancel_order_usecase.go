package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
)

// CancelOrderUsecase defines the interface for cancelling an order
type CancelOrderUsecase interface {
	Execute(ctx context.Context, orderID string, reason string) error
}

// cancelOrderUsecase implements the CancelOrderUsecase interface
type cancelOrderUsecase struct {
	orderRepo      repository.OrderRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewCancelOrderUsecase creates a new instance of cancelOrderUsecase
func NewCancelOrderUsecase(
	orderRepo repository.OrderRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) CancelOrderUsecase {
	return &cancelOrderUsecase{
		orderRepo:      orderRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute cancels an order
func (uc *cancelOrderUsecase) Execute(ctx context.Context, orderID string, reason string) error {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Check if order can be cancelled
	if !order.CanCancel() {
		return entity.ErrInvalidOrderStatus
	}

	// Create status change data
	statusData := entity.StatusChangedData{
		PreviousStatus: order.Status,
		NewStatus:      valueobject.Cancelled,
		Reason:         reason,
	}

	// Update order status
	err = uc.orderRepo.UpdateStatus(ctx, orderID, valueobject.Cancelled)
	if err != nil {
		return err
	}

	// Get updated order
	updatedOrder, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Create and store the order cancellation events
	// First, status change event
	statusDataBytes, err := json.Marshal(statusData)
	if err != nil {
		return err
	}

	statusEvent := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Type:      entity.EventStatusChanged,
		Data:      statusDataBytes,
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, statusEvent)
	if err != nil {
		return err
	}

	// Then, order cancelled event
	cancellationEvent := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Type:      entity.EventOrderCancelled,
		Data:      json.RawMessage(`{"reason":"` + reason + `"}`),
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, cancellationEvent)
	if err != nil {
		return err
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderCancelled(ctx, updatedOrder, reason)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return nil
}
