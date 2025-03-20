package entity

import "errors"

var (
	// Order errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrInvalidOrderData   = errors.New("invalid order data")
	ErrOrderCancelled     = errors.New("order is already cancelled")
	ErrOrderCompleted     = errors.New("order is already completed")
	ErrInvalidOrderStatus = errors.New("invalid order status transition")

	// Item errors
	ErrItemNotFound      = errors.New("item not found in order")
	ErrInvalidItemData   = errors.New("invalid item data")
	ErrInsufficientStock = errors.New("insufficient stock for item")

	// Payment errors
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists")
	ErrPaymentFailed        = errors.New("payment failed")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// Shipping errors
	ErrShippingNotFound      = errors.New("shipping not found")
	ErrShippingAlreadyExists = errors.New("shipping record already exists")
	ErrInvalidShippingData   = errors.New("invalid shipping data")

	// Event errors
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidEvent  = errors.New("invalid event data")

	// Generic errors
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrUnauthorized        = errors.New("unauthorized")
)
