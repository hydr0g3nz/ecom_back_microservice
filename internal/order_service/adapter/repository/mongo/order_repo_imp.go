// internal/order_service/adapter/repository/mongo/order_repo_imp.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/mongo/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoOrderRepository implements OrderRepository interface using MongoDB
type MongoOrderRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewMongoOrderRepository creates a new instance of MongoOrderRepository
func NewMongoOrderRepository(db *mongo.Database) repository.OrderRepository {
	return &MongoOrderRepository{
		db:         db,
		collection: db.Collection("orders"),
	}
}

// Create stores a new order
func (r *MongoOrderRepository) Create(ctx context.Context, order entity.Order) (*entity.Order, error) {
	// Convert domain entity to MongoDB model
	orderModel := model.FromEntity(&order)

	// Insert order into MongoDB
	_, err := r.collection.InsertOne(ctx, orderModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, entity.ErrOrderAlreadyExists
		}
		return nil, err
	}

	return &order, nil
}

// GetByID retrieves an order by ID
func (r *MongoOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	// Create filter by ID
	filter := bson.M{"_id": id}

	// Find order in MongoDB
	var orderModel model.OrderModel
	err := r.collection.FindOne(ctx, filter).Decode(&orderModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	// Convert MongoDB model to domain entity
	order := orderModel.ToEntity()
	return order, nil
}

// GetByUserID retrieves orders for a specific user
func (r *MongoOrderRepository) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*entity.Order, int, error) {
	// Create filter by user ID
	filter := bson.M{"user_id": userID}

	// Count total matching documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set up options for pagination and sorting
	findOptions := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at descending

	// Find orders in MongoDB
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var orderModels []model.OrderModel
	if err := cursor.All(ctx, &orderModels); err != nil {
		return nil, 0, err
	}

	// Convert MongoDB models to domain entities
	orders := make([]*entity.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = orderModel.ToEntity()
	}

	return orders, int(total), nil
}

// List retrieves orders with optional filtering
func (r *MongoOrderRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]*entity.Order, int, error) {
	// Build filter based on provided filters
	filter := bson.M{}

	// Apply filters if any
	if filters != nil {
		for key, value := range filters {
			switch key {
			case "status":
				if statusStr, ok := value.(string); ok {
					filter["status"] = statusStr
				}
			case "user_id":
				if userID, ok := value.(string); ok {
					filter["user_id"] = userID
				}
			case "created_after":
				if createdAfter, ok := value.(time.Time); ok {
					filter["created_at"] = bson.M{"$gte": createdAfter}
				}
			case "created_before":
				if createdBefore, ok := value.(time.Time); ok {
					if existingCreatedAt, ok := filter["created_at"].(bson.M); ok {
						existingCreatedAt["$lte"] = createdBefore
					} else {
						filter["created_at"] = bson.M{"$lte": createdBefore}
					}
				}
			case "min_total_amount":
				if minTotal, ok := value.(float64); ok {
					filter["total_amount"] = bson.M{"$gte": minTotal}
				}
			case "max_total_amount":
				if maxTotal, ok := value.(float64); ok {
					if existingTotalAmount, ok := filter["total_amount"].(bson.M); ok {
						existingTotalAmount["$lte"] = maxTotal
					} else {
						filter["total_amount"] = bson.M{"$lte": maxTotal}
					}
				}
			}
		}
	}

	// Count total matching documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Set up options for pagination and sorting
	findOptions := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at descending

	// Find orders in MongoDB
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode results
	var orderModels []model.OrderModel
	if err := cursor.All(ctx, &orderModels); err != nil {
		return nil, 0, err
	}

	// Convert MongoDB models to domain entities
	orders := make([]*entity.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = orderModel.ToEntity()
	}

	return orders, int(total), nil
}

// Update updates an existing order
func (r *MongoOrderRepository) Update(ctx context.Context, order entity.Order) (*entity.Order, error) {
	// Convert domain entity to MongoDB model
	orderModel := model.FromEntity(&order)

	// Create filter by ID
	filter := bson.M{"_id": order.ID}

	// Update order in MongoDB
	result, err := r.collection.ReplaceOne(ctx, filter, orderModel)
	if err != nil {
		return nil, err
	}

	// Check if order was found
	if result.MatchedCount == 0 {
		return nil, entity.ErrOrderNotFound
	}

	return &order, nil
}

// UpdateStatus updates the status of an order
func (r *MongoOrderRepository) UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus, comment string) (*entity.Order, error) {
	// Create filter by ID
	filter := bson.M{"_id": id}

	// Get the existing order
	var orderModel model.OrderModel
	err := r.collection.FindOne(ctx, filter).Decode(&orderModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	// Create status history item
	statusHistoryItem := model.OrderStatusHistoryItem{
		Status:    status.String(),
		Timestamp: time.Now(),
		Comment:   comment,
	}

	// Update status and add to history
	updateQuery := bson.M{
		"$set": bson.M{
			"status":     status.String(),
			"updated_at": time.Now(),
		},
		"$push": bson.M{
			"status_history": statusHistoryItem,
		},
	}

	// Update order in MongoDB
	_, err = r.collection.UpdateOne(ctx, filter, updateQuery)
	if err != nil {
		return nil, err
	}

	// Get the updated order
	var updatedOrderModel model.OrderModel
	err = r.collection.FindOne(ctx, filter).Decode(&updatedOrderModel)
	if err != nil {
		return nil, err
	}

	// Convert MongoDB model to domain entity
	updatedOrder := updatedOrderModel.ToEntity()
	return updatedOrder, nil
}

// Delete removes an order by ID
func (r *MongoOrderRepository) Delete(ctx context.Context, id string) error {
	// Create filter by ID
	filter := bson.M{"_id": id}

	// Get the existing order to check if it exists
	var orderModel model.OrderModel
	err := r.collection.FindOne(ctx, filter).Decode(&orderModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entity.ErrOrderNotFound
		}
		return err
	}

	// Instead of deleting, we'll update the status to cancelled
	statusHistoryItem := model.OrderStatusHistoryItem{
		Status:    valueobject.OrderStatusCancelled.String(),
		Timestamp: time.Now(),
		Comment:   "Order deleted",
	}

	updateQuery := bson.M{
		"$set": bson.M{
			"status":     valueobject.OrderStatusCancelled.String(),
			"updated_at": time.Now(),
		},
		"$push": bson.M{
			"status_history": statusHistoryItem,
		},
	}

	// Update order in MongoDB
	result, err := r.collection.UpdateOne(ctx, filter, updateQuery)
	if err != nil {
		return err
	}

	// Check if order was found
	if result.MatchedCount == 0 {
		return entity.ErrOrderNotFound
	}

	return nil
}
