package entity

import (
	"time"

	"github.com/google/uuid"
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/valueobject"
	"gorm.io/gorm"
)

type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Role           vo.Role   `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return nil
}
