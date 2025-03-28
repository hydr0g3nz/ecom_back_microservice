package usecase

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderUseCase is the main use case for order operations
type OrderUseCase struct {
	orderRepo    repository.OrderRepository
	eventService *service.EventService
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(orderRepo repository.OrderRepository, eventService *service.EventService) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:    orderRepo,
		eventService: eventService,
	}
}

// CreateOrder creates a new order
func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID string, items []entity.OrderItem, shippingAddress string) (*entity.Order, error) {
	// Validate input
	if userID == "" {
		return nil, entity.ErrInvalidUserID
	}
	
	if len(items) == 0 {
		return nil, entity.ErrEmptyOrderItems
	}
	
	// Create new order
	order := entity.NewOrder(userID, items, shippingAddress)
	order.ID = uuid.New().String()
	
	// Save to repository
	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	
	// Publish event
	if err := uc.eventService.PublishOrderCreated(ctx, order); err != nil {
		// Log error but don't fail the operation
		// TODO: implement proper logging
	}
	
	return order, nil
}

// GetOrderByID gets an order by ID
func (uc *OrderUseCase) GetOrderByID(ctx context.Context, id string) (*entity.Order, error) {
	return uc.orderRepo.GetByID(ctx, id)
}

// GetOrdersByUserID gets all orders for a user
func (uc *OrderUseCase) GetOrdersByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	return uc.orderRepo.GetByUserID(ctx, userID)
}

// UpdateOrderStatus updates the status of an order
func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id string, status valueobject.OrderStatus) error {
	// Validate status
	if !status.IsValid() {
		return entity.ErrInvalidOrderStatus
	}
	
	// Get current order
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Save old status for event
	oldStatus := string(order.Status)
	
	// Update status
	if err := order.UpdateStatus(status); err != nil {
		return err
	}
	
	// Update in repository
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return err
	}
	
	// Publish event
	if err := uc.eventService.PublishOrderStatusChanged(ctx, order, oldStatus); err != nil {
		// Log error but don't fail the operation
		// TODO: implement proper logging
	}
	
	return nil
}

// AddPaymentToOrder adds payment information to an order
func (uc *OrderUseCase) AddPaymentToOrder(ctx context.Context, orderID, paymentID string) error {
	// Get current order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	
	// Add payment ID
	order.AddPaymentID(paymentID)
	
	// Update order status to paid
	if err := order.UpdateStatus(valueobject.OrderStatusPaid); err != nil {
		return err
	}
	
	// Update in repository
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return err
	}
	
	return nil
}

// CancelOrder cancels an order
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
	return uc.UpdateOrderStatus(ctx, id, valueobject.OrderStatusCancelled)
}

// ListOrdersByStatus lists orders by status
func (uc *OrderUseCase) ListOrdersByStatus(ctx context.Context, status valueobject.OrderStatus) ([]*entity.Order, error) {
	if !status.IsValid() {
		return nil, entity.ErrInvalidOrderStatus
	}
	
	return uc.orderRepo.ListByStatus(ctx, status)
}

// GetOrdersPaginated gets orders with pagination
func (uc *OrderUseCase) GetOrdersPaginated(ctx context.Context, page, pageSize int) ([]*entity.Order, int, error) {
	if page < 1 {
		page = 1
	}
	
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	return uc.orderRepo.GetOrdersPaginated(ctx, page, pageSize)
}
