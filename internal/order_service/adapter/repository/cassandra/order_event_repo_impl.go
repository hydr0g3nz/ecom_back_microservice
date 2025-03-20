package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/cassandra/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// CassandraOrderEventRepository implements the OrderEventRepository interface using Cassandra
type CassandraOrderEventRepository struct {
	session *gocql.Session
}

// NewCassandraOrderEventRepository creates a new instance of CassandraOrderEventRepository
func NewCassandraOrderEventRepository(session *gocql.Session) *CassandraOrderEventRepository {
	return &CassandraOrderEventRepository{session: session}
}

// SaveEvent stores a new order event
func (r *CassandraOrderEventRepository) SaveEvent(ctx context.Context, event *entity.OrderEvent) error {
	// Generate UUID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Convert to model
	eventModel, err := model.FromOrderEventEntity(event)
	if err != nil {
		return err
	}

	// Insert into order_events table
	query := `
		INSERT INTO order_events (
			id, order_id, type, data, version, timestamp, user_id
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	err = r.session.Query(query,
		eventModel.ID,
		eventModel.OrderID,
		eventModel.Type,
		eventModel.Data,
		eventModel.Version,
		eventModel.Timestamp,
		eventModel.UserID,
	).WithContext(ctx).Exec()

	if err != nil {
		return err
	}

	return nil
}

// GetEventsByOrderID retrieves all events for an order
func (r *CassandraOrderEventRepository) GetEventsByOrderID(ctx context.Context, orderID string) ([]*entity.OrderEvent, error) {
	id, err := gocql.ParseUUID(orderID)
	if err != nil {
		return nil, err
	}

	// Query all events for the order, ordered by version
	query := `
		SELECT id, order_id, type, data, version, timestamp, user_id
		FROM order_events
		WHERE order_id = ?
		ORDER BY version ASC
	`

	iter := r.session.Query(query, id).WithContext(ctx).Iter()

	var events []*entity.OrderEvent
	var eventModel model.OrderEventModel

	for iter.Scan(
		&eventModel.ID,
		&eventModel.OrderID,
		&eventModel.Type,
		&eventModel.Data,
		&eventModel.Version,
		&eventModel.Timestamp,
		&eventModel.UserID,
	) {
		event, err := eventModel.ToEntity()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, entity.ErrEventNotFound
	}

	return events, nil
}

// GetEventsByType retrieves events of a specific type
func (r *CassandraOrderEventRepository) GetEventsByType(ctx context.Context, eventType entity.EventType) ([]*entity.OrderEvent, error) {
	// For this implementation, we'll use a secondary index or materialized view
	// In a real-world scenario, you might want to use a dedicated table for this
	query := `
		SELECT id, order_id, type, data, version, timestamp, user_id
		FROM order_events
		WHERE type = ?
		ALLOW FILTERING
	`

	iter := r.session.Query(query, string(eventType)).WithContext(ctx).Iter()

	var events []*entity.OrderEvent
	var eventModel model.OrderEventModel

	for iter.Scan(
		&eventModel.ID,
		&eventModel.OrderID,
		&eventModel.Type,
		&eventModel.Data,
		&eventModel.Version,
		&eventModel.Timestamp,
		&eventModel.UserID,
	) {
		event, err := eventModel.ToEntity()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventsByOrderIDAndType retrieves events for an order of a specific type
func (r *CassandraOrderEventRepository) GetEventsByOrderIDAndType(ctx context.Context, orderID string, eventType entity.EventType) ([]*entity.OrderEvent, error) {
	id, err := gocql.ParseUUID(orderID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, order_id, type, data, version, timestamp, user_id
		FROM order_events
		WHERE order_id = ? AND type = ?
		ORDER BY version ASC
		ALLOW FILTERING
	`

	iter := r.session.Query(query, id, string(eventType)).WithContext(ctx).Iter()

	var events []*entity.OrderEvent
	var eventModel model.OrderEventModel

	for iter.Scan(
		&eventModel.ID,
		&eventModel.OrderID,
		&eventModel.Type,
		&eventModel.Data,
		&eventModel.Version,
		&eventModel.Timestamp,
		&eventModel.UserID,
	) {
		event, err := eventModel.ToEntity()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetLatestEventByOrderID retrieves the most recent event for an order
func (r *CassandraOrderEventRepository) GetLatestEventByOrderID(ctx context.Context, orderID string) (*entity.OrderEvent, error) {
	id, err := gocql.ParseUUID(orderID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, order_id, type, data, version, timestamp, user_id
		FROM order_events
		WHERE order_id = ?
		ORDER BY version DESC
		LIMIT 1
	`

	var eventModel model.OrderEventModel

	err = r.session.Query(query, id).WithContext(ctx).Scan(
		&eventModel.ID,
		&eventModel.OrderID,
		&eventModel.Type,
		&eventModel.Data,
		&eventModel.Version,
		&eventModel.Timestamp,
		&eventModel.UserID,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, entity.ErrEventNotFound
		}
		return nil, err
	}

	return eventModel.ToEntity()
}

// RebuildOrderFromEvents reconstructs an order from its event history
func (r *CassandraOrderEventRepository) RebuildOrderFromEvents(ctx context.Context, orderID string) (*entity.Order, error) {
	events, err := r.GetEventsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Start with empty order
	var order *entity.Order

	// Process events in sequence
	for _, event := range events {
		switch event.Type {
		case entity.EventOrderCreated:
			var data entity.OrderCreatedData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling order created event: %w", err)
			}
			order = &data.Order

		case entity.EventOrderUpdated:
			// Merge the updated fields (this would be more complex in a real implementation)
			if order == nil {
				return nil, fmt.Errorf("received update event but order is nil")
			}
			var updatedOrder entity.Order
			if err := json.Unmarshal(event.Data, &updatedOrder); err != nil {
				return nil, fmt.Errorf("error unmarshaling order updated event: %w", err)
			}
			// Merge fields selectively
			order.UpdatedAt = updatedOrder.UpdatedAt
			order.Version = updatedOrder.Version
			// Add more fields as needed

		case entity.EventStatusChanged:
			if order == nil {
				return nil, fmt.Errorf("received status change event but order is nil")
			}
			var data entity.StatusChangedData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling status changed event: %w", err)
			}
			order.Status = data.NewStatus
			order.UpdatedAt = event.Timestamp
			order.Version = event.Version

		case entity.EventItemAdded:
			if order == nil {
				return nil, fmt.Errorf("received item added event but order is nil")
			}
			var data entity.ItemData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling item added event: %w", err)
			}
			order.AddItem(data.Item)

		case entity.EventItemRemoved:
			if order == nil {
				return nil, fmt.Errorf("received item removed event but order is nil")
			}
			var data entity.ItemData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling item removed event: %w", err)
			}
			order.RemoveItem(data.Item.ProductID)

		case entity.EventItemQuantityUpdated:
			if order == nil {
				return nil, fmt.Errorf("received item quantity updated event but order is nil")
			}
			var data entity.ItemData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling item quantity updated event: %w", err)
			}
			order.UpdateItemQuantity(data.Item.ProductID, data.Quantity)

		case entity.EventDiscountApplied:
			if order == nil {
				return nil, fmt.Errorf("received discount applied event but order is nil")
			}
			var data entity.DiscountData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				return nil, fmt.Errorf("error unmarshaling discount applied event: %w", err)
			}
			order.ApplyDiscount(data.Discount)

		case entity.EventOrderCancelled:
			if order == nil {
				return nil, fmt.Errorf("received order cancelled event but order is nil")
			}
			now := event.Timestamp
			order.Status = valueobject.Cancelled
			order.CancelledAt = &now
			order.UpdatedAt = now
			order.Version = event.Version

		case entity.EventOrderCompleted:
			if order == nil {
				return nil, fmt.Errorf("received order completed event but order is nil")
			}
			now := event.Timestamp
			order.Status = valueobject.Completed
			order.CompletedAt = &now
			order.UpdatedAt = now
			order.Version = event.Version
		}
	}

	if order == nil {
		return nil, entity.ErrOrderNotFound
	}

	return order, nil
}
