// internal/order_service/domain/entity/errors.go
package entity

import "errors"

var (
	// Order errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderData   = errors.New("invalid order data")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrOrderAlreadyExists = errors.New("order already exists")

	// Inventory errors
	ErrInsufficientStock = errors.New("insufficient stock")

	// Payment errors
	ErrPaymentFailed = errors.New("payment failed")

	// State transition errors
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// Generic errors
	ErrInternalServerError = errors.New("internal server error")
)
