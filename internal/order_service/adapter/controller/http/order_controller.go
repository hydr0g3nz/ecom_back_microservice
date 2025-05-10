// internal/order_service/adapter/controller/http/order_controller.go
package httpctl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderHandler handles HTTP requests for the order service
type OrderHandler struct {
	orderUsecase usecase.OrderUsecase
	logger       logger.Logger
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(ou usecase.OrderUsecase, l logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderUsecase: ou,
		logger:       l,
	}
}

// RegisterRoutes registers the routes for the order service
func (h *OrderHandler) RegisterRoutes(r fiber.Router) {
	api := r.Group("/orders")

	// Order routes
	api.Post("/", h.CreateOrder)
	api.Get("/", h.ListOrders)
	api.Get("/:id", h.GetOrder)
	api.Put("/:id", h.UpdateOrder)
	api.Patch("/:id", h.PatchOrder)
	api.Delete("/:id", h.CancelOrder)
	api.Post("/:id/status", h.UpdateOrderStatus)
	api.Get("/user/:userId", h.GetOrdersByUser)
}

// CreateOrder handles the creation of a new order
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req dto.OrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	order := req.ToEntity()
	createdOrder, err := h.orderUsecase.CreateOrder(ctx, &order)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(createdOrder)
	return SuccessResp(c, fiber.StatusCreated, "Order created successfully", response)
}

// GetOrder handles retrieving an order by ID
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	order, err := h.orderUsecase.GetOrderByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(order)
	return SuccessResp(c, fiber.StatusOK, "Order retrieved successfully", response)
}

// ListOrders handles retrieving a list of orders
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Build filters from query parameters
	filters := make(map[string]interface{})
	fmt.Println("queries", c.Queries())
	for key, value := range c.Queries() {
		switch key {
		case "status":
			filters["status"] = value
		case "user_id":
			filters["user_id"] = value
		case "created_after":
			createdAfter, err := time.Parse(time.RFC3339, value)
			if err != nil {
				h.logger.Error("Invalid date format for created_after", "error", err)
				return HandleError(c, ErrBadRequest)
			}
			filters["created_after"] = createdAfter
		case "created_before":
			createdBefore, err := time.Parse(time.RFC3339, value)
			if err != nil {
				h.logger.Error("Invalid date format for created_before", "error", err)
				return HandleError(c, ErrBadRequest)
			}
			filters["created_before"] = createdBefore
		case "min_total_amount":
			minTotal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				h.logger.Error("Invalid number format for min_total_amount", "error", err)
				return HandleError(c, ErrBadRequest)
			}
			filters["min_total_amount"] = minTotal
		case "max_total_amount":
			maxTotal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				h.logger.Error("Invalid number format for max_total_amount", "error", err)
				return HandleError(c, ErrBadRequest)
			}
			filters["max_total_amount"] = maxTotal
		}
	}

	ctx := c.Context()
	orders, total, err := h.orderUsecase.ListOrders(ctx, page, pageSize, filters)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseOrders := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		responseOrders[i] = dto.OrderResponseFromEntity(order)
	}

	paginatedResponse := dto.NewPaginatedResponse(total, page, pageSize, responseOrders)
	return SuccessResp(c, fiber.StatusOK, "Orders retrieved successfully", paginatedResponse)
}

// UpdateOrder handles updating an existing order
func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.OrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	order := req.ToEntity()
	updatedOrder, err := h.orderUsecase.UpdateOrder(ctx, id, order)
	if err != nil {
		h.logger.Error("Failed to update order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(updatedOrder)
	return SuccessResp(c, fiber.StatusOK, "Order updated successfully", response)
}

// PatchOrder handles partial updates to an order
func (h *OrderHandler) PatchOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	// Parse request into a map for flexible partial updates
	var patchData map[string]interface{}
	if err := c.BodyParser(&patchData); err != nil {
		h.logger.Error("Failed to parse patch request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	updatedOrder, err := h.orderUsecase.UpdateOrderPartial(ctx, id, patchData)
	if err != nil {
		h.logger.Error("Failed to patch order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(updatedOrder)
	return SuccessResp(c, fiber.StatusOK, "Order updated successfully", response)
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	ctx := c.Context()
	cancelledOrder, err := h.orderUsecase.CancelOrder(ctx, id, req.Reason)
	if err != nil {
		h.logger.Error("Failed to cancel order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(cancelledOrder)
	return SuccessResp(c, fiber.StatusOK, "Order cancelled successfully", response)
}

// UpdateOrderStatus handles updating the status of an order
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return HandleError(c, ErrBadRequest)
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", "error", err)
		return HandleError(c, ErrBadRequest)
	}

	// Parse status
	status, err := valueobject.ParseOrderStatus(req.Status)
	if err != nil {
		h.logger.Error("Invalid order status", "status", req.Status)
		return HandleError(c, err)
	}

	ctx := c.Context()
	updatedOrder, err := h.orderUsecase.UpdateOrderStatus(ctx, id, status, req.Comment)
	if err != nil {
		h.logger.Error("Failed to update order status", "id", id, "status", req.Status, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(updatedOrder)
	return SuccessResp(c, fiber.StatusOK, "Order status updated successfully", response)
}

// GetOrdersByUser handles retrieving orders for a specific user
func (h *OrderHandler) GetOrdersByUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return HandleError(c, ErrBadRequest)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	orders, total, err := h.orderUsecase.GetOrdersByUserID(ctx, userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get orders by user", "userId", userID, "error", err)
		return HandleError(c, err)
	}

	// Convert entities to response DTOs
	responseOrders := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		responseOrders[i] = dto.OrderResponseFromEntity(order)
	}

	paginatedResponse := dto.NewPaginatedResponse(total, page, pageSize, responseOrders)
	return SuccessResp(c, fiber.StatusOK, "User orders retrieved successfully", paginatedResponse)
}
