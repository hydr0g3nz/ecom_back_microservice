package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/utils"
)

// CassandraOrderRepository implements the OrderRepository interface using Cassandra
type CassandraOrderRepository struct {
	session    *gocql.Session
	errBuilder *utils.ErrorBuilder
}

// NewCassandraOrderRepository creates a new instance of CassandraOrderRepository
func NewCassandraOrderRepository(session *gocql.Session) *CassandraOrderRepository {
	return &CassandraOrderRepository{
		session:    session,
		errBuilder: utils.NewErrorBuilder("CassandraOrderRepository"),
	}
}

// Create stores a new order
func (r *CassandraOrderRepository) Create(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	// Generate UUID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Version = 1

	// Convert to model
	orderModel, err := model.FromOrderEntity(order)
	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	// Insert into orders table
	query := `
		INSERT INTO orders (
			id, user_id, items, total_amount, status, shipping_address, billing_address, 
			payment_id, shipping_id, notes, promotion_codes, discounts, tax_amount,
			created_at, updated_at, completed_at, cancelled_at, version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err = r.session.Query(query,
		orderModel.ID,
		orderModel.UserID,
		orderModel.Items,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.ShippingAddress,
		orderModel.BillingAddress,
		orderModel.PaymentID,
		orderModel.ShippingID,
		orderModel.Notes,
		orderModel.PromotionCodes,
		orderModel.Discounts,
		orderModel.TaxAmount,
		orderModel.CreatedAt,
		orderModel.UpdatedAt,
		orderModel.CompletedAt,
		orderModel.CancelledAt,
		orderModel.Version,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	// Insert into orders_by_user table for user-based lookups
	userQuery := `
		INSERT INTO orders_by_user (
			user_id, order_id, status, total_amount, created_at
		) VALUES (?, ?, ?, ?, ?)
	`

	err = r.session.Query(userQuery,
		orderModel.UserID,
		orderModel.ID,
		orderModel.Status,
		orderModel.TotalAmount,
		orderModel.CreatedAt,
	).WithContext(ctx).Exec()

	if err != nil {
		// Log partial failure but don't fail the operation
		return nil, r.errBuilder.Err(err)
	}

	return order, nil
}

// GetByID retrieves an order by ID
func (r *CassandraOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	orderID, err := gocql.ParseUUID(id)
	if err != nil {
		return nil, r.errBuilder.Err(entity.ErrInvalidOrderData)
	}

	var orderModel model.OrderModel

	query := `
		SELECT id, user_id, items, total_amount, status, shipping_address, billing_address, 
		       payment_id, shipping_id, notes, promotion_codes, discounts, tax_amount,
		       created_at, updated_at, completed_at, cancelled_at, version
		FROM orders
		WHERE id = ?
	`

	err = r.session.Query(query, orderID).WithContext(ctx).Scan(
		&orderModel.ID,
		&orderModel.UserID,
		&orderModel.Items,
		&orderModel.TotalAmount,
		&orderModel.Status,
		&orderModel.ShippingAddress,
		&orderModel.BillingAddress,
		&orderModel.PaymentID,
		&orderModel.ShippingID,
		&orderModel.Notes,
		&orderModel.PromotionCodes,
		&orderModel.Discounts,
		&orderModel.TaxAmount,
		&orderModel.CreatedAt,
		&orderModel.UpdatedAt,
		&orderModel.CompletedAt,
		&orderModel.CancelledAt,
		&orderModel.Version,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, r.errBuilder.Err(entity.ErrOrderNotFound)
		}
		return nil, r.errBuilder.Err(err)
	}

	orderEntity, err := orderModel.ToEntity()
	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	return orderEntity, nil
}

// Update updates an existing order
func (r *CassandraOrderRepository) Update(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	// Get current version for optimistic locking
	currentOrder, err := r.GetByID(ctx, order.ID)
	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	// Set timestamps and increment version
	order.UpdatedAt = time.Now()
	order.Version = currentOrder.Version + 1

	// Convert to model
	orderModel, err := model.FromOrderEntity(order)
	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	// Update with optimistic locking
	query := `
		UPDATE orders
		SET user_id = ?, items = ?, total_amount = ?, status = ?, 
		    shipping_address = ?, billing_address = ?, payment_id = ?, 
		    shipping_id = ?, notes = ?, promotion_codes = ?, 
		    discounts = ?, tax_amount = ?, updated_at = ?, 
		    completed_at = ?, cancelled_at = ?, version = ?
		WHERE id = ?
		IF version = ?
	`

	applied := false
	_, err = r.session.Query(query,
		orderModel.UserID,
		orderModel.Items,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.ShippingAddress,
		orderModel.BillingAddress,
		orderModel.PaymentID,
		orderModel.ShippingID,
		orderModel.Notes,
		orderModel.PromotionCodes,
		orderModel.Discounts,
		orderModel.TaxAmount,
		orderModel.UpdatedAt,
		orderModel.CompletedAt,
		orderModel.CancelledAt,
		orderModel.Version,
		orderModel.ID,
		currentOrder.Version,
	).WithContext(ctx).ScanCAS(&applied)

	if err != nil {
		return nil, r.errBuilder.Err(err)
	}

	if !applied {
		return nil, r.errBuilder.Err(entity.ErrInvalidOrderStatus)
	}

	// Update the orders_by_user table
	userQuery := `
		UPDATE orders_by_user
		SET status = ?, total_amount = ?
		WHERE user_id = ? AND order_id = ?
	`

	err = r.session.Query(userQuery,
		orderModel.Status,
		orderModel.TotalAmount,
		orderModel.UserID,
		orderModel.ID,
	).WithContext(ctx).Exec()

	if err != nil {
		// Log this error but don't fail the operation
	}

	return order, nil
}

// UpdateStatus updates the status of an order
func (r *CassandraOrderRepository) UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus) error {
	orderID, err := gocql.ParseUUID(id)
	if err != nil {
		return r.errBuilder.Err(entity.ErrInvalidOrderData)
	}

	// Get current order to check for valid status transition and optimistic locking
	currentOrder, err := r.GetByID(ctx, id)
	if err != nil {
		return r.errBuilder.Err(err)
	}

	// Verify valid status transition
	if !valueobject.IsValidTransition(currentOrder.Status, status) {
		fmt.Println("status ", status, "currentOrder.Status", currentOrder.Status)
		return r.errBuilder.Err(entity.ErrInvalidOrderStatus)
	}

	now := time.Now()
	newVersion := currentOrder.Version + 1

	// Set timestamps based on status
	var completedAt, cancelledAt *time.Time
	if status == valueobject.Completed {
		completedAt = &now
	} else if status == valueobject.Cancelled {
		cancelledAt = &now
	}

	// Update with optimistic locking
	query := `
		UPDATE orders
		SET status = ?, updated_at = ?, completed_at = ?, cancelled_at = ?, version = ?
		WHERE id = ?
		IF version = ?
	`
	fmt.Println("version", currentOrder.Version, "newVersion", newVersion)
	// applied := false
	applied, err := r.session.Query(query,
		status.String(),
		now,
		completedAt,
		cancelledAt,
		newVersion,
		orderID,
		currentOrder.Version,
	).WithContext(ctx).ScanCAS()

	if err != nil {
		return r.errBuilder.Err(err)
	}

	if !applied {
		fmt.Println("Error updating orders table:", err)
		return r.errBuilder.Err(entity.ErrInvalidOrderStatus)
	}

	// Update the orders_by_user table
	userQuery := `
		UPDATE orders_by_user
		SET status = ?
		WHERE user_id = ? AND order_id = ?
	`

	err = r.session.Query(userQuery,
		status.String(),
		currentOrder.UserID,
		orderID,
	).WithContext(ctx).Exec()

	if err != nil {
		// Log this error but don't fail the operation
		fmt.Println("Error updating orders_by_user table:", err)
	}

	return nil
}

// AddItem adds an item to an order
func (r *CassandraOrderRepository) AddItem(ctx context.Context, orderID string, item entity.OrderItem) error {
	// Get current order
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return r.errBuilder.Err(err)
	}

	// Add the item
	order.AddItem(item)

	// Update the order
	_, err = r.Update(ctx, order)
	return r.errBuilder.Err(err)
}

// RemoveItem removes an item from an order
func (r *CassandraOrderRepository) RemoveItem(ctx context.Context, orderID string, productID string) error {
	// Get current order
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return r.errBuilder.Err(err)
	}

	// Remove the item
	if !order.RemoveItem(productID) {
		return r.errBuilder.Err(entity.ErrItemNotFound)
	}

	// Update the order
	_, err = r.Update(ctx, order)
	return r.errBuilder.Err(err)
}

// UpdateItemQuantity updates the quantity of an item in an order
func (r *CassandraOrderRepository) UpdateItemQuantity(ctx context.Context, orderID string, productID string, quantity int) error {
	// Get current order
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return r.errBuilder.Err(err)
	}

	// Update item quantity
	if !order.UpdateItemQuantity(productID, quantity) {
		return r.errBuilder.Err(entity.ErrItemNotFound)
	}

	// Update the order
	_, err = r.Update(ctx, order)
	return r.errBuilder.Err(err)
}

// ApplyDiscount applies a discount to an order
func (r *CassandraOrderRepository) ApplyDiscount(ctx context.Context, orderID string, discount entity.Discount) error {
	// Get current order
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return r.errBuilder.Err(err)
	}

	// Apply the discount
	order.ApplyDiscount(discount)

	// Update the order
	_, err = r.Update(ctx, order)
	return r.errBuilder.Err(err)
}
