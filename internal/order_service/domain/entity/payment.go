package entity

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// Payment represents a payment for an order
type Payment struct {
	ID              string                    `json:"id"`
	OrderID         string                    `json:"order_id"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	Method          string                    `json:"method"` // credit_card, paypal, etc.
	Status          valueobject.PaymentStatus `json:"status"`
	TransactionID   string                    `json:"transaction_id"`
	GatewayResponse string                    `json:"gateway_response"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	CompletedAt     *time.Time                `json:"completed_at"`
	FailedAt        *time.Time                `json:"failed_at"`
}

// UpdateStatus updates the payment status and related timestamps
func (p *Payment) UpdateStatus(status valueobject.PaymentStatus) {
	p.Status = status
	p.UpdatedAt = time.Now()

	now := time.Now()
	if status == valueobject.PaymentStatusCompleted {
		p.CompletedAt = &now
	} else if status == valueobject.PaymentStatusFailed {
		p.FailedAt = &now
	}
}
