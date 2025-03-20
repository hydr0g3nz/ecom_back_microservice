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

// UpdateOrderUsecase defines the interface for updating an order
type UpdateOrderUsecase interface {
	Execute(ctx context.Context, orderID string, input UpdateOrderInput) (*entity.Order, error)
}

// UpdateOrderInput contains the data needed to update an order
type UpdateOrderInput struct {
	Notes           *string         `json:"notes"`
	ShippingAddress *entity.Address `json:"shipping_address"`
	BillingAddress  *entity.Address `json:"billing_address"`
}

// updateOrderUsecase implements the UpdateOrderUsecase interface
type updateOrderUsecase struct {
	orderRepo      repository.OrderRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewUpdateOrderUsecase creates a new instance of updateOrderUsecase
func NewUpdateOrderUsecase(
	orderRepo repository.OrderRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) UpdateOrderUsecase {
	return &updateOrderUsecase{
		orderRepo:      orderRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute updates an existing order
func (uc *updateOrderUsecase) Execute(ctx context.Context, orderID string, input UpdateOrderInput) (*entity.Order, error) {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Check if order can be updated (not in a terminal state)
	if order.Status == valueobject.Cancelled || order.Status == valueobject.Completed {
		return nil, entity.ErrInvalidOrderStatus
	}

	// Apply updates
	if input.Notes != nil {
		order.Notes = *input.Notes
	}

	if input.ShippingAddress != nil {
		order.ShippingAddress = *input.ShippingAddress
	}

	if input.BillingAddress != nil {
		order.BillingAddress = *input.BillingAddress
	}

	// Update timestamp
	order.UpdatedAt = time.Now()

	// Update the order in the repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, err
	}

	// Create and store the order update event
	eventDataBytes, err := json.Marshal(updatedOrder)
	if err != nil {
		return nil, err
	}

	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   updatedOrder.ID,
		Type:      entity.EventOrderUpdated,
		Data:      eventDataBytes,
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderUpdated(ctx, updatedOrder)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return updatedOrder, nil
}
