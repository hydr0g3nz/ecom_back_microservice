package entity

import "errors"

var (
	// Product errors
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrProductSKUExists     = errors.New("product SKU already exists")
	ErrInvalidProductData   = errors.New("invalid product data")

	// Category errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrInvalidCategoryData   = errors.New("invalid category data")

	// Inventory errors
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInsufficientStock = errors.New("insufficient stock")

	// Generic errors
	ErrInternalServerError = errors.New("internal server error")
)
