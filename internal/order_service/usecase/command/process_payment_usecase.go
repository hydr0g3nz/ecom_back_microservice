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

// ProcessPaymentUsecase defines the interface for processing a payment
type ProcessPaymentUsecase interface {
	Execute(ctx context.Context, input ProcessPaymentInput) (*entity.Payment, error)
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

// processPaymentUsecase implements the ProcessPaymentUsecase interface
type processPaymentUsecase struct {
	orderRepo      repository.OrderRepository
	paymentRepo    repository.PaymentRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewProcessPaymentUsecase creates a new instance of processPaymentUsecase
func NewProcessPaymentUsecase(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	orderEventRepo repository.OrderEventRepository,
	eventPublisher publisher.OrderEventPublisher,
) ProcessPaymentUsecase {
	return &processPaymentUsecase{
		orderRepo:      orderRepo,
		paymentRepo:    paymentRepo,
		orderEventRepo: orderEventRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute processes a payment for an order
func (uc *processPaymentUsecase) Execute(ctx context.Context, input ProcessPaymentInput) (*entity.Payment, error) {
	// Get the existing order
	order, err := uc.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}

	// Validate payment amount
	if input.Amount != order.TotalAmount {
		return nil, entity.ErrInvalidOrderData
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
		CompletedAt:     &time.Time{},
	}

	// Set completed time
	now := time.Now()
	payment.CompletedAt = &now

	// Create the payment in the repository
	createdPayment, err := uc.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, err
	}

	// Update order status to payment completed and set payment ID
	// First update order status
	err = uc.orderRepo.UpdateStatus(ctx, order.ID, valueobject.PaymentCompleted)
	if err != nil {
		return nil, err
	}

	// Get updated order
	updatedOrder, err := uc.orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	// Publish the event
	err = uc.eventPublisher.PublishPaymentProcessed(ctx, updatedOrder, createdPayment)
	if err != nil {
		// Log the error but don't fail the operation
	}

	return createdPayment, nil
}
