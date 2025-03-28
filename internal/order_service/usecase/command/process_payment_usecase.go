package command

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/event"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/mapper"
)

// ProcessPaymentUsecase defines the interface for processing a payment
type ProcessPaymentUsecase interface {
	Execute(ctx context.Context, input dto.ProcessPaymentInput) (dto.PaymentDTO, error)
}

// processPaymentUsecase implements the ProcessPaymentUsecase interface
type processPaymentUsecase struct {
	unitOfWork     repository.UnitOfWork
	idGenerator    valueobject.IDGenerator
	timeProvider   valueobject.TimeProvider
	eventPublisher event.Publisher
}

// NewProcessPaymentUsecase creates a new instance of processPaymentUsecase
func NewProcessPaymentUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) ProcessPaymentUsecase {
	return &processPaymentUsecase{
		unitOfWork:     unitOfWork,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		eventPublisher: eventPublisher,
	}
}

// Execute processes a payment for an order
func (uc *processPaymentUsecase) Execute(ctx context.Context, input dto.ProcessPaymentInput) (dto.PaymentDTO, error) {
	var paymentDTO dto.PaymentDTO

	// Validate input
	if input.OrderID == "" {
		return paymentDTO, errors.New("order ID is required")
	}
	if input.Amount <= 0 {
		return paymentDTO, errors.New("payment amount must be positive")
	}
	if input.Currency == "" {
		return paymentDTO, errors.New("currency is required")
	}
	if input.Method == "" {
		return paymentDTO, errors.New("payment method is required")
	}

	// Run in transaction
	err := uc.unitOfWork.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Get repositories
		orderRepo := uc.unitOfWork.GetOrderRepository(txCtx)
		paymentRepo := uc.unitOfWork.GetPaymentRepository(txCtx)
		orderEventRepo := uc.unitOfWork.GetOrderEventRepository(txCtx)

		// Get the existing order
		order, err := orderRepo.GetByID(txCtx, input.OrderID)
		if err != nil {
			return err
		}

		// Validate payment amount against order total
		if input.Amount != order.TotalAmount {
			return errors.New("payment amount doesn't match order total")
		}

		// Create payment entity
		payment, err := mapper.CreatePaymentFromDTO(input, uc.idGenerator, uc.timeProvider)
		if err != nil {
			return err
		}

		// Create the payment
		createdPayment, err := paymentRepo.Create(txCtx, payment)
		if err != nil {
			return err
		}

		// Update order status to payment completed
		err = orderRepo.UpdateStatus(txCtx, order.ID.String(), valueobject.PaymentCompleted)
		if err != nil {
			return err
		}

		// Get the updated order
		updatedOrder, err := orderRepo.GetByID(txCtx, input.OrderID)
		if err != nil {
			return err
		}

		// Create payment processed event
		paymentData := entity.PaymentProcessedData{
			Payment: *createdPayment,
		}

		eventDataBytes, err := json.Marshal(paymentData)
		if err != nil {
			return err
		}

		event, err := entity.NewOrderEvent(
			uc.idGenerator.NewID(),
			updatedOrder.ID,
			entity.EventPaymentProcessed,
			eventDataBytes,
			updatedOrder.Version,
			uc.timeProvider.Now(),
			updatedOrder.UserID,
		)
		if err != nil {
			return err
		}

		// Save the event
		if err := orderEventRepo.SaveEvent(txCtx, event); err != nil {
			return err
		}

		// Convert to DTO for response
		paymentDTO = mapper.ToPaymentDTO(createdPayment)

		// Publish event (outside the transaction)
		domainEvent := entity.NewDomainEvent(event, paymentData)
		go func() {
			_ = uc.eventPublisher.Publish(context.Background(), domainEvent)
		}()

		return nil
	})

	if err != nil {
		return dto.PaymentDTO{}, err
	}

	return paymentDTO, nil
}
