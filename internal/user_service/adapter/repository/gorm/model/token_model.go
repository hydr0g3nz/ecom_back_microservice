package model

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
	"gorm.io/gorm"
)

type Token struct {
	ID        string           `gorm:"primaryKey;type:char(36)" json:"id"`
	UserID    string           `gorm:"not null;index;type:char(36)" json:"user_id"`
	Token     string           `gorm:"not null;type:text" json:"token"`
	Type      entity.TokenType `gorm:"not null;type:varchar(20)" json:"type"`
	ExpiresAt time.Time        `gorm:"not null;autoUpdateTime" json:"expires_at"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *gorm.DeletedAt  `gorm:"index" json:"deleted_at"`
}

func (t *Token) TableName() string {
	return "tokens"
}
func NewTokenModel(token *entity.Token) *Token {
	return &Token{
		ID:        token.ID,
		UserID:    token.UserID,
		Token:     token.Token,
		Type:      token.Type,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		UpdatedAt: token.UpdatedAt,
		DeletedAt: &gorm.DeletedAt{Time: utils.ValueOr(token.DeletedAt)},
	}
}

func (t *Token) ToEntity() *entity.Token {
	return &entity.Token{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		Type:      t.Type,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: &t.DeletedAt.Time,
	}
}
