package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/service"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

// ReservationProcessorUsecase defines the interface for processing reservation-related operations
type ReservationProcessorUsecase interface {
	// ProcessReservationRequest processes a reservation request from an order
	ProcessReservationRequest(ctx context.Context, orderData []byte) error

	// ProcessReleaseRequest processes a request to release reserved inventory
	ProcessReleaseRequest(ctx context.Context, orderData []byte) error

	// ProcessReservationExpiry checks and processes expired reservations
	ProcessReservationExpiry(ctx context.Context) error
}

// reservationProcessorUsecase implements the ReservationProcessorUsecase interface
type reservationProcessorUsecase struct {
	repo        repository.InventoryRepository
	eventPub    service.EventPublisherService
	inventoryUC InventoryUsecase
	errBuilder  *utils.ErrorBuilder
}

// OrderReservationPayload represents the expected structure of order data for reservations
type OrderReservationPayload struct {
	OrderID string         `json:"order_id"`
	Items   map[string]int `json:"items"`
}

// NewReservationProcessorUsecase creates a new instance of ReservationProcessorUsecase
func NewReservationProcessorUsecase(
	repo repository.InventoryRepository,
	eventPub service.EventPublisherService,
	inventoryUC InventoryUsecase,
) ReservationProcessorUsecase {
	return &reservationProcessorUsecase{
		repo:        repo,
		eventPub:    eventPub,
		inventoryUC: inventoryUC,
		errBuilder:  utils.NewErrorBuilder("ReservationProcessorUsecase"),
	}
}

// ProcessReservationRequest processes a reservation request from an order
func (rpu *reservationProcessorUsecase) ProcessReservationRequest(ctx context.Context, orderData []byte) error {
	// Parse order data
	var payload OrderReservationPayload
	if err := json.Unmarshal(orderData, &payload); err != nil {
		return rpu.errBuilder.Err(fmt.Errorf("failed to unmarshal order data: %w", err))
	}

	// Validate payload
	if payload.OrderID == "" {
		return rpu.errBuilder.Err(fmt.Errorf("order ID cannot be empty"))
	}
	if len(payload.Items) == 0 {
		return rpu.errBuilder.Err(fmt.Errorf("order must contain at least one item"))
	}

	// Check for existing reservations for this order
	existingReservations, err := rpu.repo.GetReservationsByOrderID(ctx, payload.OrderID)
	if err != nil && err != entity.ErrInventoryNotFound {
		return rpu.errBuilder.Err(err)
	}

	// If reservations already exist, cancel them first
	if len(existingReservations) > 0 {
		// Check if any reservation is already completed
		for _, res := range existingReservations {
			if res.Status == valueobject.ReserveStatusCompleted.String() {
				return rpu.errBuilder.Err(fmt.Errorf("order %s already has completed reservations", payload.OrderID))
			}
		}

		// Cancel existing reservations
		if err := rpu.inventoryUC.CancelReservation(ctx, payload.OrderID); err != nil {
			return rpu.errBuilder.Err(fmt.Errorf("failed to cancel existing reservations: %w", err))
		}
	}

	// Create new reservations
	reservations, err := rpu.inventoryUC.ReserveStock(ctx, payload.OrderID, payload.Items)
	if err != nil {
		// Publish reservation failed event
		rpu.eventPub.PublishStockReservationFailed(ctx, payload.OrderID, "", err.Error())
		return rpu.errBuilder.Err(fmt.Errorf("failed to reserve stock: %w", err))
	}

	// Publish order reservation created event
	for _, reservation := range reservations {
		if err := rpu.eventPub.PublishStockReserved(ctx, reservation); err != nil {
			// Log error but continue
			fmt.Printf("Error publishing stock reserved event: %v\n", err)
		}
	}

	return nil
}

// ProcessReleaseRequest processes a request to release reserved inventory
func (rpu *reservationProcessorUsecase) ProcessReleaseRequest(ctx context.Context, orderData []byte) error {
	// Parse order data
	var payload struct {
		OrderID string `json:"order_id"`
		Action  string `json:"action"` // "complete" or "cancel"
	}
	if err := json.Unmarshal(orderData, &payload); err != nil {
		return rpu.errBuilder.Err(fmt.Errorf("failed to unmarshal release data: %w", err))
	}

	// Validate payload
	if payload.OrderID == "" {
		return rpu.errBuilder.Err(fmt.Errorf("order ID cannot be empty"))
	}

	// Get existing reservations for this order
	existingReservations, err := rpu.repo.GetReservationsByOrderID(ctx, payload.OrderID)
	if err != nil {
		return rpu.errBuilder.Err(err)
	}

	if len(existingReservations) == 0 {
		return rpu.errBuilder.Err(fmt.Errorf("no reservations found for order %s", payload.OrderID))
	}

	// Process based on action
	switch payload.Action {
	case "complete":
		// Complete the reservation
		if err := rpu.inventoryUC.CompleteReservation(ctx, payload.OrderID); err != nil {
			return rpu.errBuilder.Err(fmt.Errorf("failed to complete reservation: %w", err))
		}
	case "cancel":
		// Cancel the reservation
		if err := rpu.inventoryUC.CancelReservation(ctx, payload.OrderID); err != nil {
			return rpu.errBuilder.Err(fmt.Errorf("failed to cancel reservation: %w", err))
		}
	default:
		return rpu.errBuilder.Err(fmt.Errorf("invalid action: %s, must be 'complete' or 'cancel'", payload.Action))
	}

	return nil
}

