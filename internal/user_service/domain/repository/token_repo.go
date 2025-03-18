package repository

import (
	"context"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/user_service/domain/entity"
)

type TokenRepository interface {
	Create(ctx context.Context, token *entity.Token) error
	FindByToken(ctx context.Context, tokenStr string) (*entity.Token, error)
	// GetByUserID(ctx context.Context, userID string, tokenType entity.TokenType) (*entity.Token, error)
	Delete(ctx context.Context, tokenID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	Update(ctx context.Context, by entity.Token, token entity.Token) error
}
