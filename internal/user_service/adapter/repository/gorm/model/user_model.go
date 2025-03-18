package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"gorm.io/gorm"
)

type User struct {
	ID             string          `gorm:"primaryKey;type:char(36)" json:"id"`
	Email          string          `gorm:"uniqueIndex;not null;size:100" json:"email"`
	HashedPassword string          `gorm:"not null" json:"-"`
	FirstName      string          `gorm:"size:100" json:"first_name"`
	LastName       string          `gorm:"size:100" json:"last_name"`
	Role           vo.Role         `gorm:"type:varchar(20)" json:"role"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (u *User) TableName() string {
	return "users"
}

// ToEntity converts the GORM User model to the domain entity User
func (u *User) ToEntity() *entity.User {
	return &entity.User{
		ID:             u.ID,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Role:           u.Role,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		DeletedAt:      utils.DeletedAtPtrToTimePtr(u.DeletedAt),
	}
}
func NewUserModel(user *entity.User) *User {
	return &User{
		ID:             user.ID,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Role:           user.Role,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		DeletedAt:      utils.TimePtrToDeletedAt(user.DeletedAt),
	}
}
