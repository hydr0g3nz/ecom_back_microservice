package cassandra

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
)

type contextKey string

const (
	txCtxKey      contextKey = "cassandra_transaction"
	batchSizeCtxKey contextKey = "batch_size"
)

// CassandraUnitOfWork implements the UnitOfWork interface for Cassandra
type CassandraUnitOfWork struct {
	session              *gocql.Session
	orderRepo            *CassandraOrderRepository
	orderReadRepo        *CassandraOrderReadRepository
	orderEventRepo       *CassandraOrderEventRepository
	paymentRepo          *CassandraPaymentRepository
	shippingRepo         *CassandraShippingRepository
	defaultBatchSize     int
}

// NewCassandraUnitOfWork creates a new instance of CassandraUnitOfWork
func NewCassandraUnitOfWork(
	session *gocql.Session,
	orderRepo *CassandraOrderRepository,
	orderReadRepo *CassandraOrderReadRepository,
	orderEventRepo *CassandraOrderEventRepository,
	paymentRepo *CassandraPaymentRepository,
	shippingRepo *CassandraShippingRepository,
	defaultBatchSize int,
) *CassandraUnitOfWork {
	if defaultBatchSize <= 0 {
		defaultBatchSize = 20 // Default batch size
	}

	return &CassandraUnitOfWork{
		session:              session,
		orderRepo:            orderRepo,
		orderReadRepo:        orderReadRepo,
		orderEventRepo:       orderEventRepo,
		paymentRepo:          paymentRepo,
		shippingRepo:         shippingRepo,
		defaultBatchSize:     defaultBatchSize,
	}
}

// Begin starts a new transaction
func (uow *CassandraUnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	// Cassandra doesn't have true transactions, so we use a batch to simulate one
	batch := uow.session.NewBatch(gocql.LoggedBatch)
	
	// Store the batch in the context
	ctx = context.WithValue(ctx, txCtxKey, batch)
	ctx = context.WithValue(ctx, batchSizeCtxKey, 0) // Initialize batch size counter
	
	return ctx, nil
}

// Commit commits the current transaction
func (uow *CassandraUnitOfWork) Commit(ctx context.Context) error {
	batch, ok := ctx.Value(txCtxKey).(*gocql.Batch)
	if !ok || batch == nil {
		return errors.New("no transaction in context")
	}
	
	batchSize, _ := ctx.Value(batchSizeCtxKey).(int)
	if batchSize == 0 {
		// No operations to commit
		return nil
	}
	
	// Execute the batch
	return uow.session.ExecuteBatch(batch)
}

// Rollback rolls back the current transaction
func (uow *CassandraUnitOfWork) Rollback(ctx context.Context) error {
	// Cassandra doesn't have transaction rollback, so we just discard the batch
	_, ok := ctx.Value(txCtxKey).(*gocql.Batch)
	if !ok {
		return errors.New("no transaction in context")
	}
	
	// Nothing to do for rollback
	return nil
}

// RunInTransaction runs the provided function in a transaction
func (uow *CassandraUnitOfWork) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Start a new transaction
	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			_ = uow.Rollback(txCtx)
			panic(r) // Re-throw the panic after rolling back
		}
	}()
	
	// Run the function
	if err := fn(txCtx); err != nil {
		// Rollback on error
		_ = uow.Rollback(txCtx)
		return err
	}
	
	// Commit the transaction
	if err := uow.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetOrderRepository returns the order repository for the current transaction
func (uow *CassandraUnitOfWork) GetOrderRepository(ctx context.Context) repository.OrderRepository {
	// For Cassandra, we return the same repository as it internally uses the batch from context if available
	return uow.orderRepo
}

// GetOrderReadRepository returns the order read repository for the current transaction
func (uow *CassandraUnitOfWork) GetOrderReadRepository(ctx context.Context) repository.OrderReadRepository {
	return uow.orderReadRepo
}

// GetOrderEventRepository returns the order event repository for the current transaction
func (uow *CassandraUnitOfWork) GetOrderEventRepository(ctx context.Context) repository.OrderEventRepository {
	return uow.orderEventRepo
}

// GetPaymentRepository returns the payment repository for the current transaction
func (uow *CassandraUnitOfWork) GetPaymentRepository(ctx context.Context) repository.PaymentRepository {
	return uow.paymentRepo
}

// GetShippingRepository returns the shipping repository for the current transaction
func (uow *CassandraUnitOfWork) GetShippingRepository(ctx context.Context) repository.ShippingRepository {
	return uow.shippingRepo
}

// AddToBatch adds a query to the current batch transaction
func (uow *CassandraUnitOfWork) AddToBatch(ctx context.Context, stmt string, args ...interface{}) error {
	batch, ok := ctx.Value(txCtxKey).(*gocql.Batch)
	if !ok || batch == nil {
		// No transaction, execute directly
		return uow.session.Query(stmt, args...).Exec()
	}
	
	// Add to batch
	batch.Query(stmt, args...)
	
	// Update batch size counter
	batchSize, _ := ctx.Value(batchSizeCtxKey).(int)
	ctx = context.WithValue(ctx, batchSizeCtxKey, batchSize+1)
	
	// Check if batch is getting too large and should be executed
	if batchSize+1 >= uow.defaultBatchSize {
		if err := uow.session.ExecuteBatch(batch); err != nil {
			return err
		}
		
		// Reset batch
		newBatch := uow.session.NewBatch(gocql.LoggedBatch)
		ctx = context.WithValue(ctx, txCtxKey, newBatch)
		ctx = context.WithValue(ctx, batchSizeCtxKey, 0)
	}
	
	return nil
}
