package gormrepository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/adapter/repository/gorm/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
)

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) CreatePayment(ctx context.Context, payment *entity.Payment) error {
	model := model.NewPaymentModel(payment)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GormPaymentRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error) {
	var m model.Payment
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormPaymentRepository) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Payment, error) {
	var m model.Payment
	err := r.db.WithContext(ctx).First(&m, "order_id = ?", orderID).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormPaymentRepository) GetPaymentByGatewayTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error) {
	var m model.Payment
	err := r.db.WithContext(ctx).First(&m, "gateway_transaction_id = ?", transactionID).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormPaymentRepository) UpdatePayment(ctx context.Context, payment *entity.Payment) error {
	return r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", payment.ID).Updates(model.NewPaymentModel(payment)).Error
}

func (r *GormPaymentRepository) ListPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Payment, int, error) {
	var models []model.Payment
	var total int64
	tx := r.db.WithContext(ctx).Model(&model.Payment{}).Where("user_id = ?", userID)
	tx.Count(&total)
	err := tx.Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	entities := make([]*entity.Payment, len(models))
	for i, m := range models {
		entity, err := m.ToEntity()
		if err != nil {
			return nil, 0, err
		}
		entities[i] = entity
	}
	return entities, int(total), nil
}

// GormTransactionRepository
type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

func (r *GormTransactionRepository) CreateTransaction(ctx context.Context, t *entity.Transaction) error {
	return r.db.WithContext(ctx).Create(model.NewTransactionModel(t)).Error
}

func (r *GormTransactionRepository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var m model.Transaction
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormTransactionRepository) UpdateTransaction(ctx context.Context, t *entity.Transaction) error {
	return r.db.WithContext(ctx).Model(&model.Transaction{}).Where("id = ?", t.ID).Updates(model.NewTransactionModel(t)).Error
}

func (r *GormTransactionRepository) ListTransactionsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Transaction, error) {
	var models []model.Transaction
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	entities := make([]*entity.Transaction, len(models))
	for i, m := range models {
		entity, err := m.ToEntity()
		if err != nil {
			return nil, err
		}
		entities[i] = entity
	}
	return entities, nil
}

// GormPaymentMethodRepository
type GormPaymentMethodRepository struct {
	db *gorm.DB
}

func NewGormPaymentMethodRepository(db *gorm.DB) *GormPaymentMethodRepository {
	return &GormPaymentMethodRepository{db: db}
}

func (r *GormPaymentMethodRepository) CreatePaymentMethod(ctx context.Context, m *entity.PaymentMethod) error {
	return r.db.WithContext(ctx).Create(model.NewPaymentMethodModel(m)).Error
}

func (r *GormPaymentMethodRepository) GetPaymentMethodByID(ctx context.Context, id uuid.UUID) (*entity.PaymentMethod, error) {
	var m model.PaymentMethod
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *GormPaymentMethodRepository) UpdatePaymentMethod(ctx context.Context, m *entity.PaymentMethod) error {
	return r.db.WithContext(ctx).Model(&model.PaymentMethod{}).Where("id = ?", m.ID).Updates(model.NewPaymentMethodModel(m)).Error
}

func (r *GormPaymentMethodRepository) DeletePaymentMethod(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.PaymentMethod{}, "id = ?", id).Error
}

func (r *GormPaymentMethodRepository) ListPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PaymentMethod, error) {
	var models []model.PaymentMethod
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	entities := make([]*entity.PaymentMethod, len(models))
	for i, m := range models {
		entity, err := m.ToEntity()
		if err != nil {
			return nil, err
		}
		entities[i] = entity
	}
	return entities, nil
}

func (r *GormPaymentMethodRepository) GetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*entity.PaymentMethod, error) {
	var m model.PaymentMethod
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_default = true", userID).First(&m).Error
	if err != nil {
		return nil, err
	}
	entity, err := m.ToEntity()
	if err != nil {
		return nil, err
	}
	return entity, nil
}
