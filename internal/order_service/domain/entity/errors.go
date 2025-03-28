package entity

import "errors"

// Domain error definitions
var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrInvalidOrderStatus is returned when an invalid order status is provided
	ErrInvalidOrderStatus = errors.New("invalid order status")

	// ErrInvalidOrderInput is returned when invalid order input is provided
	ErrInvalidOrderInput = errors.New("invalid order input")
	
	// ErrOrderAlreadyExists is returned when trying to create an order that already exists
	ErrOrderAlreadyExists = errors.New("order already exists")
	
	// ErrEmptyOrderItems is returned when an order has no items
	ErrEmptyOrderItems = errors.New("order must have at least one item")
	
	// ErrInvalidUserID is returned when an invalid user ID is provided
	ErrInvalidUserID = errors.New("invalid user ID")
	
	// ErrInvalidProductID is returned when an invalid product ID is provided
	ErrInvalidProductID = errors.New("invalid product ID")
	
	// ErrInsufficientStock is returned when there is insufficient stock for a product
	ErrInsufficientStock = errors.New("insufficient stock for product")
	
	// ErrPaymentFailed is returned when a payment fails
	ErrPaymentFailed = errors.New("payment failed")
)
