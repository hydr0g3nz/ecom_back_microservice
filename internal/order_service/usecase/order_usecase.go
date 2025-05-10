// internal/order_service/usecase/order_usecase.go
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

type OrderUsecase interface {
	// CreateOrder creates a new order
	CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// GetOrderByID retrieves an order by ID
	GetOrderByID(ctx context.Context, id string) (*entity.Order, error)

	// GetOrdersByUserID retrieves orders for a specific user
	GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error)

	// ListOrders retrieves a list of orders with optional filtering
	ListOrders(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Order, int, error)

	// UpdateOrder updates an existing order
	UpdateOrder(ctx context.Context, id string, order entity.Order) (*entity.Order, error)

	// UpdateOrderStatus updates the status of an order
	UpdateOrderStatus(ctx context.Context, id string, status valueobject.OrderStatus, comment string) (*entity.Order, error)

	// CancelOrder cancels an order
	CancelOrder(ctx context.Context, id string, reason string) (*entity.Order, error)

	// ProcessInventoryReserved handles the event when inventory is reserved
	ProcessInventoryReserved(ctx context.Context, orderID string, success bool, message string) (*entity.Order, error)

	// ProcessPaymentCompleted handles the event when payment is completed
	ProcessPaymentCompleted(ctx context.Context, orderID string, transactionID string, success bool) (*entity.Order, error)

	// UpdateOrderPartial performs a partial update of an order
	UpdateOrderPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Order, error)
}

// orderUsecase implements the OrderUsecase interface
type orderUsecase struct {
	orderRepo  repository.OrderRepository
	eventPub   service.EventPublisherService
	errBuilder *utils.ErrorBuilder
}

// NewOrderUsecase creates a new instance of OrderUsecase
func NewOrderUsecase(
	or repository.OrderRepository,
	es service.EventPublisherService,
) OrderUsecase {
	return &orderUsecase{
		orderRepo:  or,
		eventPub:   es,
		errBuilder: utils.NewErrorBuilder("OrderUsecase"),
	}
}

// CreateOrder creates a new order
func (ou *orderUsecase) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	// Validate order data
	if err := order.ValidateOrder(); err != nil {
		return nil, ou.errBuilder.Err(err)
	}

	// Generate ID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Set initial status and timestamps
	order.Status = valueobject.OrderStatusPending
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Initialize status history
	order.StatusHistory = []entity.OrderStatusHistoryItem{
		{
			Status:    valueobject.OrderStatusPending,
			Timestamp: time.Now(),
			Comment:   "Order created",
		},
	}

	// Calculate total amount
	order.CalculateTotalAmount()

	// Create order in repository
	createdOrder, err := ou.orderRepo.Create(ctx, *order)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}
	go func() {

		// Publish order created event
		if err := ou.eventPub.PublishOrderCreated(ctx, createdOrder); err != nil {
			fmt.Println("Error publishing order created event:", err)
			// In a real system, you might want to implement retry logic or compensating actions
		}

		// Request inventory reservation
		if err := ou.eventPub.PublishReserveInventory(ctx, createdOrder); err != nil {
			fmt.Println("Error publishing reserve inventory event:", err)
			// Log error but continue
			// In a real system, you might want to implement retry logic
		}
	}()

	return createdOrder, nil
}

// GetOrderByID retrieves an order by ID
func (ou *orderUsecase) GetOrderByID(ctx context.Context, id string) (*entity.Order, error) {
	order, err := ou.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}
	return order, nil
}

// GetOrdersByUserID retrieves orders for a specific user
func (ou *orderUsecase) GetOrdersByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error) {
	offset := (page - 1) * pageSize
	orders, total, err := ou.orderRepo.GetByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, ou.errBuilder.Err(err)
	}
	return orders, total, nil
}

// ListOrders retrieves a list of orders with optional filtering
func (ou *orderUsecase) ListOrders(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*entity.Order, int, error) {
	offset := (page - 1) * pageSize
	orders, total, err := ou.orderRepo.List(ctx, offset, pageSize, filters)
	if err != nil {
		return nil, 0, ou.errBuilder.Err(err)
	}
	return orders, total, nil
}

