// internal/order_service/adapter/controller/http/responsener.go
package httpctl

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

// HandleError builds an appropriate Fiber error response based on the domain error
func HandleError(c *fiber.Ctx, err error) error {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, entity.ErrOrderNotFound):
		statusCode = http.StatusNotFound
		message = "Order not found"
	case errors.Is(err, entity.ErrInvalidOrderData):
		statusCode = http.StatusBadRequest
		message = "Invalid order data"
	case errors.Is(err, entity.ErrInvalidOrderStatus):
		statusCode = http.StatusBadRequest
		message = "Invalid order status transition"
	case errors.Is(err, entity.ErrOrderCancelled):
		statusCode = http.StatusBadRequest
		message = "Order is already cancelled"
	case errors.Is(err, entity.ErrOrderCompleted):
		statusCode = http.StatusBadRequest
		message = "Order is already completed"
	case errors.Is(err, entity.ErrPaymentNotFound):
		statusCode = http.StatusNotFound
		message = "Payment not found"
	case errors.Is(err, entity.ErrPaymentFailed):
		statusCode = http.StatusBadRequest
		message = "Payment failed"
	case errors.Is(err, entity.ErrShippingNotFound):
		statusCode = http.StatusNotFound
		message = "Shipping not found"
	case errors.Is(err, entity.ErrItemNotFound):
		statusCode = http.StatusNotFound
		message = "Item not found in order"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = http.StatusBadRequest
		message = "Insufficient stock for item"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	default:
		statusCode = http.StatusInternalServerError
		message = "Something went wrong"
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"error":  message,
		"status": statusCode,
	})
}
