package mapper

import (
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
)

// ToPaymentDTO converts a payment entity to a payment DTO
func ToPaymentDTO(payment *entity.Payment) dto.PaymentDTO {
	paymentDTO := dto.PaymentDTO{
		ID:              payment.ID.String(),
		OrderID:         payment.OrderID.String(),
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Method:          payment.Method,
		Status:          payment.Status.String(),
		TransactionID:   payment.TransactionID,
		GatewayResponse: payment.GatewayResponse,
		CreatedAt:       payment.CreatedAt.String(),
		UpdatedAt:       payment.UpdatedAt.String(),
	}

	if payment.CompletedAt != nil {
		completedAt := payment.CompletedAt.String()
		paymentDTO.CompletedAt = &completedAt
	}

	if payment.FailedAt != nil {
		failedAt := payment.FailedAt.String()
		paymentDTO.FailedAt = &failedAt
	}

	return paymentDTO
}

// CreatePaymentFromDTO creates a payment entity from a process payment input DTO
func CreatePaymentFromDTO(
	input dto.ProcessPaymentInput,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
) (*entity.Payment, error) {
	// Default status for a new payment
	status := valueobject.PaymentStatusCompleted

	// Create the payment
	return entity.NewPayment(
		idGenerator.NewID(),
		valueobject.ID(input.OrderID),
		input.Amount,
		input.Currency,
		input.Method,
		status,
		input.TransactionID,
		input.GatewayResponse,
		timeProvider,
	)
}
