// internal/inventory_service/adapter/httpctl/inventory_handler.go
package httpctl

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10" // Assuming you use a validator
	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/adapter/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/domain/entity"
	uc "github.com/hydr0g3nz/ecom_back_microservice/internal/inventory_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
	"gorm.io/gorm"
)

// InventoryHandler handles HTTP requests for the inventory service
type InventoryHandler struct {
	usecase  uc.InventoryUsecase
	logger   logger.Logger
	validate *validator.Validate // Assuming you use a validator
}

// NewInventoryHandler creates a new instance of InventoryHandler
func NewInventoryHandler(usecase uc.InventoryUsecase, logger logger.Logger) *InventoryHandler {
	return &InventoryHandler{
		usecase:  usecase,
		logger:   logger,
		validate: validator.New(), // Initialize validator
	}
}

// RegisterRoutes registers the routes for the inventory service
func (h *InventoryHandler) RegisterRoutes(r fiber.Router) {
	inventoryGroup := r.Group("/inventory")

	inventoryGroup.Post("/", h.CreateInventoryItem)
	inventoryGroup.Get("/low-stock", h.GetLowStockItems)
	inventoryGroup.Get("/:sku", h.GetInventoryItem)
	inventoryGroup.Put("/:sku", h.UpdateInventoryItem)
	inventoryGroup.Post("/:sku/stock/add", h.AddStock) // e.g., /inventory/SKU123/stock/add
	inventoryGroup.Post("/reserve", h.ReserveStock)
	inventoryGroup.Post("/:orderID/complete", h.CompleteReservation)         // e.g., /inventory/ORDERID456/complete
	inventoryGroup.Post("/:orderID/cancel", h.CancelReservation)             // e.g., /inventory/ORDERID456/cancel
	inventoryGroup.Get("/reservations/:orderID", h.GetReservationsByOrderID) // e.g., /inventory/reservations/ORDERID456
	inventoryGroup.Get("/:sku/transactions", h.GetStockTransactionHistory)   // e.g., /inventory/SKU123/transactions
}

// updateHandleError should be a shared helper function or extended
// For simplicity, I'll include a localized version here that includes inventory errors.
// In a real application, this should be part of a shared adapter/httpctl package.
func (h *InventoryHandler) handleInventoryError(c *fiber.Ctx, err error) error {
	var statusCode int
	var message string

	// Log the actual error for debugging purposes (optional based on logging policy)
	h.logger.Error("Handler encountered an error", "error", err)

	switch {
	case errors.Is(err, ErrBadRequest): // Assuming ErrBadRequest is defined in the same package
		statusCode = http.StatusBadRequest
		message = "Bad request"
	case errors.Is(err, entity.ErrInventoryNotFound):
		statusCode = http.StatusNotFound
		message = "Inventory item or reservation not found" // Message can be generic
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = http.StatusBadRequest // Or StatusConflict, depending on preferred semantics
		message = "Insufficient stock"
	case errors.Is(err, entity.ErrInvalidProductData):
		statusCode = http.StatusBadRequest
		message = "Invalid product data"
	case errors.Is(err, gorm.ErrRecordNotFound):
		statusCode = http.StatusNotFound
		message = "Record not found"
	case errors.Is(err, entity.ErrSKUAlreadyExists):
		statusCode = http.StatusConflict
		message = "SKU already exists"
	// Add other specific domain errors here
	default:
		// Fallback for unexpected errors
		statusCode = http.StatusInternalServerError
		message = "Something went wrong"
	}

	// Use the ErrorResponse struct defined in the shared httpctl package
	return c.Status(statusCode).JSON(ErrorResponse{
		Status:  statusCode,
		Message: message,
	})
}

// CreateInventoryItem handles the creation of a new inventory item
// POST /inventory
func (h *InventoryHandler) CreateInventoryItem(c *fiber.Ctx) error {
	var req dto.CreateInventoryItemRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body for CreateInventoryItem", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("Request validation failed for CreateInventoryItem", "error", err)
		return h.handleInventoryError(c, ErrBadRequest) // Or a more specific validation error handler
	}

	ctx := c.Context()
	itemEntity := req.ToEntity()
	newItem, err := h.usecase.CreateInventoryItem(ctx, &itemEntity)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	// Use the SuccessResp struct defined in the shared httpctl package
	return SuccessResp(c, fiber.StatusCreated, "Inventory item created", newItem)
}

// GetInventoryItem handles retrieving an inventory item by SKU
// GET /inventory/:sku
func (h *InventoryHandler) GetInventoryItem(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	item, err := h.usecase.GetInventoryItem(ctx, sku)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Inventory item retrieved", item)
}

