package model

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// PaymentModel represents a payment in Cassandra
type PaymentModel struct {
	ID              gocql.UUID `json:"id"`
	OrderID         gocql.UUID `json:"order_id"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	Method          string     `json:"method"`
	Status          string     `json:"status"`
	TransactionID   string     `json:"transaction_id"`
	GatewayResponse string     `json:"gateway_response"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	FailedAt        *time.Time `json:"failed_at"`
}

// ToEntity converts a Cassandra PaymentModel to a domain Payment entity
func (pm *PaymentModel) ToEntity() (*entity.Payment, error) {
	status, err := valueobject.ParsePaymentStatus(pm.Status)
	if err != nil {
		return nil, err
	}

	return &entity.Payment{
		ID:              pm.ID.String(),
		OrderID:         pm.OrderID.String(),
		Amount:          pm.Amount,
		Currency:        pm.Currency,
		Method:          pm.Method,
		Status:          status,
		TransactionID:   pm.TransactionID,
		GatewayResponse: pm.GatewayResponse,
		CreatedAt:       pm.CreatedAt,
		UpdatedAt:       pm.UpdatedAt,
		CompletedAt:     pm.CompletedAt,
		FailedAt:        pm.FailedAt,
	}, nil
}

// FromEntity converts a domain Payment entity to a Cassandra PaymentModel
func FromPaymentEntity(payment *entity.Payment) (*PaymentModel, error) {
	id, err := gocql.ParseUUID(payment.ID)
	if err != nil {
		return nil, err
	}

	orderID, err := gocql.ParseUUID(payment.OrderID)
	if err != nil {
		return nil, err
	}

	return &PaymentModel{
		ID:              id,
		OrderID:         orderID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Method:          payment.Method,
		Status:          payment.Status.String(),
		TransactionID:   payment.TransactionID,
		GatewayResponse: payment.GatewayResponse,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
		CompletedAt:     payment.CompletedAt,
		FailedAt:        payment.FailedAt,
	}, nil
}