// UpdateOrder updates an existing order
func (ou *orderUsecase) UpdateOrder(ctx context.Context, id string, order entity.Order) (*entity.Order, error) {
	// Validate order data
	if err := order.ValidateOrder(); err != nil {
		return nil, ou.errBuilder.Err(err)
	}
	// Ensure the order exists
	existingOrder, err := ou.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ou.errBuilder.Err(entity.ErrOrderNotFound)
	}

	// Preserve important fields from the existing order
	order.ID = id
	order.CreatedAt = existingOrder.CreatedAt
	order.StatusHistory = existingOrder.StatusHistory
	order.UpdatedAt = time.Now()

	// Recalculate total amount
	order.CalculateTotalAmount()

	// Update the order
	updatedOrder, err := ou.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}

	// Publish order updated event
	if err := ou.eventPub.PublishOrderUpdated(ctx, updatedOrder); err != nil {
		// Log error but continue
	}

	return updatedOrder, nil
}

// UpdateOrderStatus updates the status of an order
func (ou *orderUsecase) UpdateOrderStatus(ctx context.Context, id string, status valueobject.OrderStatus, comment string) (*entity.Order, error) {
	// Ensure the order exists
	existingOrder, err := ou.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ou.errBuilder.Err(entity.ErrOrderNotFound)
	}

	// Validate status transition
	if !existingOrder.CanTransitionToStatus(status) {
		return nil, ou.errBuilder.Err(entity.ErrInvalidStatusTransition)
	}

	// Update the order status
	updatedOrder, err := ou.orderRepo.UpdateStatus(ctx, id, status, comment)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}

	// Publish appropriate events based on the new status
	switch status {
	case valueobject.OrderStatusCancelled:
		if err := ou.eventPub.PublishOrderCancelled(ctx, updatedOrder); err != nil {
			// Log error but continue
			fmt.Println("Error publishing order cancelled event:", err)
		}
		if err := ou.eventPub.PublishReleaseInventory(ctx, updatedOrder); err != nil {
			// Log error but continue
			fmt.Println("Error publishing release inventory event:", err)
		}
	case valueobject.OrderStatusCompleted:
		if err := ou.eventPub.PublishOrderCompleted(ctx, updatedOrder); err != nil {
			// Log error but continue
			fmt.Println("Error publishing order completed event:", err)
		}
	case valueobject.OrderStatusProcessing:
		// If transitioning to processing, request payment
		if existingOrder.Status == valueobject.OrderStatusPending {
			if err := ou.eventPub.PublishPaymentRequest(ctx, updatedOrder); err != nil {
				// Log error but continue
				fmt.Println("Error publishing payment request event:", err)
			}
		}
	}
	return updatedOrder, nil
}

// CancelOrder cancels an order
func (ou *orderUsecase) CancelOrder(ctx context.Context, id string, reason string) (*entity.Order, error) {
	return ou.UpdateOrderStatus(ctx, id, valueobject.OrderStatusCancelled, reason)
}

// ProcessInventoryReserved handles the event when inventory is reserved
func (ou *orderUsecase) ProcessInventoryReserved(ctx context.Context, orderID string, success bool, message string) (*entity.Order, error) {
	// Get the order
	order, err := ou.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}
	fmt.Println("order", order)
	// Update order based on inventory reservation result
	if success {
		// If inventory is successfully reserved and order is still pending, proceed to payment
		if order.Status == valueobject.OrderStatusPending {
			updatedOrder, err := ou.UpdateOrderStatus(ctx, orderID, valueobject.OrderStatusProcessing, "Inventory reserved successfully")
			if err != nil {
				return nil, ou.errBuilder.Err(err)
			}

			// Request payment processing
			if err := ou.eventPub.PublishPaymentRequest(ctx, updatedOrder); err != nil {
				// Log error but continue
				fmt.Println("Error publishing payment request event:", err)
			}

			return updatedOrder, nil
		}
		return order, nil
	} else {
		// If inventory reservation failed, mark the order as failed
		return ou.UpdateOrderStatus(ctx, orderID, valueobject.OrderStatusFailed, "Inventory reservation failed: "+message)
	}
}

