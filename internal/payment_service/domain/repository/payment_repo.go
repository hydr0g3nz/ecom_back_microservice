package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
)

// PaymentRepository เป็น interface สำหรับการเข้าถึงข้อมูล Payment
type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *entity.Payment) error
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Payment, error)
	GetPaymentByGatewayTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error)
	UpdatePayment(ctx context.Context, payment *entity.Payment) error
	ListPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Payment, int, error)
}

// TransactionRepository เป็น interface สำหรับการเข้าถึงข้อมูล Transaction
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *entity.Transaction) error
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *entity.Transaction) error
	ListTransactionsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Transaction, error)
}

// PaymentMethodRepository เป็น interface สำหรับการเข้าถึงข้อมูล PaymentMethod
type PaymentMethodRepository interface {
	CreatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	GetPaymentMethodByID(ctx context.Context, id uuid.UUID) (*entity.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, id uuid.UUID) error
	ListPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PaymentMethod, error)
	GetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*entity.PaymentMethod, error)
}
