package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/product_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"gorm.io/gorm"
)

type Inventory struct {
	ID        string          `gorm:"primaryKey;type:char(36)" json:"id"`
	ProductID string          `gorm:"uniqueIndex;type:char(36);not null" json:"product_id"`
	Quantity  int             `gorm:"not null;default:0" json:"quantity"`
	Reserved  int             `gorm:"not null;default:0" json:"reserved"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (i *Inventory) TableName() string {
	return "inventories"
}

// ToEntity converts the GORM Inventory model to the domain entity Inventory
func (i *Inventory) ToEntity() *entity.Inventory {
	return &entity.Inventory{
		ID:        i.ID,
		ProductID: i.ProductID,
		Quantity:  i.Quantity,
		Reserved:  i.Reserved,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		DeletedAt: utils.DeletedAtPtrToTimePtr(i.DeletedAt),
	}
}

// NewInventoryModel creates a new GORM Inventory model from a domain entity Inventory
func NewInventoryModel(inventory *entity.Inventory) *Inventory {
	return &Inventory{
		ID:        inventory.ID,
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
		Reserved:  inventory.Reserved,
		CreatedAt: inventory.CreatedAt,
		UpdatedAt: inventory.UpdatedAt,
		DeletedAt: utils.TimePtrToDeletedAt(inventory.DeletedAt),
	}
}
