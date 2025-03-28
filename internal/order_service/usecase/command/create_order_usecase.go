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

// CreateOrderUsecase defines the interface for creating an order
type CreateOrderUsecase interface {
	Execute(ctx context.Context, input dto.CreateOrderInput) (dto.OrderDTO, error)
}

// createOrderUsecase implements the CreateOrderUsecase interface
type createOrderUsecase struct {
	unitOfWork     repository.UnitOfWork
	idGenerator    valueobject.IDGenerator
	timeProvider   valueobject.TimeProvider
	eventPublisher event.Publisher
}

// NewCreateOrderUsecase creates a new instance of createOrderUsecase
func NewCreateOrderUsecase(
	unitOfWork repository.UnitOfWork,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
	eventPublisher event.Publisher,
) CreateOrderUsecase {
	return &createOrderUsecase{
		unitOfWork:     unitOfWork,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		eventPublisher: eventPublisher,
	}
}

// Execute creates a new order
func (uc *createOrderUsecase) Execute(ctx context.Context, input dto.CreateOrderInput) (dto.OrderDTO, error) {
	var orderDTO dto.OrderDTO
	
	// Validate input
	if input.UserID == "" {
		return orderDTO, errors.New("user ID is required")
	}

	if len(input.Items) == 0 {
		return orderDTO, errors.New("order must have at least one item")
	}

	// Run in transaction
	err := uc.unitOfWork.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Convert input to domain entity
		order, err := mapper.CreateOrderFromDTO(input, uc.idGenerator, uc.timeProvider)
		if err != nil {
			return err
		}

		// Get repositories
		orderRepo := uc.unitOfWork.GetOrderRepository(txCtx)
		orderEventRepo := uc.unitOfWork.GetOrderEventRepository(txCtx)

		// Create the order
		createdOrder, err := orderRepo.Create(txCtx, order)
		if err != nil {
			return err
		}

		// Create order created event
		eventData := entity.OrderCreatedData{
			Order: *createdOrder,
		}

		eventDataBytes, err := json.Marshal(eventData)
		if err != nil {
			return err
		}

		event, err := entity.NewOrderEvent(
			uc.idGenerator.NewID(),
			createdOrder.ID,
			entity.EventOrderCreated,
			eventDataBytes,
			1,
			uc.timeProvider.Now(),
			createdOrder.UserID,
		)
		if err != nil {
			return err
		}

		// Save the event
		if err := orderEventRepo.SaveEvent(txCtx, event); err != nil {
			return err
		}

		// Convert to DTO for response
		orderDTO = mapper.ToOrderDTO(createdOrder)

		// Publish event (outside the transaction)
		domainEvent := entity.NewDomainEvent(event, eventData)
		go func() {
			_ = uc.eventPublisher.Publish(context.Background(), domainEvent)
		}()

		return nil
	})

	if err != nil {
		return dto.OrderDTO{}, err
	}

	return orderDTO, nil
}
