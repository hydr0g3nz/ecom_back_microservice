package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
)

// PaymentEventHandler handles events related to payments
type PaymentEventHandler struct {
	orderRepo             repository.OrderRepository
	paymentRepo           repository.PaymentRepository
	orderEventRepo        repository.OrderEventRepository
	cancelOrderUsecase    command.CancelOrderUsecase
	processPaymentUsecase command.ProcessPaymentUsecase
}

// NewPaymentEventHandler creates a new instance of PaymentEventHandler
func NewPaymentEventHandler(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	orderEventRepo repository.OrderEventRepository,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
) *PaymentEventHandler {
	return &PaymentEventHandler{
		orderRepo:             orderRepo,
		paymentRepo:           paymentRepo,
		orderEventRepo:        orderEventRepo,
		cancelOrderUsecase:    cancelOrderUsecase,
		processPaymentUsecase: processPaymentUsecase,
	}
}

// HandlePaymentSuccess handles a payment success event
func (h *PaymentEventHandler) HandlePaymentSuccess(ctx context.Context, eventData []byte) error {
	// Parse the event data
	var event struct {
		OrderID       string  `json:"order_id"`
		PaymentID     string  `json:"payment_id"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		Method        string  `json:"method"`
		TransactionID string  `json:"transaction_id"`
	}

	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("error unmarshaling payment success event: %w", err)
	}

	// Process the payment
	input := command.ProcessPaymentInput{
		OrderID:         event.OrderID,
		Amount:          event.Amount,
		Currency:        event.Currency,
		Method:          event.Method,
		TransactionID:   event.TransactionID,
		GatewayResponse: "", // This would come from the payment service
	}

	_, err := h.processPaymentUsecase.Execute(ctx, input)
	return err
}

// HandlePaymentFailure handles a payment failure event
func (h *PaymentEventHandler) HandlePaymentFailure(ctx context.Context, eventData []byte) error {
	// Parse the event data
	var event struct {
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}

	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("error unmarshaling payment failure event: %w", err)
	}

	// Get the order
	order, err := h.orderRepo.GetByID(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("error getting order %s: %w", event.OrderID, err)
	}

	// Check if order can be updated
	if order.Status != valueobject.PaymentPending {
		log.Printf("Order %s payment failure ignored, order in status %s", order.ID, order.Status)
		return nil
	}

	// Update order status
	err = h.orderRepo.UpdateStatus(ctx, order.ID, valueobject.PaymentFailed)
	if err != nil {
		return err
	}

	// Create a failed payment record
	now := time.Now()
	payment := &entity.Payment{
		ID:              uuid.New().String(),
		OrderID:         order.ID,
		Amount:          order.TotalAmount,
		Currency:        "USD", // Default currency
		Method:          "unknown",
		Status:          valueobject.PaymentStatusFailed,
		TransactionID:   "",
		GatewayResponse: event.Reason,
		CreatedAt:       now,
		UpdatedAt:       now,
		FailedAt:        &now,
	}

	_, err = h.paymentRepo.Create(ctx, payment)
	if err != nil {
		log.Printf("Error creating failed payment record: %v", err)
		// Don't fail the operation if we can't create the payment record
	}

	// Create and store event for audit purposes
	orderEvent := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   event.OrderID,
		Type:      entity.EventPaymentProcessed,
		Data:      eventData,
		Version:   order.Version + 1,
		Timestamp: now,
		UserID:    order.UserID,
	}

	err = h.orderEventRepo.SaveEvent(ctx, orderEvent)
	if err != nil {
		log.Printf("Error saving payment failure event: %v", err)
		// Don't fail the operation if we can't save the event
	}

	return nil
}
