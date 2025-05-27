package httpctl

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"gorm.io/gorm"
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
	case errors.Is(err, entity.ErrProductNotFound):
		statusCode = http.StatusNotFound
		message = "Product not found"
	case errors.Is(err, entity.ErrCategoryNotFound):
		statusCode = http.StatusNotFound
		message = "Category not found"
	case errors.Is(err, entity.ErrInventoryNotFound):
		statusCode = http.StatusNotFound
		message = "Inventory not found"
	case errors.Is(err, entity.ErrProductSKUExists):
		statusCode = http.StatusConflict
		message = "Product SKU already exists"
	case errors.Is(err, entity.ErrCategoryAlreadyExists):
		statusCode = http.StatusConflict
		message = "Category already exists"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = http.StatusBadRequest
		message = "Insufficient stock"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	case errors.Is(err, gorm.ErrRecordNotFound):
		statusCode = http.StatusNotFound
		message = "Record not found"
	default:
		statusCode = http.StatusInternalServerError
		message = "Something went wrong"
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Status:  statusCode,
		Message: message,
	})
}
