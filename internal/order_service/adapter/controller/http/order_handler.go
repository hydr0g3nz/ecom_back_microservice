package http

import (
	"github.com/gofiber/fiber/v2"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/http/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	applogger "github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderUsecase    usecase.OrderUsecase
	paymentUsecase  usecase.PaymentUsecase
	shippingUsecase usecase.ShippingUsecase
	queryUsecase    usecase.QueryUsecase
	logger          applogger.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	shippingUsecase usecase.ShippingUsecase,
	queryUsecase usecase.QueryUsecase,
	logger applogger.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderUsecase:    orderUsecase,
		paymentUsecase:  paymentUsecase,
		shippingUsecase: shippingUsecase,
		queryUsecase:    queryUsecase,
		logger:          logger,
	}
}

// RegisterRoutes registers all routes
func (h *OrderHandler) RegisterRoutes(router fiber.Router) {
	orders := router.Group("/orders")

	// Order endpoints
	orders.Post("/", h.CreateOrder)
	orders.Get("/", h.ListOrders)
	orders.Get("/:id", h.GetOrder)
	orders.Patch("/:id", h.UpdateOrder)
	orders.Delete("/:id", h.CancelOrder)
	
	// Order history
	orders.Get("/:id/history", h.GetOrderHistory)
	
	// Payment endpoints
	orders.Post("/:id/payments", h.ProcessPayment)
	orders.Post("/:id/refunds", h.RefundPayment)
	
	// Shipping endpoints
	orders.Post("/:id/shipping", h.UpdateShipping)
	orders.Get("/:id/tracking", h.TrackShipment)
}

// CreateOrder handles order creation
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req dto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Convert DTO to domain entity
	order := req.ToEntity()

	// Call usecase
	createdOrder, err := h.orderUsecase.CreateOrder(c.Context(), order)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		return handleDomainError(err)
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(dto.OrderToResponse(createdOrder))
}

// GetOrder handles getting a single order
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	// Call usecase
	order, err := h.queryUsecase.GetOrder(c.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return response
	return c.JSON(dto.OrderToResponse(order))
}

// ListOrders handles listing orders with filtering and pagination
func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	
	// Build filter
	filter := make(map[string]interface{})
	
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	
	if userID := c.Query("user_id"); userID != "" {
		filter["user_id"] = userID
	}
	
	// Call usecase
	orders, total, err := h.queryUsecase.ListOrders(c.Context(), filter, page, limit)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		return handleDomainError(err)
	}

	// Convert to response DTOs
	var responseOrders []dto.OrderResponse
	for _, order := range orders {
		responseOrders = append(responseOrders, dto.OrderToResponse(order))
	}

	// Return response with pagination metadata
	return c.JSON(fiber.Map{
		"data": responseOrders,
		"meta": fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// UpdateOrder handles order updates
func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	var req dto.UpdateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Convert DTO to map of updates
	updates := req.ToMap()
	if len(updates) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No updates provided")
	}

	// Call usecase
	updatedOrder, err := h.orderUsecase.UpdateOrder(c.Context(), id, updates)
	if err != nil {
		h.logger.Error("Failed to update order", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return response
	return c.JSON(dto.OrderToResponse(updatedOrder))
}

// CancelOrder handles order cancellation
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	var req dto.CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Call usecase
	if err := h.orderUsecase.CancelOrder(c.Context(), id, req.Reason); err != nil {
		h.logger.Error("Failed to cancel order", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order cancelled successfully",
		"id":      id,
	})
}

// GetOrderHistory handles retrieving order history
func (h *OrderHandler) GetOrderHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	// Call usecase
	events, err := h.queryUsecase.GetOrderHistory(c.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order history", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Convert to response DTOs
	var responseEvents []dto.OrderEventResponse
	for _, event := range events {
		responseEvents = append(responseEvents, dto.OrderEventToResponse(event))
	}

	// Return response
	return c.JSON(responseEvents)
}

// ProcessPayment handles payment processing
func (h *OrderHandler) ProcessPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	var req dto.ProcessPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Convert DTO to domain entity
	payment := req.ToEntity()

	// Call usecase
	if err := h.paymentUsecase.ProcessPayment(c.Context(), id, payment); err != nil {
		h.logger.Error("Failed to process payment", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment processed successfully",
		"order_id": id,
	})
}

// RefundPayment handles payment refunds
func (h *OrderHandler) RefundPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	var req dto.RefundPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Call usecase
	if err := h.paymentUsecase.RefundPayment(c.Context(), id, req.Amount, req.Reason); err != nil {
		h.logger.Error("Failed to process refund", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Refund processed successfully",
		"order_id": id,
		"amount": req.Amount,
	})
}

// UpdateShipping handles shipping updates
func (h *OrderHandler) UpdateShipping(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	var req dto.UpdateShippingRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Convert DTO to domain entity
	shipping := req.ToEntity()

	// Call usecase
	if err := h.shippingUsecase.UpdateShipping(c.Context(), id, shipping); err != nil {
		h.logger.Error("Failed to update shipping", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Shipping updated successfully",
		"order_id": id,
	})
}

// TrackShipment handles shipment tracking
func (h *OrderHandler) TrackShipment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	// Call usecase
	shipping, err := h.shippingUsecase.TrackShipment(c.Context(), id)
	if err != nil {
		h.logger.Error("Failed to track shipment", "error", err, "id", id)
		return handleDomainError(err)
	}

	// Return tracking information
	return c.JSON(dto.ShippingToResponse(shipping))
}

// handleDomainError converts domain errors to appropriate HTTP errors
func handleDomainError(err error) error {
	switch err {
	case entity.ErrOrderNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Order not found")
	case entity.ErrPaymentNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Payment not found")
	case entity.ErrShippingNotFound:
		return fiber.NewError(fiber.StatusNotFound, "Shipping information not found")
	case entity.ErrOrderCannotBeUpdated:
		return fiber.NewError(fiber.StatusConflict, "Order cannot be updated in its current state")
	case entity.ErrOrderCannotBeCancelled:
		return fiber.NewError(fiber.StatusConflict, "Order cannot be cancelled in its current state")
	case entity.ErrOrderCannotProcessPayment:
		return fiber.NewError(fiber.StatusConflict, "Cannot process payment for order in its current state")
	case entity.ErrOrderCannotBeRefunded:
		return fiber.NewError(fiber.StatusConflict, "Order cannot be refunded in its current state")
	case entity.ErrOrderCannotUpdateShipping:
		return fiber.NewError(fiber.StatusConflict, "Cannot update shipping for order in its current state")
	case entity.ErrPaymentCannotBeRefunded:
		return fiber.NewError(fiber.StatusConflict, "Payment cannot be refunded in its current state")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}