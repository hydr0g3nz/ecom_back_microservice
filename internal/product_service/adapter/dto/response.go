package dto

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
)

// ProductResponse represents a product response
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CategoryID  string    `json:"category_id"`
	ImageURL    string    `json:"image_url"`
	SKU         string    `json:"sku"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FromEntity converts a product entity to ProductResponse
func ProductResponseFromEntity(product *entity.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CategoryID:  product.CategoryID,
		ImageURL:    product.ImageURL,
		SKU:         product.SKU,
		Status:      product.Status,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

// CategoryResponse represents a category response
type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Level       int       `json:"level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FromEntity converts a category entity to CategoryResponse
func CategoryResponseFromEntity(category *entity.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ParentID:    category.ParentID,
		Level:       category.Level,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

// InventoryResponse represents an inventory response
type InventoryResponse struct {
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Reserved  int       `json:"reserved"`
	Available int       `json:"available"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromEntity converts an inventory entity to InventoryResponse
func InventoryResponseFromEntity(inventory *entity.Inventory) InventoryResponse {
	return InventoryResponse{
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
		Reserved:  inventory.Reserved,
		Available: inventory.Quantity - inventory.Reserved,
		UpdatedAt: inventory.UpdatedAt,
	}
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(total, page, pageSize int, data interface{}) PaginatedResponse {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return PaginatedResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Data:       data,
	}
}
