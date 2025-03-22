package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/event/publisher"
)

// shippingUsecase implements the ShippingUsecase interface
type shippingUsecase struct {
	orderRepo      repository.OrderRepository
	shippingRepo   repository.ShippingRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewShippingUsecase creates a new shipping usecase
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
	}
}

// UpdateShipping updates shipping information for an order
func (u *shippingUsecase) UpdateShipping(ctx context.Context, orderID string, shipping *entity.Shipping) error {
	// Get current order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if shipping can be updated
	if !order.CanUpdateShipping() {
		return entity.ErrOrderCannotUpdateShipping
	}

	// Set shipping ID if not provided
	if shipping.ID == "" {
		shipping.ID = uuid.New().String()
	}

	// Set timestamps and order ID
	now := time.Now()
	shipping.OrderID = orderID
	shipping.UpdatedAt = now

	// Check if this is a new shipping record
	currentShipping, err := u.shippingRepo.GetByOrderID(ctx, orderID)
	if err != nil && err != entity.ErrShippingNotFound {
		return fmt.Errorf("failed to check existing shipping: %w", err)
	}

	var isNew bool
	if err == entity.ErrShippingNotFound {
		// New shipping record
		isNew = true
		shipping.CreatedAt = now
	} else {
		// Update existing record
		shipping.ID = currentShipping.ID
		shipping.CreatedAt = currentShipping.CreatedAt
		
		// Don't allow changing carrier or tracking number if already shipped
		if currentShipping.Status == entity.ShippingStatusShipped {
			shipping.Carrier = currentShipping.Carrier
			shipping.TrackingNumber = currentShipping.TrackingNumber
		}
	}

	// Validate shipping
	if err := shipping.Validate(); err != nil {
		return fmt.Errorf("invalid shipping: %w", err)
	}

	// Handle status transitions
	if shipping.Status == entity.ShippingStatusShipped && 
		(isNew || currentShipping.Status != entity.ShippingStatusShipped) {
		// If changing to shipped status, set shipped date
		shipping.ShippedAt = &now
		
		// Update order status
		order.Status = entity.OrderStatusShipped
		order.UpdatedAt = now
		
		if err := u.orderRepo.Update(ctx, order); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	} else if shipping.Status == entity.ShippingStatusDelivered && 
		(isNew || currentShipping.Status != entity.ShippingStatusDelivered) {
		// If changing to delivered status, set delivered date
		shipping.DeliveredAt = &now
		
		// Update order status
		order.Status = entity.OrderStatusDelivered
		order.UpdatedAt = now
		
		if err := u.orderRepo.Update(ctx, order); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}
	}

	// Save shipping information
	if isNew {
		if err := u.shippingRepo.Create(ctx, shipping); err != nil {
			return fmt.Errorf("failed to create shipping: %w", err)
		}
	} else {
		if err := u.shippingRepo.Update(ctx, shipping); err != nil {
			return fmt.Errorf("failed to update shipping: %w", err)
		}
	}

	// Create shipping updated event
	eventType := entity.OrderEventTypeShippingCreated
	if !isNew {
		eventType = entity.OrderEventTypeShippingUpdated
	}
	
	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   order.ID,
		Type:      eventType,
		Data:      shipping,
		CreatedAt: now,
	}

	// Save event
	if err := u.orderEventRepo.SaveEvent(ctx, event); err != nil {
		// Log error but don't fail the shipping update
		fmt.Printf("Failed to save order event: %v\n", err)
	}

	// Publish event
	if err := u.eventPublisher.PublishShippingUpdated(ctx, order, shipping); err != nil {
		// Log error but don't fail the shipping update
		fmt.Printf("Failed to publish shipping event: %v\n", err)
	}

	return nil
}

// TrackShipment gets tracking information for an order
func (u *shippingUsecase) TrackShipment(ctx context.Context, orderID string) (*entity.Shipping, error) {
	// Get shipping information
	shipping, err := u.shippingRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping: %w", err)
	}

	// In a real system, this would integrate with shipping carriers' APIs to get up-to-date tracking information
	// For now, just return the stored shipping info

	return shipping, nil
}