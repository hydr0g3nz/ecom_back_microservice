package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
)

// OrderController handles order-related HTTP requests
type OrderController struct {
	orderUseCase *usecase.OrderUseCase
	responsener  *Responsener
}

// NewOrderController creates a new order controller
func NewOrderController(orderUseCase *usecase.OrderUseCase) *OrderController {
	return &OrderController{
		orderUseCase: orderUseCase,
		responsener:  NewResponsner(),
	}
}

// RegisterRoutes registers the routes for the order controller
func (c *OrderController) RegisterRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.Post("/", c.CreateOrder)
		r.Get("/", c.ListOrders)
		r.Get("/{id}", c.GetOrder)
		r.Put("/{id}/status", c.UpdateOrderStatus)
		r.Post("/{id}/payment", c.AddPayment)
		r.Delete("/{id}", c.CancelOrder)
		r.Get("/user/{user_id}", c.GetUserOrders)
	})
}

// CreateOrder handles the creation of a new order
func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responsener.BadRequest(w, "Invalid request body")
		return
	}
	
	if err := req.Validate(); err != nil {
		c.responsener.BadRequest(w, err.Error())
		return
	}
	
	order, err := c.orderUseCase.CreateOrder(r.Context(), req.UserID, req.ToOrderItems(), req.ShippingAddress)
	if err != nil {
		switch err {
		case entity.ErrInvalidUserID, entity.ErrEmptyOrderItems:
			c.responsener.BadRequest(w, err.Error())
		case entity.ErrOrderAlreadyExists:
			c.responsener.UnprocessableEntity(w, err.Error())
		default:
			c.responsener.InternalServerError(w, "Failed to create order")
		}
		return
	}
	
	c.responsener.Created(w, dto.NewOrderResponse(order))
}

// GetOrder gets an order by ID
func (c *OrderController) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		c.responsener.BadRequest(w, "Order ID is required")
		return
	}
	
	order, err := c.orderUseCase.GetOrderByID(r.Context(), id)
	if err != nil {
		if err == entity.ErrOrderNotFound {
			c.responsener.NotFound(w, "Order not found")
		} else {
			c.responsener.InternalServerError(w, "Failed to get order")
		}
		return
	}
	
	c.responsener.JSON(w, http.StatusOK, dto.NewOrderResponse(order))
}

// UpdateOrderStatus updates the status of an order
func (c *OrderController) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		c.responsener.BadRequest(w, "Order ID is required")
		return
	}
	
	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responsener.BadRequest(w, "Invalid request body")
		return
	}
	
	if err := req.Validate(); err != nil {
		c.responsener.BadRequest(w, err.Error())
		return
	}
	
	err := c.orderUseCase.UpdateOrderStatus(r.Context(), id, valueobject.OrderStatus(req.Status))
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			c.responsener.NotFound(w, "Order not found")
		case entity.ErrInvalidOrderStatus:
			c.responsener.BadRequest(w, "Invalid order status")
		default:
			c.responsener.InternalServerError(w, "Failed to update order status")
		}
		return
	}
	
	c.responsener.Success(w, "Order status updated")
}

// AddPayment adds payment information to an order
func (c *OrderController) AddPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		c.responsener.BadRequest(w, "Order ID is required")
		return
	}
	
	var req dto.OrderPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.responsener.BadRequest(w, "Invalid request body")
		return
	}
	
	if err := req.Validate(); err != nil {
		c.responsener.BadRequest(w, err.Error())
		return
	}
	
	// Ensure the order ID in the URL matches the one in the request
	if req.OrderID != "" && req.OrderID != id {
		c.responsener.BadRequest(w, "Order ID mismatch")
		return
	}
	
	err := c.orderUseCase.AddPaymentToOrder(r.Context(), id, req.PaymentID)
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			c.responsener.NotFound(w, "Order not found")
		default:
			c.responsener.InternalServerError(w, "Failed to process payment")
		}
		return
	}
	
	c.responsener.Success(w, "Payment processed successfully")
}

// CancelOrder cancels an order
func (c *OrderController) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		c.responsener.BadRequest(w, "Order ID is required")
		return
	}
	
	err := c.orderUseCase.CancelOrder(r.Context(), id)
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			c.responsener.NotFound(w, "Order not found")
		default:
			c.responsener.InternalServerError(w, "Failed to cancel order")
		}
		return
	}
	
	c.responsener.Success(w, "Order cancelled")
}

// GetUserOrders gets all orders for a user
func (c *OrderController) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		c.responsener.BadRequest(w, "User ID is required")
		return
	}
	
	orders, err := c.orderUseCase.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		c.responsener.InternalServerError(w, "Failed to get user orders")
		return
	}
	
	// Convert to response DTOs
	responses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		response := dto.OrderResponse{}
		response.FromEntity(order)
		responses[i] = response
	}
	
	c.responsener.JSON(w, http.StatusOK, map[string]interface{}{
		"orders": responses,
		"count":  len(responses),
	})
}

// ListOrders lists orders with pagination and filtering
func (c *OrderController) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	
	limit := 10 // Default limit
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	
	var orders []*entity.Order
	var totalCount int
	var err error
	
	if status != "" {
		// Filter by status if provided
		orderStatus := valueobject.OrderStatus(status)
		if !orderStatus.IsValid() {
			c.responsener.BadRequest(w, "Invalid order status")
			return
		}
		
		orders, err = c.orderUseCase.ListOrdersByStatus(r.Context(), orderStatus)
		totalCount = len(orders)
		
		// Apply pagination manually (not ideal but works for now)
		if len(orders) > 0 {
			start := (page - 1) * limit
			end := start + limit
			
			if start >= len(orders) {
				orders = []*entity.Order{}
			} else if end > len(orders) {
				orders = orders[start:]
			} else {
				orders = orders[start:end]
			}
		}
	} else {
		// Get paginated orders
		orders, totalCount, err = c.orderUseCase.GetOrdersPaginated(r.Context(), page, limit)
	}
	
	if err != nil {
		c.responsener.InternalServerError(w, "Failed to list orders")
		return
	}
	
	// Create response
	response := dto.NewOrdersResponse(orders, totalCount, page, limit)
	c.responsener.JSON(w, http.StatusOK, response)
}