// ProcessReservationExpiry checks and processes expired reservations
func (rpu *reservationProcessorUsecase) ProcessReservationExpiry(ctx context.Context) error {
	// Get all expired reservations that are still in RESERVED status
	now := time.Now()

	// In a real implementation, you would add a method to the repository to get expired reservations
	// For now, we'll simulate this by processing all reservations
	// This is a simplification - in production code, add a specific repo method for this

	// Example implementation (assuming repository has GetExpiredReservations method):
	// expiredReservations, err := rpu.repo.GetExpiredReservations(ctx, now)

	// For each expired reservation
	// Here's how you would process each expired reservation:

	// for _, reservation := range expiredReservations {
	// 	// Update reservation status
	// 	reservation.Status = valueobject.ReserveStatusCancelled.String()
	// 	_, err := rpu.repo.UpdateReservation(ctx, reservation)
	// 	if err != nil {
	// 		// Log error but continue processing other reservations
	// 		fmt.Printf("Error updating reservation status: %v\n", err)
	// 		continue
	// 	}

	// 	// Get inventory item
	// 	inventoryItem, err := rpu.repo.GetInventoryItem(ctx, reservation.SKU)
	// 	if err != nil {
	// 		fmt.Printf("Error getting inventory item for SKU %s: %v\n", reservation.SKU, err)
	// 		continue
	// 	}

	// 	// Update inventory item
	// 	inventoryItem.AvailableQty += reservation.Qty
	// 	inventoryItem.ReservedQty -= reservation.Qty
	// 	inventoryItem.UpdatedAt = now

	// 	_, err = rpu.repo.UpdateInventoryItem(ctx, inventoryItem)
	// 	if err != nil {
	// 		fmt.Printf("Error updating inventory item for SKU %s: %v\n", reservation.SKU, err)
	// 		continue
	// 	}

	// 	// Record stock transaction
	// 	refID := reservation.ReservationID
	// 	transaction := &entity.StockTransaction{
	// 		TransactionID: uuid.New().String(),
	// 		SKU:           reservation.SKU,
	// 		Type:          valueobject.StockTypeReleased.String(),
	// 		Qty:           reservation.Qty,
	// 		OccurredAt:    now,
	// 		ReferenceID:   &refID,
	// 	}

	// 	_, err = rpu.repo.RecordStockTransaction(ctx, transaction)
	// 	if err != nil {
	// 		fmt.Printf("Error recording stock transaction: %v\n", err)
	// 	}

	// 	// Publish stock released event
	// 	if err := rpu.eventPub.PublishStockReleased(ctx, reservation); err != nil {
	// 		fmt.Printf("Error publishing stock released event: %v\n", err)
	// 	}
	// }

	// For now, implement a basic version that works with existing repository methods

	// This is a simplified implementation
	// Get all reservations (in production, filter by expiration)
	allReservations, err := rpu.getAllReservations(ctx)
	if err != nil {
		return rpu.errBuilder.Err(fmt.Errorf("failed to get reservations: %w", err))
	}

	expiredCount := 0

	// Find and process expired reservations
	for _, reservation := range allReservations {
		// Only process RESERVED status reservations that have expired
		if reservation.Status == valueobject.ReserveStatusReserved.String() &&
			reservation.ExpiresAt.Before(now) {

			// Process the cancellation using the existing CancelReservation method
			// Group reservations by order ID to avoid multiple cancellations
			if err := rpu.inventoryUC.CancelReservation(ctx, reservation.OrderID); err != nil {
				fmt.Printf("Error cancelling expired reservation for order %s: %v\n",
					reservation.OrderID, err)
				continue
			}

			expiredCount++

			// Publish reservation expired event
			// Implementation depends on your event publishing system
		}
	}

	fmt.Printf("Processed %d expired reservations\n", expiredCount)

	return nil
}

// getAllReservations is a helper method to get all reservations
// In a real implementation, your repository should have a method to directly query expired reservations
func (rpu *reservationProcessorUsecase) getAllReservations(ctx context.Context) ([]*entity.InventoryReservation, error) {
	// This is a placeholder - in a real implementation, you would have a dedicated repository method
	// to efficiently query for all reservations or specifically expired ones

	// For demonstration purposes, we'll return an empty slice
	// In production, implement a proper method in your repository
	return []*entity.InventoryReservation{}, nil
}
