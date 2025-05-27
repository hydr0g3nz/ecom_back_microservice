package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
	"github.com/shopspring/decimal"
)

// PaymentGateway ระบุเมธอดที่จำเป็นสำหรับการทำธุรกรรมกับ payment gateway
type PaymentGateway interface {
	ProcessPayment(ctx context.Context, amount decimal.Decimal, tokenizedData string, orderID uuid.UUID) (*entity.GatewayResponse, error)
	VerifyCallback(ctx context.Context, callbackData map[string]interface{}) (bool, error)
	ProcessRefund(ctx context.Context, transactionID string, amount decimal.Decimal) (*entity.GatewayResponse, error)
}
