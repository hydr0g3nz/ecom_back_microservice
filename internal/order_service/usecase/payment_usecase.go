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

// paymentUsecase implements the PaymentUsecase interface
type paymentUsecase struct {
	orderRepo      repository.OrderRepository
	paymentRepo    repository.PaymentRepository
	orderEventRepo repository.OrderEventRepository
	eventPublisher publisher.OrderEventPublisher
}

// NewPaymentUsecase creates a new payment usecase
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
	}
}

// ProcessPayment processes a payment for an order
func (u *paymentUsecase) ProcessPayment(ctx context.Context, orderID string, payment *entity.Payment) error {
	// Get current order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if payment can be processed
	if !order.CanProcessPayment() {
		return entity.ErrOrderCannotProcessPayment
	}

	// Set payment ID if not provided
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}

	// Set timestamps and order ID
	now := time.Now()
	payment.OrderID = orderID
	payment.CreatedAt = now
	payment.UpdatedAt = now

	// Validate payment
	if err := payment.Validate(); err != nil {
		return fmt.Errorf("invalid payment: %w", err)
	}

	// Set initial status
	payment.Status = entity.PaymentStatusPending

	// Process payment (in a real system, this would integrate with a payment gateway)
	// Simulating successful payment for now
	payment.Status = entity.PaymentStatusCompleted
	payment.ProcessedAt = &now

	// Save payment to repository
	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// Update order status
	order.Status = entity.OrderStatusPaid
	order.UpdatedAt = now
	order.PaymentID = payment.ID

	// Save updated order
	if err := u.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// Create order payment processed event
	event := &entity.OrderEvent{
		ID:        uuid.New().String(),
		OrderID:   order.ID,
		Type:      entity.OrderEventTypePaymentProcessed,
		Data:      payment,
		CreatedAt: now,
	}

	// Save event
	if err := u.orderEventRepo.SaveEvent(ctx, event); err != nil {
		// Log error but don't fail the payment processing
		fmt.Printf("Failed to save order event: %v\n", err)
	}

	// Publish event
	if err := u.eventPublisher.PublishPaymentProcessed(ctx, order, payment); err != nil {
		// Log error but don't fail the payment processing
		fmt.Printf("Failed to publish payment event: %v\n", err)
	}

	return nil
}

// RefundPayment processes a refund for an order
func (u *paymentUsecase) RefundPayment(ctx context.Context, orderID string, amount float64, reason string) error {
	// Get current order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if refund can be processed
	if !order.CanRefund() {
		return entity.ErrOrderCannotBeRefunded
	}

	// Get payment
	payment, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Check if payment is in a refundable state
	if !payment.CanRefund() {
		return entity.ErrPaymentCannotBeRefunded
	}

	now := time.Now()

	// Create refund
	refund := &entity.Payment{
		ID:          uuid.New().String(),
		OrderID:     orderID,
		ParentID:    payment.ID,
		Type:        entity.PaymentTypeRefund,
		Amount:      amount,
		Currency:    payment.Currency,
		Method:      payment.Method,
		Status:      entity.PaymentStatusCompleted,
		Description: reason,
		CreatedAt:   now,
		UpdatedAt:   now,
		ProcessedAt: &now,
	}

	// Save refund
	if err := u.paymentRepo.Create(ctx, refund); err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	// Update payment
	payment.RefundedAmount += amount
	payment.UpdatedAt = now
	if payment.RefundedAmount >= payment.Amount {
		payment.Status = entity.PaymentStatusRefunded
	} else {
		payment.Status = entity.PaymentStatusPartiallyRefunded
	}

	// Save updated payment
	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Update order status if fully refunded
	if payment.Status == entity.PaymentStatusRefunded {
		order.Status = entity.OrderStatusRefunded
		order.UpdatedAt = now

		// Save updated order
		if err := u.orderRepo.Update(ctx, order); err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
	}

	// Create refund event
	event := &entity.OrderEvent{
		ID:      uuid.New().String(),
		OrderID: order.ID,
		Type:    entity.OrderEventTypeRefundProcessed,
		Data: map[string]interface{}{
			"refund_id":     refund.ID,
			"payment_id":    payment.ID,
			"amount":        amount,
			"reason":        reason,
			"fully_refunded": payment.Status == entity.PaymentStatusRefunded,
		},
		CreatedAt: now,
	}

	// Save event
	if err := u.orderEventRepo.SaveEvent(ctx, event); err != nil {
		// Log error but don't fail the refund
		fmt.Printf("Failed to save order event: %v\n", err)
	}

	// Publish event
	if err := u.eventPublisher.PublishRefundProcessed(ctx, order, refund, reason); err != nil {
		// Log error but don't fail the refund
		fmt.Printf("Failed to publish refund event: %v\n", err)
	}

	return nil
}