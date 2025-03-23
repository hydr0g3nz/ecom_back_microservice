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

// OrderUsecase defines the interface for order operations
type OrderUsecase interface {
	// Command operations
	CreateOrder(ctx context.Context, input CreateOrderInput) (*entity.Order, error)
	UpdateOrder(ctx context.Context, id string, input UpdateOrderInput) (*entity.Order, error)
	CancelOrder(ctx context.Context, id string, reason string) error

	// Query operations
	GetOrder(ctx context.Context, id string) (*entity.Order, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error)
	ListByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error)
	Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error)
}

// CreateOrderInput contains the data needed to create an order
type CreateOrderInput struct {
	UserID          string             `json:"user_id"`
	Items           []entity.OrderItem `json:"items"`
	ShippingAddress entity.Address     `json:"shipping_address"`
	BillingAddress  entity.Address     `json:"billing_address"`
	Notes           string             `json:"notes"`
	PromotionCodes  []string           `json:"promotion_codes"`
}

// UpdateOrderInput contains the data needed to update an order
type UpdateOrderInput struct {
	Notes           *string         `json:"notes"`
	ShippingAddress *entity.Address `json:"shipping_address"`
	BillingAddress  *entity.Address `json:"billing_address"`
}

// orderUsecase implements the OrderUsecase interface
type orderUsecase struct {
	orderRepo      repository.OrderRepository
	orderReadRepo  repository.OrderReadRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
	errBuilder     *utils.ErrorBuilder
}

// NewOrderUsecase creates a new instance of orderUsecase
func NewOrderUsecase(
	orderRepo repository.OrderRepository,
	orderReadRepo repository.OrderReadRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) OrderUsecase {
	return &orderUsecase{
		orderRepo:      orderRepo,
		orderReadRepo:  orderReadRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
		errBuilder:     utils.NewErrorBuilder("OrderUsecase"),
	}
}

// CreateOrder creates a new order
func (uc *orderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*entity.Order, error) {
	// Validate input
	if input.UserID == "" {
		return nil, uc.errBuilder.Err(entity.ErrInvalidOrderData)
	}

	if len(input.Items) == 0 {
		return nil, uc.errBuilder.Err(entity.ErrInvalidOrderData)
	}

	// Create new order entity
	order := &entity.Order{
		ID:              uuid.New().String(),
		UserID:          input.UserID,
		Items:           input.Items,
		ShippingAddress: input.ShippingAddress,
		BillingAddress:  input.BillingAddress,
		Notes:           input.Notes,
		PromotionCodes:  input.PromotionCodes,
		Status:          valueobject.Created,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Version:         1,
	}

	// Calculate item totals and order total
	for i := range order.Items {
		if order.Items[i].ID == "" {
			order.Items[i].ID = uuid.New().String()
		}
		order.Items[i].TotalPrice = order.Items[i].Price * float64(order.Items[i].Quantity)
	}

	order.TotalAmount = order.CalculateTotal()

	// Create the order in the repository
	createdOrder, err := uc.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Create and store the order creation event
	eventData := entity.OrderCreatedData{
		Order: *createdOrder,
	}

	eventDataBytes, err := json.Marshal(eventData)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   createdOrder.ID,
		Type:      entity.EventOrderCreated,
		Data:      eventDataBytes,
		Version:   1,
		Timestamp: time.Now(),
		UserID:    input.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, event)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderCreated(ctx, createdOrder)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return createdOrder, nil
}

// GetOrder retrieves an order by ID
func (uc *orderUsecase) GetOrder(ctx context.Context, id string) (*entity.Order, error) {
	order, err := uc.orderReadRepo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return order, nil
}

// UpdateOrder updates an existing order
func (uc *orderUsecase) UpdateOrder(ctx context.Context, id string, input UpdateOrderInput) (*entity.Order, error) {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Check if order can be updated (not in a terminal state)
	if order.Status == valueobject.Cancelled || order.Status == valueobject.Completed {
		return nil, uc.errBuilder.Err(entity.ErrInvalidOrderStatus)
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
		return nil, uc.errBuilder.Err(err)
	}

	// Create and store the order update event
	eventDataBytes, err := json.Marshal(updatedOrder)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
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
		return nil, uc.errBuilder.Err(err)
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderUpdated(ctx, updatedOrder)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return updatedOrder, nil
}

// CancelOrder cancels an order
func (uc *orderUsecase) CancelOrder(ctx context.Context, id string, reason string) error {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	// Check if order can be cancelled
	if !order.CanCancel() {
		return uc.errBuilder.Err(entity.ErrInvalidOrderStatus)
	}

	// Create status change data
	statusData := entity.StatusChangedData{
		PreviousStatus: order.Status,
		NewStatus:      valueobject.Cancelled,
		Reason:         reason,
	}

	// Update order status
	err = uc.orderRepo.UpdateStatus(ctx, id, valueobject.Cancelled)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	// Get updated order
	updatedOrder, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	// Create and store the order cancellation events
	// First, status change event
	statusDataBytes, err := json.Marshal(statusData)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	statusEvent := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   id,
		Type:      entity.EventStatusChanged,
		Data:      statusDataBytes,
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, statusEvent)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	// Then, order cancelled event
	cancellationEvent := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   id,
		Type:      entity.EventOrderCancelled,
		Data:      json.RawMessage(`{"reason":"` + reason + `"}`),
		Version:   updatedOrder.Version,
		Timestamp: time.Now(),
		UserID:    updatedOrder.UserID,
	}

	err = uc.orderEventRepo.SaveEvent(ctx, cancellationEvent)
	if err != nil {
		return uc.errBuilder.Err(err)
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderCancelled(ctx, updatedOrder, reason)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return nil
}

// ListByUser retrieves orders for a specific user
func (uc *orderUsecase) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error) {
	orders, total, err := uc.orderReadRepo.GetByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, uc.errBuilder.Err(err)
	}
	return orders, total, nil
}

// ListByStatus retrieves orders with a specific status
func (uc *orderUsecase) ListByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error) {
	orders, total, err := uc.orderReadRepo.FindByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, uc.errBuilder.Err(err)
	}
	return orders, total, nil
}

// Search searches for orders based on various criteria
func (uc *orderUsecase) Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error) {
	orders, total, err := uc.orderReadRepo.Search(ctx, criteria, page, pageSize)
	if err != nil {
		return nil, 0, uc.errBuilder.Err(err)
	}
	return orders, total, nil
}
