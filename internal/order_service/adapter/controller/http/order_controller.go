package httpctl

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderHandler handles HTTP requests for the order service
type OrderHandler struct {
	orderUsecase        usecase.OrderUsecase
	paymentUsecase      usecase.PaymentUsecase
	shippingUsecase     usecase.ShippingUsecase
	orderHistoryUsecase usecase.OrderHistoryUsecase
	logger              logger.Logger
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	shippingUsecase usecase.ShippingUsecase,
	orderHistoryUsecase usecase.OrderHistoryUsecase,
	logger logger.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderUsecase:        orderUsecase,
		paymentUsecase:      paymentUsecase,
		shippingUsecase:     shippingUsecase,
		orderHistoryUsecase: orderHistoryUsecase,
		logger:              logger,
	}
}

// RegisterRoutes registers all the routes for the order service
func (h *OrderHandler) RegisterRoutes(r fiber.Router) {
	orders := r.Group("/orders")

	// Order routes
	orders.Post("/", h.CreateOrder)
	orders.Get("/:id", h.GetOrder)
	orders.Put("/:id", h.UpdateOrder)
	orders.Delete("/:id", h.CancelOrder)
	orders.Get("/user/:userID", h.GetUserOrders)
	orders.Get("/status/:status", h.GetOrdersByStatus)
	orders.Get("/:id/history", h.GetOrderHistory)
	orders.Post("/search", h.SearchOrders)

	// Payment routes
	orders.Post("/:id/payments", h.ProcessPayment)

	// Shipping routes
	orders.Post("/:id/shipping", h.UpdateShipping)
	orders.Get("/:id/shipping", h.GetShipping)
}

// CreateOrder handles creating a new order
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req dto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse order request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	ctx := c.Context()
	order, err := h.orderUsecase.CreateOrder(ctx, req.ToUsecaseInput())
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(order)
	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetOrder handles retrieving an order by ID
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	ctx := c.Context()
	order, err := h.orderUsecase.GetOrder(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(order)
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateOrder handles updating an existing order
func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req dto.UpdateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse update request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	ctx := c.Context()
	order, err := h.orderUsecase.UpdateOrder(ctx, id, req.ToUsecaseInput())
	if err != nil {
		h.logger.Error("Failed to update order", "id", id, "error", err)
		return HandleError(c, err)
	}

	response := dto.OrderResponseFromEntity(order)
	return c.Status(fiber.StatusOK).JSON(response)
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req dto.CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse cancel request", "error", err)
		// Default reason if not provided
		req.Reason = "Cancelled by customer"
	}

	ctx := c.Context()
	err := h.orderUsecase.CancelOrder(ctx, id, req.Reason)
	if err != nil {
		h.logger.Error("Failed to cancel order", "id", id, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Order cancelled successfully",
		"order_id": id,
	})
}

// GetUserOrders handles retrieving orders for a user
func (h *OrderHandler) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Params("userID")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	orders, total, err := h.orderUsecase.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user orders", "user_id", userID, "error", err)
		return HandleError(c, err)
	}

	response := dto.CreatePaginatedResponse(orders, total, page, pageSize)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetOrdersByStatus handles retrieving orders with a specific status
func (h *OrderHandler) GetOrdersByStatus(c *fiber.Ctx) error {
	statusStr := c.Params("status")
	if statusStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	status, err := valueobject.ParseOrderStatus(statusStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	orders, total, err := h.orderUsecase.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get orders by status", "status", statusStr, "error", err)
		return HandleError(c, err)
	}

	response := dto.CreatePaginatedResponse(orders, total, page, pageSize)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetOrderHistory handles retrieving the event history for an order
func (h *OrderHandler) GetOrderHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	ctx := c.Context()
	events, err := h.orderHistoryUsecase.GetEvents(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get order history", "id", id, "error", err)
		return HandleError(c, err)
	}

	// Convert events to response format
	eventResponses := make([]dto.OrderEventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = dto.OrderEventResponseFromEntity(event)
	}

	response := dto.OrderHistoryResponse{
		OrderID: id,
		Events:  eventResponses,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// SearchOrders handles searching orders based on criteria
func (h *OrderHandler) SearchOrders(c *fiber.Ctx) error {
	var req dto.SearchOrdersRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse search request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	criteria := req.ToCriteria()

	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	orders, total, err := h.orderUsecase.Search(ctx, criteria, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search orders", "error", err)
		return HandleError(c, err)
	}

	response := dto.CreatePaginatedResponse(orders, total, page, pageSize)
	return c.Status(fiber.StatusOK).JSON(response)
}

// ProcessPayment handles processing a payment for an order
func (h *OrderHandler) ProcessPayment(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req dto.ProcessPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse payment request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	ctx := c.Context()
	payment, err := h.paymentUsecase.ProcessPayment(ctx, req.ToUsecaseInput(orderID))
	if err != nil {
		h.logger.Error("Failed to process payment", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	response := dto.PaymentResponseFromEntity(payment)
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateShipping handles updating shipping information for an order
func (h *OrderHandler) UpdateShipping(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req dto.UpdateShippingRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse shipping request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	input, err := req.ToUsecaseInput(orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid shipping status",
		})
	}

	ctx := c.Context()
	shipping, err := h.shippingUsecase.UpdateShipping(ctx, input)
	if err != nil {
		h.logger.Error("Failed to update shipping", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	response := dto.ShippingResponseFromEntity(shipping)
	return c.Status(fiber.StatusOK).JSON(response)
}

// GetShipping handles retrieving shipping information for an order
func (h *OrderHandler) GetShipping(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	ctx := c.Context()
	shipping, err := h.shippingUsecase.GetShippingByOrderID(ctx, orderID)
	if err != nil {
		h.logger.Error("Failed to get shipping information", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	response := dto.ShippingResponseFromEntity(shipping)
	return c.Status(fiber.StatusOK).JSON(response)
}
