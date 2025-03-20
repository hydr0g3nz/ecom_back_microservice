package httpctl

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/query"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderHandler handles HTTP requests for the order service
type OrderHandler struct {
	createOrderUsecase    command.CreateOrderUsecase
	updateOrderUsecase    command.UpdateOrderUsecase
	cancelOrderUsecase    command.CancelOrderUsecase
	processPaymentUsecase command.ProcessPaymentUsecase
	updateShippingUsecase command.UpdateShippingUsecase
	getOrderUsecase       query.GetOrderUsecase
	listOrdersUsecase     query.ListOrdersUsecase
	orderHistoryUsecase   query.OrderHistoryUsecase
	shippingRepository    repository.ShippingRepository
	logger                logger.Logger
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(
	createOrderUsecase command.CreateOrderUsecase,
	updateOrderUsecase command.UpdateOrderUsecase,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
	getOrderUsecase query.GetOrderUsecase,
	listOrdersUsecase query.ListOrdersUsecase,
	orderHistoryUsecase query.OrderHistoryUsecase,
	shippingRepository repository.ShippingRepository,
	logger logger.Logger,
) *OrderHandler {
	return &OrderHandler{
		createOrderUsecase:    createOrderUsecase,
		updateOrderUsecase:    updateOrderUsecase,
		cancelOrderUsecase:    cancelOrderUsecase,
		processPaymentUsecase: processPaymentUsecase,
		updateShippingUsecase: updateShippingUsecase,
		getOrderUsecase:       getOrderUsecase,
		listOrdersUsecase:     listOrdersUsecase,
		orderHistoryUsecase:   orderHistoryUsecase,
		shippingRepository:    shippingRepository,
		logger:                logger,
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

// OrderItemRequest represents an order item in the request
type OrderItemRequest struct {
	ProductID    string  `json:"product_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	CurrencyCode string  `json:"currency_code"`
}

// AddressRequest represents an address in the request
type AddressRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	UserID          string             `json:"user_id"`
	Items           []OrderItemRequest `json:"items"`
	ShippingAddress AddressRequest     `json:"shipping_address"`
	BillingAddress  AddressRequest     `json:"billing_address"`
	Notes           string             `json:"notes"`
	PromotionCodes  []string           `json:"promotion_codes"`
}

// CreateOrder handles creating a new order
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse order request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Convert request to domain model
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID:    item.ProductID,
			Name:         item.Name,
			SKU:          item.SKU,
			Quantity:     item.Quantity,
			Price:        item.Price,
			CurrencyCode: item.CurrencyCode,
		}
	}

	input := command.CreateOrderInput{
		UserID: req.UserID,
		Items:  items,
		ShippingAddress: entity.Address{
			FirstName:    req.ShippingAddress.FirstName,
			LastName:     req.ShippingAddress.LastName,
			AddressLine1: req.ShippingAddress.AddressLine1,
			AddressLine2: req.ShippingAddress.AddressLine2,
			City:         req.ShippingAddress.City,
			State:        req.ShippingAddress.State,
			PostalCode:   req.ShippingAddress.PostalCode,
			Country:      req.ShippingAddress.Country,
			Phone:        req.ShippingAddress.Phone,
			Email:        req.ShippingAddress.Email,
		},
		BillingAddress: entity.Address{
			FirstName:    req.BillingAddress.FirstName,
			LastName:     req.BillingAddress.LastName,
			AddressLine1: req.BillingAddress.AddressLine1,
			AddressLine2: req.BillingAddress.AddressLine2,
			City:         req.BillingAddress.City,
			State:        req.BillingAddress.State,
			PostalCode:   req.BillingAddress.PostalCode,
			Country:      req.BillingAddress.Country,
			Phone:        req.BillingAddress.Phone,
			Email:        req.BillingAddress.Email,
		},
		Notes:          req.Notes,
		PromotionCodes: req.PromotionCodes,
	}

	ctx := c.Context()
	order, err := h.createOrderUsecase.Execute(ctx, input)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(OrderToResponse(order))
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
	order, err := h.getOrderUsecase.Execute(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get order", "id", id, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(OrderToResponse(order))
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	Notes           *string         `json:"notes,omitempty"`
	ShippingAddress *AddressRequest `json:"shipping_address,omitempty"`
	BillingAddress  *AddressRequest `json:"billing_address,omitempty"`
}

// UpdateOrder handles updating an existing order
func (h *OrderHandler) UpdateOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req UpdateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse update request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	var shippingAddress *entity.Address
	var billingAddress *entity.Address
	var notes *string

	if req.ShippingAddress != nil {
		addr := entity.Address{
			FirstName:    req.ShippingAddress.FirstName,
			LastName:     req.ShippingAddress.LastName,
			AddressLine1: req.ShippingAddress.AddressLine1,
			AddressLine2: req.ShippingAddress.AddressLine2,
			City:         req.ShippingAddress.City,
			State:        req.ShippingAddress.State,
			PostalCode:   req.ShippingAddress.PostalCode,
			Country:      req.ShippingAddress.Country,
			Phone:        req.ShippingAddress.Phone,
			Email:        req.ShippingAddress.Email,
		}
		shippingAddress = &addr
	}

	if req.BillingAddress != nil {
		addr := entity.Address{
			FirstName:    req.BillingAddress.FirstName,
			LastName:     req.BillingAddress.LastName,
			AddressLine1: req.BillingAddress.AddressLine1,
			AddressLine2: req.BillingAddress.AddressLine2,
			City:         req.BillingAddress.City,
			State:        req.BillingAddress.State,
			PostalCode:   req.BillingAddress.PostalCode,
			Country:      req.BillingAddress.Country,
			Phone:        req.BillingAddress.Phone,
			Email:        req.BillingAddress.Email,
		}
		billingAddress = &addr
	}

	if req.Notes != nil {
		notes = req.Notes
	}

	input := command.UpdateOrderInput{
		Notes:           notes,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
	}

	ctx := c.Context()
	order, err := h.updateOrderUsecase.Execute(ctx, id, input)
	if err != nil {
		h.logger.Error("Failed to update order", "id", id, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(OrderToResponse(order))
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

// CancelOrder handles cancelling an order
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse cancel request", "error", err)
		// Default reason if not provided
		req.Reason = "Cancelled by customer"
	}

	ctx := c.Context()
	err := h.cancelOrderUsecase.Execute(ctx, id, req.Reason)
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
	orders, total, err := h.listOrdersUsecase.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get user orders", "user_id", userID, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(CreatePaginatedResponse(orders, total, page, pageSize))
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
	orders, total, err := h.listOrdersUsecase.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get orders by status", "status", statusStr, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(CreatePaginatedResponse(orders, total, page, pageSize))
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
	eventResponses := make([]map[string]interface{}, len(events))
	for i, event := range events {
		eventResponses[i] = map[string]interface{}{
			"id":        event.ID,
			"order_id":  event.OrderID,
			"type":      event.Type,
			"data":      string(event.Data), // Convert to string for JSON display
			"version":   event.Version,
			"timestamp": event.Timestamp,
			"user_id":   event.UserID,
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"order_id": id,
		"events":   eventResponses,
	})
}

// SearchOrdersRequest represents the request to search orders
type SearchOrdersRequest struct {
	UserID    string     `json:"user_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	ProductID string     `json:"product_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	MinAmount float64    `json:"min_amount,omitempty"`
	MaxAmount float64    `json:"max_amount,omitempty"`
	Page      int        `json:"page,omitempty"`
	PageSize  int        `json:"page_size,omitempty"`
}

// SearchOrders handles searching orders based on criteria
func (h *OrderHandler) SearchOrders(c *fiber.Ctx) error {
	var req SearchOrdersRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse search request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	criteria := make(map[string]interface{})

	if req.UserID != "" {
		criteria["user_id"] = req.UserID
	}

	if req.Status != "" {
		criteria["status"] = req.Status
	}

	if req.ProductID != "" {
		criteria["product_id"] = req.ProductID
	}

	if req.StartDate != nil {
		criteria["start_date"] = req.StartDate
	}

	if req.EndDate != nil {
		criteria["end_date"] = req.EndDate
	}

	if req.MinAmount > 0 {
		criteria["min_amount"] = req.MinAmount
	}

	if req.MaxAmount > 0 {
		criteria["max_amount"] = req.MaxAmount
	}

	page := req.Page
	pageSize := req.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	ctx := c.Context()
	orders, total, err := h.listOrdersUsecase.Search(ctx, criteria, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search orders", "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(CreatePaginatedResponse(orders, total, page, pageSize))
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Method          string  `json:"method"`
	TransactionID   string  `json:"transaction_id"`
	GatewayResponse string  `json:"gateway_response"`
}

// ProcessPayment handles processing a payment for an order
func (h *OrderHandler) ProcessPayment(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req ProcessPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse payment request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	input := command.ProcessPaymentInput{
		OrderID:         orderID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Method:          req.Method,
		TransactionID:   req.TransactionID,
		GatewayResponse: req.GatewayResponse,
	}

	ctx := c.Context()
	payment, err := h.processPaymentUsecase.Execute(ctx, input)
	if err != nil {
		h.logger.Error("Failed to process payment", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(PaymentToResponse(payment))
}

// UpdateShippingRequest represents the request to update shipping information
type UpdateShippingRequest struct {
	Carrier           string     `json:"carrier"`
	TrackingNumber    string     `json:"tracking_number"`
	Status            string     `json:"status"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	ShippingMethod    string     `json:"shipping_method"`
	ShippingCost      float64    `json:"shipping_cost"`
	Notes             string     `json:"notes,omitempty"`
}

// UpdateShipping handles updating shipping information for an order
func (h *OrderHandler) UpdateShipping(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	var req UpdateShippingRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse shipping request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	status, err := valueobject.ParseShippingStatus(req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid shipping status",
		})
	}

	input := command.UpdateShippingInput{
		OrderID:           orderID,
		Carrier:           req.Carrier,
		TrackingNumber:    req.TrackingNumber,
		Status:            status,
		EstimatedDelivery: req.EstimatedDelivery,
		ShippingMethod:    req.ShippingMethod,
		ShippingCost:      req.ShippingCost,
		Notes:             req.Notes,
	}

	ctx := c.Context()
	shipping, err := h.updateShippingUsecase.Execute(ctx, input)
	if err != nil {
		h.logger.Error("Failed to update shipping", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(ShippingToResponse(shipping))
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

	// First get the order to confirm it exists and has shipping info
	order, err := h.getOrderUsecase.Execute(ctx, orderID)
	if err != nil {
		h.logger.Error("Failed to get order for shipping", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	if order.ShippingID == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No shipping information found for this order",
		})
	}

	// Get shipping info
	shipping, err := h.shippingRepository.GetByOrderID(ctx, orderID)
	if err != nil {
		h.logger.Error("Failed to get shipping information", "order_id", orderID, "error", err)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(ShippingToResponse(shipping))
}

// Helper functions for creating responses
func OrderToResponse(order *entity.Order) map[string]interface{} {
	items := make([]map[string]interface{}, len(order.Items))
	for i, item := range order.Items {
		items[i] = map[string]interface{}{
			"id":            item.ID,
			"product_id":    item.ProductID,
			"name":          item.Name,
			"sku":           item.SKU,
			"quantity":      item.Quantity,
			"price":         item.Price,
			"total_price":   item.TotalPrice,
			"currency_code": item.CurrencyCode,
		}
	}

	discounts := make([]map[string]interface{}, len(order.Discounts))
	for i, discount := range order.Discounts {
		discounts[i] = map[string]interface{}{
			"code":        discount.Code,
			"description": discount.Description,
			"type":        discount.Type,
			"amount":      discount.Amount,
		}
	}

	response := map[string]interface{}{
		"id":              order.ID,
		"user_id":         order.UserID,
		"items":           items,
		"total_amount":    order.TotalAmount,
		"status":          order.Status.String(),
		"payment_id":      order.PaymentID,
		"shipping_id":     order.ShippingID,
		"notes":           order.Notes,
		"promotion_codes": order.PromotionCodes,
		"discounts":       discounts,
		"tax_amount":      order.TaxAmount,
		"created_at":      order.CreatedAt,
		"updated_at":      order.UpdatedAt,
		"version":         order.Version,
		"shipping_address": map[string]interface{}{
			"first_name":    order.ShippingAddress.FirstName,
			"last_name":     order.ShippingAddress.LastName,
			"address_line1": order.ShippingAddress.AddressLine1,
			"address_line2": order.ShippingAddress.AddressLine2,
			"city":          order.ShippingAddress.City,
			"state":         order.ShippingAddress.State,
			"postal_code":   order.ShippingAddress.PostalCode,
			"country":       order.ShippingAddress.Country,
			"phone":         order.ShippingAddress.Phone,
			"email":         order.ShippingAddress.Email,
		},
		"billing_address": map[string]interface{}{
			"first_name":    order.BillingAddress.FirstName,
			"last_name":     order.BillingAddress.LastName,
			"address_line1": order.BillingAddress.AddressLine1,
			"address_line2": order.BillingAddress.AddressLine2,
			"city":          order.BillingAddress.City,
			"state":         order.BillingAddress.State,
			"postal_code":   order.BillingAddress.PostalCode,
			"country":       order.BillingAddress.Country,
			"phone":         order.BillingAddress.Phone,
			"email":         order.BillingAddress.Email,
		},
	}

	if order.CompletedAt != nil {
		response["completed_at"] = order.CompletedAt
	}

	if order.CancelledAt != nil {
		response["cancelled_at"] = order.CancelledAt
	}

	return response
}

func PaymentToResponse(payment *entity.Payment) map[string]interface{} {
	response := map[string]interface{}{
		"id":               payment.ID,
		"order_id":         payment.OrderID,
		"amount":           payment.Amount,
		"currency":         payment.Currency,
		"method":           payment.Method,
		"status":           payment.Status,
		"transaction_id":   payment.TransactionID,
		"gateway_response": payment.GatewayResponse,
		"created_at":       payment.CreatedAt,
		"updated_at":       payment.UpdatedAt,
	}

	if payment.CompletedAt != nil {
		response["completed_at"] = payment.CompletedAt
	}

	if payment.FailedAt != nil {
		response["failed_at"] = payment.FailedAt
	}

	return response
}

func ShippingToResponse(shipping *entity.Shipping) map[string]interface{} {
	response := map[string]interface{}{
		"id":              shipping.ID,
		"order_id":        shipping.OrderID,
		"carrier":         shipping.Carrier,
		"tracking_number": shipping.TrackingNumber,
		"status":          shipping.Status,
		"shipping_method": shipping.ShippingMethod,
		"shipping_cost":   shipping.ShippingCost,
		"notes":           shipping.Notes,
		"created_at":      shipping.CreatedAt,
		"updated_at":      shipping.UpdatedAt,
	}

	if shipping.EstimatedDelivery != nil {
		response["estimated_delivery"] = shipping.EstimatedDelivery
	}

	if shipping.ShippedAt != nil {
		response["shipped_at"] = shipping.ShippedAt
	}

	if shipping.DeliveredAt != nil {
		response["delivered_at"] = shipping.DeliveredAt
	}

	return response
}

func CreatePaginatedResponse(orders []*entity.Order, total, page, pageSize int) map[string]interface{} {
	orderResponses := make([]map[string]interface{}, len(orders))
	for i, order := range orders {
		orderResponses[i] = OrderToResponse(order)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"orders":      orderResponses,
	}
}
