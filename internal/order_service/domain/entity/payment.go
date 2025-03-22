package entity

import (
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// Payment represents a payment for an order
type Payment struct {
	ID              valueobject.ID                `json:"id"`
	OrderID         valueobject.ID                `json:"order_id"`
	Amount          float64                       `json:"amount"`
	Currency        string                        `json:"currency"`
	Method          string                        `json:"method"` // credit_card, paypal, etc.
	Status          valueobject.PaymentStatus     `json:"status"`
	TransactionID   string                        `json:"transaction_id"`
	GatewayResponse string                        `json:"gateway_response"`
	CreatedAt       valueobject.Timestamp         `json:"created_at"`
	UpdatedAt       valueobject.Timestamp         `json:"updated_at"`
	CompletedAt     *valueobject.Timestamp        `json:"completed_at"`
	FailedAt        *valueobject.Timestamp        `json:"failed_at"`
}

// ValidatePayment validates the payment
func ValidatePayment(payment Payment) error {
	if payment.OrderID.String() == "" {
		return errors.New("order ID is required")
	}
	if payment.Amount <= 0 {
		return errors.New("payment amount must be positive")
	}
	if payment.Currency == "" {
		return errors.New("currency is required")
	}
	if payment.Method == "" {
		return errors.New("payment method is required")
	}
	if !payment.Status.IsValid() {
		return errors.New("invalid payment status")
	}
	return nil
}

// NewPayment creates a new payment
func NewPayment(
	id valueobject.ID,
	orderID valueobject.ID,
	amount float64,
	currency string,
	method string,
	status valueobject.PaymentStatus,
	transactionID string,
	gatewayResponse string,
	timeProvider valueobject.TimeProvider,
) (*Payment, error) {
	payment := &Payment{
		ID:              id,
		OrderID:         orderID,
		Amount:          amount,
		Currency:        currency,
		Method:          method,
		Status:          status,
		TransactionID:   transactionID,
		GatewayResponse: gatewayResponse,
		CreatedAt:       timeProvider.Now(),
		UpdatedAt:       timeProvider.Now(),
	}

	// Set timestamps based on status
	if status == valueobject.PaymentStatusCompleted {
		completedAt := timeProvider.Now()
		payment.CompletedAt = &completedAt
	} else if status == valueobject.PaymentStatusFailed {
		failedAt := timeProvider.Now()
		payment.FailedAt = &failedAt
	}

	// Validate the payment
	if err := ValidatePayment(*payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// UpdateStatus updates the payment status and related timestamps
func (p *Payment) UpdateStatus(status valueobject.PaymentStatus, timeProvider valueobject.TimeProvider) error {
	if !status.IsValid() {
		return errors.New("invalid payment status")
	}

	p.Status = status
	p.UpdatedAt = timeProvider.Now()

	// Update timestamps based on status
	if status == valueobject.PaymentStatusCompleted {
		completedAt := timeProvider.Now()
		p.CompletedAt = &completedAt
	} else if status == valueobject.PaymentStatusFailed {
		failedAt := timeProvider.Now()
		p.FailedAt = &failedAt
	}

	return nil
}
