package cassandra

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// CassandraPaymentRepository implements the PaymentRepository interface using Cassandra
type CassandraPaymentRepository struct {
	session *gocql.Session
}

// NewCassandraPaymentRepository creates a new instance of CassandraPaymentRepository
func NewCassandraPaymentRepository(session *gocql.Session) *CassandraPaymentRepository {
	return &CassandraPaymentRepository{session: session}
}

// Create stores a new payment
func (r *CassandraPaymentRepository) Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error) {
	// Generate UUID if not provided
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	// Convert to model
	paymentModel, err := model.FromPaymentEntity(payment)
	if err != nil {
		return nil, err
	}

	// Insert into payments table
	query := `
		INSERT INTO payments (
			id, order_id, amount, currency, method, status, transaction_id, 
			gateway_response, created_at, updated_at, completed_at, failed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err = r.session.Query(query,
		paymentModel.ID,
		paymentModel.OrderID,
		paymentModel.Amount,
		paymentModel.Currency,
		paymentModel.Method,
		paymentModel.Status,
		paymentModel.TransactionID,
		paymentModel.GatewayResponse,
		paymentModel.CreatedAt,
		paymentModel.UpdatedAt,
		paymentModel.CompletedAt,
		paymentModel.FailedAt,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, err
	}

	return payment, nil
}

// GetByID retrieves a payment by ID
func (r *CassandraPaymentRepository) GetByID(ctx context.Context, id string) (*entity.Payment, error) {
	paymentID, err := gocql.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	var paymentModel model.PaymentModel

	query := `
		SELECT id, order_id, amount, currency, method, status, transaction_id, 
		       gateway_response, created_at, updated_at, completed_at, failed_at
		FROM payments
		WHERE id = ?
	`

	err = r.session.Query(query, paymentID).WithContext(ctx).Scan(
		&paymentModel.ID,
		&paymentModel.OrderID,
		&paymentModel.Amount,
		&paymentModel.Currency,
		&paymentModel.Method,
		&paymentModel.Status,
		&paymentModel.TransactionID,
		&paymentModel.GatewayResponse,
		&paymentModel.CreatedAt,
		&paymentModel.UpdatedAt,
		&paymentModel.CompletedAt,
		&paymentModel.FailedAt,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, entity.ErrPaymentNotFound
		}
		return nil, err
	}

	return paymentModel.ToEntity()
}

// GetByOrderID retrieves payments for an order
func (r *CassandraPaymentRepository) GetByOrderID(ctx context.Context, orderID string) ([]*entity.Payment, error) {
	id, err := gocql.ParseUUID(orderID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, order_id, amount, currency, method, status, transaction_id, 
		       gateway_response, created_at, updated_at, completed_at, failed_at
		FROM payments
		WHERE order_id = ?
		ALLOW FILTERING
	`

	iter := r.session.Query(query, id).WithContext(ctx).Iter()

	var payments []*entity.Payment
	var paymentModel model.PaymentModel

	for iter.Scan(
		&paymentModel.ID,
		&paymentModel.OrderID,
		&paymentModel.Amount,
		&paymentModel.Currency,
		&paymentModel.Method,
		&paymentModel.Status,
		&paymentModel.TransactionID,
		&paymentModel.GatewayResponse,
		&paymentModel.CreatedAt,
		&paymentModel.UpdatedAt,
		&paymentModel.CompletedAt,
		&paymentModel.FailedAt,
	) {
		payment, err := paymentModel.ToEntity()
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return payments, nil
}

// Update updates an existing payment
func (r *CassandraPaymentRepository) Update(ctx context.Context, payment *entity.Payment) (*entity.Payment, error) {
	// Set timestamp
	payment.UpdatedAt = time.Now()

	// Convert to model
	paymentModel, err := model.FromPaymentEntity(payment)
	if err != nil {
		return nil, err
	}

	// Update payment
	query := `
		UPDATE payments
		SET amount = ?, currency = ?, method = ?, status = ?, transaction_id = ?, 
		    gateway_response = ?, updated_at = ?, completed_at = ?, failed_at = ?
		WHERE id = ?
	`

	err = r.session.Query(query,
		paymentModel.Amount,
		paymentModel.Currency,
		paymentModel.Method,
		paymentModel.Status,
		paymentModel.TransactionID,
		paymentModel.GatewayResponse,
		paymentModel.UpdatedAt,
		paymentModel.CompletedAt,
		paymentModel.FailedAt,
		paymentModel.ID,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, err
	}

	return payment, nil
}

// UpdateStatus updates the status of a payment
func (r *CassandraPaymentRepository) UpdateStatus(ctx context.Context, id string, status valueobject.PaymentStatus) error {
	// paymentID, err := gocql.ParseUUID(id)
	// if err != nil {
	// 	return err
	// }

	// Get current payment to update timestamps appropriately
	payment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update status and timestamps
	payment.UpdateStatus(status)

	// Update the payment
	_, err = r.Update(ctx, payment)
	return err
}
