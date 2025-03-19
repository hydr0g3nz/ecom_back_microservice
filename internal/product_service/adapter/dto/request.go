package dto

import (
	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
)

// ProductRequest represents a request to create or update a product
type ProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	CategoryID  string  `json:"category_id" validate:"required"`
	ImageURL    string  `json:"image_url"`
	SKU         string  `json:"sku" validate:"required"`
	Status      string  `json:"status"`
}

// ToEntity converts ProductRequest to a product entity
func (p ProductRequest) ToEntity() entity.Product {
	return entity.Product{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		SKU:         p.SKU,
		Status:      p.Status,
	}
}

// CategoryRequest represents a request to create or update a category
type CategoryRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// ToEntity converts CategoryRequest to a category entity
func (c CategoryRequest) ToEntity() entity.Category {
	return entity.Category{
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
	}
}

// InventoryUpdateRequest represents a request to update product inventory
type InventoryUpdateRequest struct {
	Quantity int `json:"quantity" validate:"gte=0"`
}

// ReservationRequest represents a request to reserve inventory
type ReservationRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
