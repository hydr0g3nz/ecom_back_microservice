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

// CassandraShippingRepository implements the ShippingRepository interface using Cassandra
type CassandraShippingRepository struct {
	session *gocql.Session
}

// NewCassandraShippingRepository creates a new instance of CassandraShippingRepository
func NewCassandraShippingRepository(session *gocql.Session) *CassandraShippingRepository {
	return &CassandraShippingRepository{session: session}
}

// Create stores a new shipping record
func (r *CassandraShippingRepository) Create(ctx context.Context, shipping *entity.Shipping) (*entity.Shipping, error) {
	// Generate UUID if not provided
	if shipping.ID == "" {
		shipping.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	shipping.CreatedAt = now
	shipping.UpdatedAt = now

	// Convert to model
	shippingModel, err := model.FromShippingEntity(shipping)
	if err != nil {
		return nil, err
	}

	// Insert into shipping table
	query := `
		INSERT INTO shipping (
			id, order_id, carrier, tracking_number, status, estimated_delivery, 
			shipped_at, delivered_at, shipping_method, shipping_cost, notes, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err = r.session.Query(query,
		shippingModel.ID,
		shippingModel.OrderID,
		shippingModel.Carrier,
		shippingModel.TrackingNumber,
		shippingModel.Status,
		shippingModel.EstimatedDelivery,
		shippingModel.ShippedAt,
		shippingModel.DeliveredAt,
		shippingModel.ShippingMethod,
		shippingModel.ShippingCost,
		shippingModel.Notes,
		shippingModel.CreatedAt,
		shippingModel.UpdatedAt,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, err
	}

	return shipping, nil
}

// GetByID retrieves a shipping record by ID
func (r *CassandraShippingRepository) GetByID(ctx context.Context, id string) (*entity.Shipping, error) {
	shippingID, err := gocql.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	var shippingModel model.ShippingModel

	query := `
		SELECT id, order_id, carrier, tracking_number, status, estimated_delivery, 
		       shipped_at, delivered_at, shipping_method, shipping_cost, notes, 
		       created_at, updated_at
		FROM shipping
		WHERE id = ?
	`

	err = r.session.Query(query, shippingID).WithContext(ctx).Scan(
		&shippingModel.ID,
		&shippingModel.OrderID,
		&shippingModel.Carrier,
		&shippingModel.TrackingNumber,
		&shippingModel.Status,
		&shippingModel.EstimatedDelivery,
		&shippingModel.ShippedAt,
		&shippingModel.DeliveredAt,
		&shippingModel.ShippingMethod,
		&shippingModel.ShippingCost,
		&shippingModel.Notes,
		&shippingModel.CreatedAt,
		&shippingModel.UpdatedAt,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, entity.ErrShippingNotFound
		}
		return nil, err
	}

	return shippingModel.ToEntity()
}

// GetByOrderID retrieves shipping for an order
func (r *CassandraShippingRepository) GetByOrderID(ctx context.Context, orderID string) (*entity.Shipping, error) {
	id, err := gocql.ParseUUID(orderID)
	if err != nil {
		return nil, err
	}

	var shippingModel model.ShippingModel

	query := `
		SELECT id, order_id, carrier, tracking_number, status, estimated_delivery, 
		       shipped_at, delivered_at, shipping_method, shipping_cost, notes, 
		       created_at, updated_at
		FROM shipping
		WHERE order_id = ?
		LIMIT 1
		ALLOW FILTERING
	`

	err = r.session.Query(query, id).WithContext(ctx).Scan(
		&shippingModel.ID,
		&shippingModel.OrderID,
		&shippingModel.Carrier,
		&shippingModel.TrackingNumber,
		&shippingModel.Status,
		&shippingModel.EstimatedDelivery,
		&shippingModel.ShippedAt,
		&shippingModel.DeliveredAt,
		&shippingModel.ShippingMethod,
		&shippingModel.ShippingCost,
		&shippingModel.Notes,
		&shippingModel.CreatedAt,
		&shippingModel.UpdatedAt,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, entity.ErrShippingNotFound
		}
		return nil, err
	}

	return shippingModel.ToEntity()
}

// Update updates an existing shipping record
func (r *CassandraShippingRepository) Update(ctx context.Context, shipping *entity.Shipping) (*entity.Shipping, error) {
	// Set timestamp
	shipping.UpdatedAt = time.Now()

	// Convert to model
	shippingModel, err := model.FromShippingEntity(shipping)
	if err != nil {
		return nil, err
	}

	// Update shipping
	query := `
		UPDATE shipping
		SET carrier = ?, tracking_number = ?, status = ?, estimated_delivery = ?, 
		    shipped_at = ?, delivered_at = ?, shipping_method = ?, shipping_cost = ?, 
		    notes = ?, updated_at = ?
		WHERE id = ?
	`

	err = r.session.Query(query,
		shippingModel.Carrier,
		shippingModel.TrackingNumber,
		shippingModel.Status,
		shippingModel.EstimatedDelivery,
		shippingModel.ShippedAt,
		shippingModel.DeliveredAt,
		shippingModel.ShippingMethod,
		shippingModel.ShippingCost,
		shippingModel.Notes,
		shippingModel.UpdatedAt,
		shippingModel.ID,
	).WithContext(ctx).Exec()

	if err != nil {
		return nil, err
	}

	return shipping, nil
}

// UpdateStatus updates the status of a shipping record
func (r *CassandraShippingRepository) UpdateStatus(ctx context.Context, id string, status valueobject.ShippingStatus) error {
	// shippingID, err := gocql.ParseUUID(id)
	// if err != nil {
	// 	return err
	// }

	// Get current shipping record to update timestamps appropriately
	shipping, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update status and timestamps
	shipping.UpdateStatus(status)

	// Update the shipping record
	_, err = r.Update(ctx, shipping)
	return err
}

// UpdateTrackingInfo updates the tracking information for a shipment
func (r *CassandraShippingRepository) UpdateTrackingInfo(ctx context.Context, id string, carrier string, trackingNumber string) error {
	shippingID, err := gocql.ParseUUID(id)
	if err != nil {
		return err
	}

	now := time.Now()

	query := `
		UPDATE shipping
		SET carrier = ?, tracking_number = ?, updated_at = ?
		WHERE id = ?
	`

	err = r.session.Query(query,
		carrier,
		trackingNumber,
		now,
		shippingID,
	).WithContext(ctx).Exec()

	return err
}
