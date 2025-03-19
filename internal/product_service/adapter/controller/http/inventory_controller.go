package httpctl

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/adapter/dto"
)

// GetInventory handles retrieving inventory for a product
func (h *ProductHandler) GetInventory(c *fiber.Ctx) error {
	productId := c.Params("productId")
	if productId == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	inventory, err := h.inventoryUsecase.GetInventory(ctx, productId)
	if err != nil {
		h.logger.Error("Failed to get inventory", "productId", productId, "error", err)
		return HandleError(c, err)
	}

	response := dto.InventoryResponseFromEntity(inventory)
	return SuccessResp(c, fiber.StatusOK, "Inventory retrieved successfully", response)
}

// UpdateInventory handles updating the inventory for a product
func (h *ProductHandler) UpdateInventory(c *fiber.Ctx) error {
	productId := c.Params("productId")
	if productId == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.InventoryUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.inventoryUsecase.UpdateInventory(ctx, productId, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to update inventory", "productId", productId, "error", err)
		return HandleError(c, err)
	}

	inventory, _ := h.inventoryUsecase.GetInventory(ctx, productId)
	response := dto.InventoryResponseFromEntity(inventory)
	return SuccessResp(c, fiber.StatusOK, "Inventory updated successfully", response)
}

// ReserveStock handles reserving stock for a product
func (h *ProductHandler) ReserveStock(c *fiber.Ctx) error {
	var req dto.ReservationRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.inventoryUsecase.ReserveStock(ctx, req.ProductID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to reserve stock", "productId", req.ProductID, "quantity", req.Quantity, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Stock reserved successfully", nil)
}

// ReleaseStock handles releasing reserved stock for a product
func (h *ProductHandler) ReleaseStock(c *fiber.Ctx) error {
	var req dto.ReservationRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.inventoryUsecase.CancelReservation(ctx, req.ProductID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to release stock", "productId", req.ProductID, "quantity", req.Quantity, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Stock released successfully", nil)
}

// ConfirmReservation handles confirming a reservation and reducing the actual stock
func (h *ProductHandler) ConfirmReservation(c *fiber.Ctx) error {
	var req dto.ReservationRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	err := h.inventoryUsecase.ConfirmReservation(ctx, req.ProductID, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to confirm reservation", "productId", req.ProductID, "quantity", req.Quantity, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Reservation confirmed successfully", nil)
}

// CheckStock checks if a product is in stock
func (h *ProductHandler) CheckStock(c *fiber.Ctx) error {
	productId := c.Params("productId")
	if productId == "" {
		return HandleError(c, ErrBadRequest)
	}

	quantity, _ := strconv.Atoi(c.Query("quantity", "1"))
	if quantity < 1 {
		quantity = 1
	}

	ctx := c.Context()
	inStock, err := h.inventoryUsecase.IsInStock(ctx, productId, quantity)
	if err != nil {
		h.logger.Error("Failed to check stock", "productId", productId, "quantity", quantity, "error", err)
		return HandleError(c, err)
	}

	return SuccessResp(c, fiber.StatusOK, "Stock check completed", fiber.Map{
		"product_id": productId,
		"quantity":   quantity,
		"in_stock":   inStock,
	})
}
