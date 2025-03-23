package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

// ShippingUsecase defines the interface for shipping operations
type ShippingUsecase interface {
	// Command operations
	UpdateShipping(ctx context.Context, input UpdateShippingInput) (*entity.Shipping, error)

	// Query operations
	GetShippingByID(ctx context.Context, id string) (*entity.Shipping, error)
	GetShippingByOrderID(ctx context.Context, orderID string) (*entity.Shipping, error)
}

// UpdateShippingInput contains the data needed to update shipping information
type UpdateShippingInput struct {
	OrderID           string                     `json:"order_id"`
	Carrier           string                     `json:"carrier"`
	TrackingNumber    string                     `json:"tracking_number"`
	Status            valueobject.ShippingStatus `json:"status"`
	EstimatedDelivery *time.Time                 `json:"estimated_delivery"`
	ShippingMethod    string                     `json:"shipping_method"`
	ShippingCost      float64                    `json:"shipping_cost"`
	Notes             string                     `json:"notes"`
}

// shippingUsecase implements the ShippingUsecase interface
type shippingUsecase struct {
	orderRepo      repository.OrderRepository
	shippingRepo   repository.ShippingRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
	errBuilder     *utils.ErrorBuilder
}

// NewShippingUsecase creates a new instance of shippingUsecase
func NewShippingUsecase(
	orderRepo repository.OrderRepository,
	shippingRepo repository.ShippingRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) ShippingUsecase {
	return &shippingUsecase{
		orderRepo:      orderRepo,
		shippingRepo:   shippingRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
		errBuilder:     utils.NewErrorBuilder("ShippingUsecase"),
	}
}

// UpdateShipping updates shipping information for an order
func (uc *shippingUsecase) UpdateShipping(ctx context.Context, input UpdateShippingInput) (*entity.Shipping, error) {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Check if order can be shipped
	if !order.CanShip() {
		return nil, uc.errBuilder.Err(entity.ErrInvalidOrderStatus)
	}

	// Check if shipping record exists
	var shipping *entity.Shipping
	var existingShipping bool

	shipping, err = uc.shippingRepo.GetByOrderID(ctx, input.OrderID)
	if err != nil {
		if err != entity.ErrShippingNotFound {
			return nil, uc.errBuilder.Err(err)
		}

		// Create new shipping record
		shipping = &entity.Shipping{
			ID:        uuid.New().String(),
			OrderID:   input.OrderID,
			Status:    valueobject.ShippingStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		existingShipping = false
	} else {
		existingShipping = true
	}

	// Update shipping information
	shipping.Carrier = input.Carrier
	shipping.TrackingNumber = input.TrackingNumber
	shipping.Status = input.Status
	shipping.EstimatedDelivery = input.EstimatedDelivery
	shipping.ShippingMethod = input.ShippingMethod
	shipping.ShippingCost = input.ShippingCost
	shipping.Notes = input.Notes
	shipping.UpdatedAt = time.Now()

	// Update shipping-specific timestamps based on status
	if input.Status == valueobject.ShippingStatusShipped {
		now := time.Now()
		shipping.ShippedAt = &now

		// Update order status to Shipped if shipping status is Shipped
		err = uc.orderRepo.UpdateStatus(ctx, input.OrderID, valueobject.Shipped)
		if err != nil {
			return nil, uc.errBuilder.Err(err)
		}
	} else if input.Status == valueobject.ShippingStatusDelivered {
		now := time.Now()
		shipping.DeliveredAt = &now

		// Update order status to Delivered if shipping status is Delivered
		err = uc.orderRepo.UpdateStatus(ctx, input.OrderID, valueobject.Delivered)
		if err != nil {
			return nil, uc.errBuilder.Err(err)
		}
	}

	// Create or update shipping record
	var updatedShipping *entity.Shipping
	if existingShipping {
		updatedShipping, err = uc.shippingRepo.Update(ctx, shipping)
	} else {
		updatedShipping, err = uc.shippingRepo.Create(ctx, shipping)
	}

	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Get updated order for event
	updatedOrder, err := uc.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Create and store the shipping updated event
	eventDataBytes, err := json.Marshal(map[string]interface{}{
		"shipping_id":        updatedShipping.ID,
		"status":             updatedShipping.Status,
		"carrier":            updatedShipping.Carrier,
		"tracking_number":    updatedShipping.TrackingNumber,
		"estimated_delivery": updatedShipping.EstimatedDelivery,
	})
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   input.OrderID,
		Type:      entity.EventShippingUpdated,
		Data:      eventDataBytes,
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, event)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Publish the event
	err = uc.eventPublisher.PublishShippingUpdated(ctx, updatedOrder, updatedShipping)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return updatedShipping, nil
}

// GetShippingByID retrieves a shipping record by ID
func (uc *shippingUsecase) GetShippingByID(ctx context.Context, id string) (*entity.Shipping, error) {
	shipping, err := uc.shippingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return shipping, nil
}

// GetShippingByOrderID retrieves shipping for an order
func (uc *shippingUsecase) GetShippingByOrderID(ctx context.Context, orderID string) (*entity.Shipping, error) {
	shipping, err := uc.shippingRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return shipping, nil
}
