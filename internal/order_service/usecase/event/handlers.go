package event

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
)

// InventoryEventHandler handles events related to inventory
type InventoryEventHandler interface {
	// HandleInventoryReserved handles the event when inventory is reserved
	HandleInventoryReserved(ctx context.Context, orderID string, success bool, reason string) error
	
	// HandleInventoryReleased handles the event when inventory is released
	HandleInventoryReleased(ctx context.Context, orderID string, success bool, reason string) error
}

// PaymentEventHandler handles events related to payments
type PaymentEventHandler interface {
	// HandlePaymentAuthorized handles the event when a payment is authorized
	HandlePaymentAuthorized(ctx context.Context, orderID string, paymentID string, success bool, reason string) error
	
	// HandlePaymentCaptured handles the event when a payment is captured
	HandlePaymentCaptured(ctx context.Context, orderID string, paymentID string, success bool, reason string) error
	
	// HandlePaymentFailed handles the event when a payment fails
	HandlePaymentFailed(ctx context.Context, orderID string, paymentID string, reason string) error
}

// inventoryEventHandler implements the InventoryEventHandler interface
type inventoryEventHandler struct {
	orderRepo      repository.OrderRepository
	orderEventRepo repository.OrderEventRepository
	orderUsecase   usecase.OrderUsecase
}

// NewInventoryEventHandler creates a new inventory event handler
func NewInventoryEventHandler(
	orderRepo repository.OrderRepository,
	orderEventRepo repository.OrderEventRepository,
	orderUsecase usecase.OrderUsecase,
) *inventoryEventHandler {
	return &inventoryEventHandler{
		orderRepo:      orderRepo,
		orderEventRepo: orderEventRepo,
		orderUsecase:   orderUsecase,
	}
}

// paymentEventHandler implements the PaymentEventHandler interface
type paymentEventHandler struct {
	orderRepo      repository.OrderRepository
	paymentRepo    repository.PaymentRepository
	orderEventRepo repository.OrderEventRepository
	orderUsecase   usecase.OrderUsecase
	paymentUsecase usecase.PaymentUsecase
}

// NewPaymentEventHandler creates a new payment event handler
func NewPaymentEventHandler(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	orderEventRepo repository.OrderEventRepository,
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
) *paymentEventHandler {
	return &paymentEventHandler{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		orderEventRepo: orderEventRepo,
		orderUsecase:   orderUsecase,
		paymentUsecase: paymentUsecase,
	}
}