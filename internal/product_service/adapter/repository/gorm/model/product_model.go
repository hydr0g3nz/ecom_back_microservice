package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"gorm.io/gorm"
)

type Product struct {
	ID          string          `gorm:"primaryKey;type:char(36)" json:"id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Price       float64         `gorm:"not null" json:"price"`
	CategoryID  string          `gorm:"index;type:char(36)" json:"category_id"`
	ImageURL    string          `gorm:"size:255" json:"image_url"`
	SKU         string          `gorm:"uniqueIndex;size:100;not null" json:"sku"`
	Status      string          `gorm:"size:20;not null" json:"status"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (p *Product) TableName() string {
	return "products"
}

// ToEntity converts the GORM Product model to the domain entity Product
func (p *Product) ToEntity() *entity.Product {
	return &entity.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		SKU:         p.SKU,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   utils.DeletedAtPtrToTimePtr(p.DeletedAt),
	}
}

// NewProductModel creates a new GORM Product model from a domain entity Product
func NewProductModel(product *entity.Product) *Product {
	return &Product{
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
		DeletedAt:   utils.TimePtrToDeletedAt(product.DeletedAt),
	}
}
