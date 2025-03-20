package cassandra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gocql/gocql"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// CassandraOrderReadRepository implements the OrderReadRepository interface using Cassandra
// This is a separate read model for CQRS
type CassandraOrderReadRepository struct {
	session *gocql.Session
}

// NewCassandraOrderReadRepository creates a new instance of CassandraOrderReadRepository
func NewCassandraOrderReadRepository(session *gocql.Session) *CassandraOrderReadRepository {
	return &CassandraOrderReadRepository{session: session}
}

// GetByID retrieves an order by ID (read model)
func (r *CassandraOrderReadRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	orderID, err := gocql.ParseUUID(id)
	if err != nil {
		return nil, err
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
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	return orderModel.ToEntity()
}

// GetByUserID retrieves orders for a user
func (r *CassandraOrderReadRepository) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*entity.Order, int, error) {
	// Query to get total count
	countQuery := `
		SELECT COUNT(*) FROM orders_by_user WHERE user_id = ?
	`

	var total int
	err := r.session.Query(countQuery, userID).WithContext(ctx).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// In Cassandra, we don't have OFFSET, so we would typically use LIMIT and pagination tokens
	// For simplicity, we'll just fetch all and filter in memory (not recommended for production)
	query := `
		SELECT order_id FROM orders_by_user WHERE user_id = ?
	`

	iter := r.session.Query(query, userID).WithContext(ctx).Iter()

	var orderIDs []gocql.UUID
	var orderID gocql.UUID

	for iter.Scan(&orderID) {
		orderIDs = append(orderIDs, orderID)
	}

	if err := iter.Close(); err != nil {
		return nil, 0, err
	}

	// Apply pagination manually
	if offset >= len(orderIDs) {
		return []*entity.Order{}, total, nil
	}

	end := offset + pageSize
	if end > len(orderIDs) {
		end = len(orderIDs)
	}

	paginatedOrderIDs := orderIDs[offset:end]

	// Fetch full order details for each ID
	var orders []*entity.Order

	for _, id := range paginatedOrderIDs {
		order, err := r.GetByID(ctx, id.String())
		if err != nil {
			continue // Skip orders with errors
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// FindByStatus retrieves orders with a specific status
func (r *CassandraOrderReadRepository) FindByStatus(ctx context.Context, status valueobject.OrderStatus, page, pageSize int) ([]*entity.Order, int, error) {
	// In Cassandra, we'd typically have a dedicated table or materialized view for this query
	// For this example, we'll use ALLOW FILTERING (not recommended for production)

	// Query to get total count
	countQuery := `
		SELECT COUNT(*) FROM orders WHERE status = ? ALLOW FILTERING
	`

	var total int
	err := r.session.Query(countQuery, status.String()).WithContext(ctx).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// In Cassandra, we don't have OFFSET, so we'll limit and manually skip
	query := `
		SELECT id FROM orders WHERE status = ? ALLOW FILTERING
	`

	iter := r.session.Query(query, status.String()).WithContext(ctx).Iter()

	var orderIDs []gocql.UUID
	var orderID gocql.UUID

	for iter.Scan(&orderID) {
		orderIDs = append(orderIDs, orderID)
	}

	if err := iter.Close(); err != nil {
		return nil, 0, err
	}

	// Apply pagination manually
	offset := (page - 1) * pageSize
	if offset >= len(orderIDs) {
		return []*entity.Order{}, total, nil
	}

	end := offset + pageSize
	if end > len(orderIDs) {
		end = len(orderIDs)
	}

	paginatedOrderIDs := orderIDs[offset:end]

	// Fetch full order details for each ID
	var orders []*entity.Order

	for _, id := range paginatedOrderIDs {
		order, err := r.GetByID(ctx, id.String())
		if err != nil {
			continue // Skip orders with errors
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// FindByDateRange retrieves orders created within a date range
func (r *CassandraOrderReadRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.Order, int, error) {
	// In Cassandra, this query is best served by a dedicated table with time-based partitioning
	// For this example, we'll use ALLOW FILTERING (not recommended for production)

	// Query to get total count
	countQuery := `
		SELECT COUNT(*) FROM orders WHERE created_at >= ? AND created_at <= ? ALLOW FILTERING
	`

	var total int
	err := r.session.Query(countQuery, startDate, endDate).WithContext(ctx).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// In Cassandra, we don't have OFFSET, so we'll limit and manually skip
	query := `
		SELECT id FROM orders WHERE created_at >= ? AND created_at <= ? ALLOW FILTERING
	`

	iter := r.session.Query(query, startDate, endDate).WithContext(ctx).Iter()

	var orderIDs []gocql.UUID
	var orderID gocql.UUID

	for iter.Scan(&orderID) {
		orderIDs = append(orderIDs, orderID)
	}

	if err := iter.Close(); err != nil {
		return nil, 0, err
	}

	// Apply pagination manually
	offset := (page - 1) * pageSize
	if offset >= len(orderIDs) {
		return []*entity.Order{}, total, nil
	}

	end := offset + pageSize
	if end > len(orderIDs) {
		end = len(orderIDs)
	}

	paginatedOrderIDs := orderIDs[offset:end]

	// Fetch full order details for each ID
	var orders []*entity.Order

	for _, id := range paginatedOrderIDs {
		order, err := r.GetByID(ctx, id.String())
		if err != nil {
			continue // Skip orders with errors
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// FindByProductID retrieves orders containing a specific product
func (r *CassandraOrderReadRepository) FindByProductID(ctx context.Context, productID string, page, pageSize int) ([]*entity.Order, int, error) {
	// In Cassandra, this would typically be implemented with a dedicated index table
	// For this example, we'll scan all orders (very inefficient, not for production)

	// Fetch all orders (in a real implementation, we'd have a better way)
	query := `SELECT id, items FROM orders`
	iter := r.session.Query(query).WithContext(ctx).Iter()

	var orderIDs []gocql.UUID
	var id gocql.UUID
	var itemsBytes []byte

	for iter.Scan(&id, &itemsBytes) {
		var items []entity.OrderItem
		if err := json.Unmarshal(itemsBytes, &items); err != nil {
			continue
		}

		// Check if order contains the product
		for _, item := range items {
			if item.ProductID == productID {
				orderIDs = append(orderIDs, id)
				break
			}
		}
	}

	if err := iter.Close(); err != nil {
		return nil, 0, err
	}

	total := len(orderIDs)

	// Calculate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Apply pagination manually
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*entity.Order{}, total, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	paginatedOrderIDs := orderIDs[offset:end]

	// Fetch full order details for each ID
	var orders []*entity.Order

	for _, id := range paginatedOrderIDs {
		order, err := r.GetByID(ctx, id.String())
		if err != nil {
			continue // Skip orders with errors
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// Search searches for orders based on various criteria
func (r *CassandraOrderReadRepository) Search(ctx context.Context, criteria map[string]interface{}, page, pageSize int) ([]*entity.Order, int, error) {
	// In Cassandra, complex search is difficult
	// For a real implementation, consider using a search engine like Elasticsearch
	// This is a simplified example that scans all orders (not for production)

	// Start with base query
	query := `SELECT id FROM orders`

	// Build where clauses (very limited support for Cassandra)
	// var conditions []string
	var params []interface{}

	whereAdded := false

	// Check for supported search criteria
	if status, ok := criteria["status"].(string); ok {
		if !whereAdded {
			query += " WHERE"
			whereAdded = true
		} else {
			query += " AND"
		}
		query += " status = ?"
		params = append(params, status)
	}

	if userId, ok := criteria["user_id"].(string); ok {
		if !whereAdded {
			query += " WHERE"
			whereAdded = true
		} else {
			query += " AND"
		}
		query += " user_id = ?"
		params = append(params, userId)
	}

	// Add ALLOW FILTERING for multiple conditions
	if len(params) > 0 {
		query += " ALLOW FILTERING"
	}

	var stmt *gocql.Query
	if len(params) > 0 {
		stmt = r.session.Query(query, params...)
	} else {
		stmt = r.session.Query(query)
	}

	iter := stmt.WithContext(ctx).Iter()

	var orderIDs []gocql.UUID
	var id gocql.UUID

	for iter.Scan(&id) {
		orderIDs = append(orderIDs, id)
	}

	if err := iter.Close(); err != nil {
		return nil, 0, err
	}

	// Apply post-filtering for criteria not supported by Cassandra
	// (e.g., product ID, amount ranges, etc.)
	// This would be done by fetching each order and filtering in-memory
	// Omitted for brevity

	total := len(orderIDs)

	// Calculate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Apply pagination manually
	offset := (page - 1) * pageSize
	if offset >= total {
		return []*entity.Order{}, total, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	paginatedOrderIDs := orderIDs[offset:end]

	// Fetch full order details for each ID
	var orders []*entity.Order

	for _, id := range paginatedOrderIDs {
		order, err := r.GetByID(ctx, id.String())
		if err != nil {
			continue // Skip orders with errors
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}
