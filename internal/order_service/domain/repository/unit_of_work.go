package repository

import "context"

// UnitOfWork defines the interface for transaction management
type UnitOfWork interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the current transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the current transaction
	Rollback(ctx context.Context) error

	// RunInTransaction runs the provided function in a transaction
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// GetOrderRepository returns the order repository for the current transaction
	GetOrderRepository(ctx context.Context) OrderRepository

	// GetOrderReadRepository returns the order read repository for the current transaction
	GetOrderReadRepository(ctx context.Context) OrderReadRepository

	// GetOrderEventRepository returns the order event repository for the current transaction
	GetOrderEventRepository(ctx context.Context) OrderEventRepository

	// GetPaymentRepository returns the payment repository for the current transaction
	GetPaymentRepository(ctx context.Context) PaymentRepository

	// GetShippingRepository returns the shipping repository for the current transaction
	GetShippingRepository(ctx context.Context) ShippingRepository
}
