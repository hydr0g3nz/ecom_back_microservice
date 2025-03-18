package httpctl

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
)

var (
	ErrBadRequest = errors.New("bad request")
)

type successResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewSuccessResponse builds a success response
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
	case errors.Is(err, entity.ErrUserNotFound):
		statusCode = http.StatusNotFound
		message = "User not found"
	case errors.Is(err, entity.ErrUserAlreadyExists) || errors.Is(err, entity.ErrUserExists):
		statusCode = http.StatusConflict
		message = "User already exists"
	case errors.Is(err, entity.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		message = "Invalid credentials"
	case errors.Is(err, entity.ErrInvalidToken) || errors.Is(err, entity.ErrTokenHasBeenRevoked):
		statusCode = http.StatusUnauthorized
		message = "Invalid or revoked token"
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
