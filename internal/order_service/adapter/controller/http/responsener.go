// internal/order_service/adapter/controller/http/responsener.go
package httpctl

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
)

var (
	ErrBadRequest = errors.New("bad request")
)

type successResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// SuccessResp builds a success response
func SuccessResp(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(successResponse{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

// HandleError builds an appropriate Fiber error response based on the domain error
func HandleError(c *fiber.Ctx, err error) error {
	var statusCode int
	var message string

	switch {
	case errors.Is(err, ErrBadRequest):
		statusCode = http.StatusBadRequest
		message = "Bad request"
	case errors.Is(err, entity.ErrOrderNotFound):
		statusCode = http.StatusNotFound
		message = "Order not found"
	case errors.Is(err, entity.ErrInvalidOrderData):
		statusCode = http.StatusBadRequest
		message = "Invalid order data"
	case errors.Is(err, entity.ErrInvalidOrderStatus):
		statusCode = http.StatusBadRequest
		message = "Invalid order status"
	case errors.Is(err, entity.ErrOrderAlreadyExists):
		statusCode = http.StatusConflict
		message = "Order already exists"
	case errors.Is(err, entity.ErrInvalidStatusTransition):
		statusCode = http.StatusBadRequest
		message = "Invalid status transition"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = http.StatusBadRequest
		message = "Insufficient stock"
	case errors.Is(err, entity.ErrPaymentFailed):
		statusCode = http.StatusBadRequest
		message = "Payment failed"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	default:
		statusCode = http.StatusInternalServerError
		message = "Something went wrong"
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Status:  statusCode,
		Message: message,
	})
}
