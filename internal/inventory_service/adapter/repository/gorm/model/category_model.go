package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"gorm.io/gorm"
)

type Category struct {
	ID          string          `gorm:"primaryKey;type:char(36)" json:"id"`
	Name        string          `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	ParentID    *string         `gorm:"index;type:char(36)" json:"parent_id"`
	Level       int             `gorm:"not null;default:1" json:"level"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (c *Category) TableName() string {
	return "categories"
}

// ToEntity converts the GORM Category model to the domain entity Category
func (c *Category) ToEntity() *entity.Category {
	return &entity.Category{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		Level:       c.Level,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		DeletedAt:   utils.DeletedAtPtrToTimePtr(c.DeletedAt),
	}
}

// NewCategoryModel creates a new GORM Category model from a domain entity Category
func NewCategoryModel(category *entity.Category) *Category {
	return &Category{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ParentID:    category.ParentID,
		Level:       category.Level,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		DeletedAt:   utils.TimePtrToDeletedAt(category.DeletedAt),
	}
}
