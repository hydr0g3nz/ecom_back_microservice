package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
)

// InventoryEventHandler handles events related to inventory
type InventoryEventHandler struct {
	orderRepo          repository.OrderRepository
	orderEventRepo     repository.OrderEventRepository
	cancelOrderUsecase command.CancelOrderUsecase
}

// NewInventoryEventHandler creates a new instance of InventoryEventHandler
func NewInventoryEventHandler(
	orderRepo repository.OrderRepository,
	orderEventRepo repository.OrderEventRepository,
	cancelOrderUsecase command.CancelOrderUsecase,
) *InventoryEventHandler {
	return &InventoryEventHandler{
		orderRepo:          orderRepo,
		orderEventRepo:     orderEventRepo,
		cancelOrderUsecase: cancelOrderUsecase,
	}
}

// HandleStockReservationFailed handles a stock reservation failure event
func (h *InventoryEventHandler) HandleStockReservationFailed(ctx context.Context, eventData []byte) error {
	// Parse the event data
	var event struct {
		OrderID  string   `json:"order_id"`
		Products []string `json:"products"`
		Reason   string   `json:"reason"`
	}

	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("error unmarshaling stock reservation failed event: %w", err)
	}

	// Get the order
	order, err := h.orderRepo.GetByID(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("error getting order %s: %w", event.OrderID, err)
	}

	// Check if order can be cancelled
	if !order.CanCancel() {
		log.Printf("Order %s cannot be cancelled, already in status %s", order.ID, order.Status)
		return nil
	}

	// Cancel the order
	reason := fmt.Sprintf("Stock reservation failed: %s", event.Reason)
	return h.cancelOrderUsecase.Execute(ctx, order.ID, reason)
}

// HandleStockReleased handles a stock release event
func (h *InventoryEventHandler) HandleStockReleased(ctx context.Context, eventData []byte) error {
	// Parse the event data
	var event struct {
		OrderID string `json:"order_id"`
		Success bool   `json:"success"`
		Reason  string `json:"reason"`
	}

	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("error unmarshaling stock released event: %w", err)
	}

	// Create and store event for audit purposes
	orderEvent := &entity.OrderEvent{
		OrderID:   event.OrderID,
		Type:      "stock_released",
		Data:      eventData,
		Timestamp: time.Now(),
	}

	return h.orderEventRepo.SaveEvent(ctx, orderEvent)
}
