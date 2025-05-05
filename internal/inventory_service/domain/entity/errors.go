// internal/inventory_service/domain/entity/errors.go
package entity

import "errors"

var (
	ErrInventoryNotFound  = errors.New("inventory item not found")
	ErrInsufficientStock  = errors.New("insufficient stock for reservation")
	ErrInvalidProductData = errors.New("invalid product data")
	ErrSKUAlreadyExists   = errors.New("SKU already exists")
	// Add other domain-specific errors here
)

// Re-declare user service errors if they are needed for shared HandleError function
// Or, ideally, HandleError should be more generic or separated.
// For this example, let's assume ErrInventoryNotFound and ErrInsufficientStock
// are the main new errors to handle explicitly.
// If you need user errors in a shared handler, you'd import the user service entity errors package.
