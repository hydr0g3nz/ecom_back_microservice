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

// PaymentUsecase defines the interface for payment operations
type PaymentUsecase interface {
	// Command operations
	ProcessPayment(ctx context.Context, input ProcessPaymentInput) (*entity.Payment, error)

	// Query operations
	GetPaymentByID(ctx context.Context, id string) (*entity.Payment, error)
	GetPaymentsByOrderID(ctx context.Context, orderID string) ([]*entity.Payment, error)
}

// ProcessPaymentInput contains the data needed to process a payment
type ProcessPaymentInput struct {
	OrderID         string  `json:"order_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Method          string  `json:"method"`
	TransactionID   string  `json:"transaction_id"`
	GatewayResponse string  `json:"gateway_response"`
}

// paymentUsecase implements the PaymentUsecase interface
type paymentUsecase struct {
	orderRepo      repository.OrderRepository
	paymentRepo    repository.PaymentRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
	errBuilder     *utils.ErrorBuilder
}

// NewPaymentUsecase creates a new instance of paymentUsecase
func NewPaymentUsecase(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) PaymentUsecase {
	return &paymentUsecase{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
		errBuilder:     utils.NewErrorBuilder("PaymentUsecase"),
	}
}

// ProcessPayment processes a payment for an order
func (uc *paymentUsecase) ProcessPayment(ctx context.Context, input ProcessPaymentInput) (*entity.Payment, error) {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Validate payment amount
	if input.Amount != order.TotalAmount {
		return nil, uc.errBuilder.Err(entity.ErrInvalidOrderData)
	}

	// Create payment entity
	payment := &entity.Payment{
		ID:              uuid.New().String(),
		OrderID:         input.OrderID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Method:          input.Method,
		Status:          valueobject.PaymentStatusCompleted,
		TransactionID:   input.TransactionID,
		GatewayResponse: input.GatewayResponse,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set completed time
	now := time.Now()
	payment.CompletedAt = &now

	// Create the payment in the repository
	createdPayment, err := uc.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Update order status to payment completed
	err = uc.orderRepo.UpdateStatus(ctx, order.ID, valueobject.PaymentCompleted)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Get updated order
	updatedOrder, err := uc.orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	// Create and store the payment processed event
	eventDataBytes, err := json.Marshal(map[string]interface{}{
		"payment_id":     createdPayment.ID,
		"amount":         createdPayment.Amount,
		"currency":       createdPayment.Currency,
		"method":         createdPayment.Method,
		"transaction_id": createdPayment.TransactionID,
	})
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}

	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   order.ID,
		Type:      entity.EventPaymentProcessed,
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
	err = uc.eventPublisher.PublishPaymentProcessed(ctx, updatedOrder, createdPayment)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return createdPayment, nil
}

// GetPaymentByID retrieves a payment by ID
func (uc *paymentUsecase) GetPaymentByID(ctx context.Context, id string) (*entity.Payment, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return payment, nil
}

// GetPaymentsByOrderID retrieves payments for an order
func (uc *paymentUsecase) GetPaymentsByOrderID(ctx context.Context, orderID string) ([]*entity.Payment, error) {
	payments, err := uc.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, uc.errBuilder.Err(err)
	}
	return payments, nil
}
