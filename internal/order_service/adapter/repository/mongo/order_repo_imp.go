package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/repository/mongo/model"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

const (
	// CollectionName is the name of the orders collection
	CollectionName = "orders"
)

// OrderRepository is the MongoDB implementation of the OrderRepository interface
type OrderRepository struct {
	collection *mongo.Collection
}

// NewOrderRepository creates a new MongoDB order repository
func NewOrderRepository(db *mongo.Database) *OrderRepository {
	collection := db.Collection(CollectionName)
	
	// Ensure indexes for better query performance
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
	}
	
	_, err := collection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		// Log error but continue
		// TODO: implement proper logging
	}
	
	return &OrderRepository{
		collection: collection,
	}
}

// Create saves a new order in the repository
func (r *OrderRepository) Create(ctx context.Context, order *entity.Order) error {
	orderModel := model.NewOrderModel(order)
	
	_, err := r.collection.InsertOne(ctx, orderModel)
	if mongo.IsDuplicateKeyError(err) {
		return entity.ErrOrderAlreadyExists
	}
	
	return err
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	var orderModel model.OrderModel
	
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&orderModel)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}
	
	return orderModel.ToEntity(), nil
}

// GetByUserID retrieves all orders for a specific user
func (r *OrderRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var orderModels []model.OrderModel
	if err := cursor.All(ctx, &orderModels); err != nil {
		return nil, err
	}
	
	orders := make([]*entity.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = orderModel.ToEntity()
	}
	
	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, order *entity.Order) error {
	orderModel := model.NewOrderModel(order)
	
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": order.ID}, orderModel)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return entity.ErrOrderNotFound
	}
	
	return nil
}

// Delete removes an order by its ID
func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return entity.ErrOrderNotFound
	}
	
	return nil
}

// UpdateStatus updates the status of an order
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status valueobject.OrderStatus) error {
	if !status.IsValid() {
		return entity.ErrInvalidOrderStatus
	}
	
	now := time.Now()
	if t, ok := ctx.Value("now_time").(time.Time); ok {
		now = t
	}
	
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":     string(status),
				"updated_at": now,
			},
		},
	)
	
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return entity.ErrOrderNotFound
	}
	
	return nil
}

// ListByStatus retrieves orders by their status
func (r *OrderRepository) ListByStatus(ctx context.Context, status valueobject.OrderStatus) ([]*entity.Order, error) {
	if !status.IsValid() {
		return nil, entity.ErrInvalidOrderStatus
	}
	
	cursor, err := r.collection.Find(ctx, bson.M{"status": string(status)})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var orderModels []model.OrderModel
	if err := cursor.All(ctx, &orderModels); err != nil {
		return nil, err
	}
	
	orders := make([]*entity.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = orderModel.ToEntity()
	}
	
	return orders, nil
}

// GetOrdersPaginated retrieves orders with pagination
func (r *OrderRepository) GetOrdersPaginated(ctx context.Context, page, pageSize int) ([]*entity.Order, int, error) {
	skip := (page - 1) * pageSize
	
	// Get total count
	totalCount, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	
	// Set options for pagination and sorting
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	// Execute query
	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var orderModels []model.OrderModel
	if err := cursor.All(ctx, &orderModels); err != nil {
		return nil, 0, err
	}
	
	orders := make([]*entity.Order, len(orderModels))
	for i, orderModel := range orderModels {
		orders[i] = orderModel.ToEntity()
	}
	
	return orders, int(totalCount), nil
}