// ProcessPaymentCompleted handles the event when payment is completed
func (ou *orderUsecase) ProcessPaymentCompleted(ctx context.Context, orderID string, transactionID string, success bool) (*entity.Order, error) {
	// Get the order
	order, err := ou.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}

	// Update order payment info
	order.Payment.TransactionID = transactionID
	order.Payment.PaidAt = utils.TimePtr(time.Now())

	if success {
		order.Payment.Status = "completed"
		// Update order status to Shipped or other appropriate status
		return ou.UpdateOrderStatus(ctx, orderID, valueobject.OrderStatusShipped, "Payment processed successfully")
	} else {
		order.Payment.Status = "failed"
		// Mark order as failed and release inventory
		updatedOrder, err := ou.UpdateOrderStatus(ctx, orderID, valueobject.OrderStatusFailed, "Payment failed")
		if err != nil {
			return nil, ou.errBuilder.Err(err)
		}

		// Release reserved inventory
		if err := ou.eventPub.PublishReleaseInventory(ctx, updatedOrder); err != nil {
			fmt.Println("Error publishing release inventory event:", err)
			// Log error but continue
		}

		return updatedOrder, nil
	}
}

// UpdateOrderPartial performs a partial update of an order
func (ou *orderUsecase) UpdateOrderPartial(ctx context.Context, id string, patch map[string]interface{}) (*entity.Order, error) {
	// Get the existing order
	existingOrder, err := ou.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ou.errBuilder.Err(entity.ErrOrderNotFound)
	}

	// Create a modifiable copy
	updatedOrder := *existingOrder
	updatedOrder.UpdatedAt = time.Now()

	// Apply updates from patch
	modified := false

	// Handle status updates separately for proper validation and event publishing
	if statusValue, ok := patch["status"]; ok {
		if statusStr, ok := statusValue.(string); ok {
			status, err := valueobject.ParseOrderStatus(statusStr)
			if err == nil && existingOrder.CanTransitionToStatus(status) {
				comment := "Status updated via API"
				if commentValue, ok := patch["comment"]; ok {
					if commentStr, ok := commentValue.(string); ok {
						comment = commentStr
					}
				}

				return ou.UpdateOrderStatus(ctx, id, status, comment)
			}
		}
	}

	// Handle shipping info updates
	if shippingInfo, ok := patch["shipping_info"].(map[string]interface{}); ok {
		if street, ok := shippingInfo["street"].(string); ok {
			updatedOrder.ShippingInfo.Street = street
			modified = true
		}
		if city, ok := shippingInfo["city"].(string); ok {
			updatedOrder.ShippingInfo.City = city
			modified = true
		}
		if state, ok := shippingInfo["state"].(string); ok {
			updatedOrder.ShippingInfo.State = state
			modified = true
		}
		if country, ok := shippingInfo["country"].(string); ok {
			updatedOrder.ShippingInfo.Country = country
			modified = true
		}
		if postalCode, ok := shippingInfo["postal_code"].(string); ok {
			updatedOrder.ShippingInfo.PostalCode = postalCode
			modified = true
		}
	}

	// Handle billing info updates
	if billingInfo, ok := patch["billing_info"].(map[string]interface{}); ok {
		if street, ok := billingInfo["street"].(string); ok {
			updatedOrder.BillingInfo.Street = street
			modified = true
		}
		if city, ok := billingInfo["city"].(string); ok {
			updatedOrder.BillingInfo.City = city
			modified = true
		}
		if state, ok := billingInfo["state"].(string); ok {
			updatedOrder.BillingInfo.State = state
			modified = true
		}
		if country, ok := billingInfo["country"].(string); ok {
			updatedOrder.BillingInfo.Country = country
			modified = true
		}
		if postalCode, ok := billingInfo["postal_code"].(string); ok {
			updatedOrder.BillingInfo.PostalCode = postalCode
			modified = true
		}
	}

	// Handle notes update
	if notes, ok := patch["notes"].(string); ok {
		updatedOrder.Notes = notes
		modified = true
	}

	if !modified {
		return existingOrder, nil
	}

	// Update the order
	updatedOrderRes, err := ou.orderRepo.Update(ctx, updatedOrder)
	if err != nil {
		return nil, ou.errBuilder.Err(err)
	}

	// Publish order updated event
	if err := ou.eventPub.PublishOrderUpdated(ctx, updatedOrderRes); err != nil {
		// Log error but continue
	}

	return updatedOrderRes, nil
}
