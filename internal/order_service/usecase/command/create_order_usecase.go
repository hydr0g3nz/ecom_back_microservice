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

// CreateOrderUsecase defines the interface for creating an order
type CreateOrderUsecase interface {
	Execute(ctx context.Context, input CreateOrderInput) (*entity.Order, error)
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

// createOrderUsecase implements the CreateOrderUsecase interface
type createOrderUsecase struct {
	orderRepo      repository.OrderRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewCreateOrderUsecase creates a new instance of createOrderUsecase
func NewCreateOrderUsecase(
	orderRepo repository.OrderRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) CreateOrderUsecase {
	return &createOrderUsecase{
		orderRepo:      orderRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute creates a new order
func (uc *createOrderUsecase) Execute(ctx context.Context, input CreateOrderInput) (*entity.Order, error) {
	// Validate input
	if input.UserID == "" {
		return nil, entity.ErrInvalidOrderData
	}

	if len(input.Items) == 0 {
		return nil, entity.ErrInvalidOrderData
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
		return nil, err
	}

	// Create and store the order creation event
	eventData := entity.OrderCreatedData{
		Order: *createdOrder,
	}

	eventDataBytes, err := json.Marshal(eventData)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Publish the event
	err = uc.eventPublisher.PublishOrderCreated(ctx, createdOrder)
	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}