// UpdateInventoryItem handles updating an existing inventory item
// PUT /inventory/:sku
func (h *InventoryHandler) UpdateInventoryItem(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	var req dto.UpdateInventoryItemRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body for UpdateInventoryItem", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("Request validation failed for UpdateInventoryItem", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	// Create entity from DTO, ensuring SKU is set from the path param
	itemEntity := req.ToEntity(sku)
	updatedItem, err := h.usecase.UpdateInventoryItem(ctx, &itemEntity)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Inventory item updated", updatedItem)
}

// AddStock handles adding stock to an inventory item
// POST /inventory/:sku/stock/add
func (h *InventoryHandler) AddStock(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	var req dto.AddStockRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body for AddStock", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("Request validation failed for AddStock", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	updatedItem, err := h.usecase.AddStock(ctx, sku, req.Quantity, req.ReferenceID)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Stock added successfully", updatedItem)
}

// ReserveStock handles reserving stock for an order
// POST /inventory/reserve
func (h *InventoryHandler) ReserveStock(c *fiber.Ctx) error {
	var req dto.ReserveStockRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to decode request body for ReserveStock", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("Request validation failed for ReserveStock", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	reservations, err := h.usecase.ReserveStock(ctx, req.OrderID, req.Items)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	return SuccessResp(c, fiber.StatusCreated, "Stock reserved successfully", reservations)
}

// CompleteReservation marks a reservation as completed and deducts stock
// POST /inventory/:orderID/complete
func (h *InventoryHandler) CompleteReservation(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.usecase.CompleteReservation(ctx, orderID)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	// No content to return on success for completion
	return SuccessResp(c, fiber.StatusNoContent, "Reservation completed successfully", nil)
}

// CancelReservation cancels a reservation and releases stock
// POST /inventory/:orderID/cancel
func (h *InventoryHandler) CancelReservation(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.usecase.CancelReservation(ctx, orderID)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	// No content to return on success for cancellation
	return SuccessResp(c, fiber.StatusNoContent, "Reservation cancelled successfully", nil)
}

// GetReservationsByOrderID gets all reservations for an order
// GET /inventory/reservations/:orderID
func (h *InventoryHandler) GetReservationsByOrderID(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	ctx := c.Context()
	reservations, err := h.usecase.GetReservationsByOrderID(ctx, orderID)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Reservations retrieved", reservations)
}

// GetStockTransactionHistory gets the transaction history for a SKU with pagination
// GET /inventory/:sku/transactions?page=1&pageSize=10
func (h *InventoryHandler) GetStockTransactionHistory(c *fiber.Ctx) error {
	sku := c.Params("sku")
	if sku == "" {
		return h.handleInventoryError(c, ErrBadRequest)
	}

	var req dto.GetTransactionHistoryRequest
	// Use QueryParser for query parameters
	if err := c.QueryParser(&req); err != nil {
		h.logger.Error("Failed to parse query params for GetStockTransactionHistory", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	// Set default pagination values if not provided or invalid
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10 // Default page size
	}

	// No struct validation needed if defaults are handled

	ctx := c.Context()
	transactions, total, err := h.usecase.GetStockTransactionHistory(ctx, sku, req.Page, req.PageSize)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	// Return a response structure that includes both the list of transactions and the total count
	responsePayload := dto.StockTransactionsWithTotal{
		Transactions: transactions,
		Total:        total,
	}

	return SuccessResp(c, fiber.StatusOK, "Stock transaction history retrieved", responsePayload)
}

// GetLowStockItems gets items with stock below reorder level with pagination
// GET /inventory/low-stock?page=1&pageSize=10
func (h *InventoryHandler) GetLowStockItems(c *fiber.Ctx) error {
	var req dto.GetLowStockItemsRequest
	if err := c.QueryParser(&req); err != nil {
		h.logger.Error("Failed to parse query params for GetLowStockItems", "error", err)
		return h.handleInventoryError(c, ErrBadRequest)
	}

	// Set default pagination values if not provided or invalid
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10 // Default page size
	}

	// No struct validation needed if defaults are handled

	ctx := c.Context()
	items, total, err := h.usecase.GetLowStockItems(ctx, req.Page, req.PageSize)
	if err != nil {
		return h.handleInventoryError(c, err)
	}

	// Return a response structure that includes both the list of items and the total count
	responsePayload := dto.InventoryItemsWithTotal{
		Items: items,
		Total: total,
	}

	return SuccessResp(c, fiber.StatusOK, "Low stock items retrieved", responsePayload)
}

// Note: Ensure the shared httpctl package (containing SuccessResp, ErrorResponse, HandleError, ErrBadRequest)
// is accessible or copy the necessary parts into this file if it's not a shared package.
// In a multi-service monorepo, a shared 'pkg/httpctl' would be ideal.
// I've included a local handleInventoryError for demonstration,
// but ideally, the shared HandleError would be extended or refactored.
