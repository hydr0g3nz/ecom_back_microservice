package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// OrderModel represents the order model for MongoDB
type OrderModel struct {
	ID              string             `bson:"_id"`
	UserID          string             `bson:"user_id"`
	Items           []OrderItemModel   `bson:"items"`
	TotalAmount     float64            `bson:"total_amount"`
	Status          string             `bson:"status"`
	ShippingAddress string             `bson:"shipping_address"`
	PaymentID       string             `bson:"payment_id"`
	CreatedAt       primitive.DateTime `bson:"created_at"`
	UpdatedAt       primitive.DateTime `bson:"updated_at"`
}

// OrderItemModel represents an item in an order for MongoDB
type OrderItemModel struct {
	ProductID  string  `bson:"product_id"`
	Quantity   int     `bson:"quantity"`
	Price      float64 `bson:"price"`
	TotalPrice float64 `bson:"total_price"`
}

// FromEntity converts a domain entity to a MongoDB model
func (m *OrderModel) FromEntity(order *entity.Order) {
	m.ID = order.ID
	m.UserID = order.UserID
	m.TotalAmount = order.TotalAmount
	m.Status = string(order.Status)
	m.ShippingAddress = order.ShippingAddress
	m.PaymentID = order.PaymentID
	m.CreatedAt = primitive.NewDateTimeFromTime(order.CreatedAt)
	m.UpdatedAt = primitive.NewDateTimeFromTime(order.UpdatedAt)

	m.Items = make([]OrderItemModel, len(order.Items))
	for i, item := range order.Items {
		m.Items[i] = OrderItemModel{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			TotalPrice: item.TotalPrice,
		}
	}
}

// ToEntity converts a MongoDB model to a domain entity
func (m *OrderModel) ToEntity() *entity.Order {
	items := make([]entity.OrderItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = entity.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.Price,
			TotalPrice: item.TotalPrice,
		}
	}

	return &entity.Order{
		ID:              m.ID,
		UserID:          m.UserID,
		Items:           items,
		TotalAmount:     m.TotalAmount,
		Status:          valueobject.OrderStatus(m.Status),
		ShippingAddress: m.ShippingAddress,
		PaymentID:       m.PaymentID,
		CreatedAt:       m.CreatedAt.Time(),
		UpdatedAt:       m.UpdatedAt.Time(),
	}
}

// NewOrderModel creates a new OrderModel from an entity
func NewOrderModel(order *entity.Order) *OrderModel {
	model := &OrderModel{}
	model.FromEntity(order)
	return model
}
