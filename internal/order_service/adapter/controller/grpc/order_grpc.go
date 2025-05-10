// internal/order_service/adapter/controller/grpc/order_grpc.go
package grpcctl

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderServer implements the gRPC OrderService interface
type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	orderUsecase usecase.OrderUsecase
	logger       logger.Logger
}

// NewOrderServer creates a new OrderServer instance
func NewOrderServer(orderUsecase usecase.OrderUsecase, logger logger.Logger) *OrderServer {
	return &OrderServer{
		orderUsecase: orderUsecase,
		logger:       logger,
	}
}

// CreateOrder creates a new order
func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC CreateOrder request received", "user_id", req.UserId)

	// Convert items from proto to entity
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Quantity:    int(item.Quantity),
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Create order entity
	order := entity.Order{
		UserID: req.UserId,
		Items:  items,
		ShippingInfo: entity.Address{
			Street:     req.ShippingInfo.Street,
			City:       req.ShippingInfo.City,
			State:      req.ShippingInfo.State,
			Country:    req.ShippingInfo.Country,
			PostalCode: req.ShippingInfo.PostalCode,
		},
		BillingInfo: entity.Address{
			Street:     req.BillingInfo.Street,
			City:       req.BillingInfo.City,
			State:      req.BillingInfo.State,
			Country:    req.BillingInfo.Country,
			PostalCode: req.BillingInfo.PostalCode,
		},
		Payment: entity.Payment{
			Method: req.Payment.Method,
			Amount: req.Payment.Amount,
		},
	}

	if req.Notes != nil {
		order.Notes = *req.Notes
	}

	// Call usecase to create order
	createdOrder, err := s.orderUsecase.CreateOrder(ctx, &order)
	if err != nil {
		s.logger.Error("Failed to create order", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	return convertOrderToProto(createdOrder), nil
}

// GetOrder gets an order by ID
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC GetOrder request received", "id", req.Id)

	// Validate required fields
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	// Get order by ID
	order, err := s.orderUsecase.GetOrderByID(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get order", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	return convertOrderToProto(order), nil
}

// ListOrders lists orders with optional filtering
func (s *OrderServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("gRPC ListOrders request received", "page", req.Page, "pageSize", req.PageSize)

	// Set default pagination values
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Convert filters
	filters := make(map[string]interface{})
	for k, v := range req.Filters {
		filters[k] = v
	}

	// Get orders
	orders, total, err := s.orderUsecase.ListOrders(ctx, page, pageSize, filters)
	if err != nil {
		s.logger.Error("Failed to list orders", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	orderResponses := make([]*pb.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = convertOrderToProto(order)
	}

	// Calculate total pages
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListOrdersResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Orders:     orderResponses,
	}, nil
}

// UpdateOrder updates an existing order
func (s *OrderServer) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC UpdateOrder request received", "id", req.Id)

	// Validate required fields
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	// Convert items from proto to entity
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID:   item.ProductId,
			ProductName: item.ProductName,
			Quantity:    int(item.Quantity),
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Create order entity
	order := entity.Order{
		UserID: req.UserId,
		Items:  items,
		ShippingInfo: entity.Address{
			Street:     req.ShippingInfo.Street,
			City:       req.ShippingInfo.City,
			State:      req.ShippingInfo.State,
			Country:    req.ShippingInfo.Country,
			PostalCode: req.ShippingInfo.PostalCode,
		},
		BillingInfo: entity.Address{
			Street:     req.BillingInfo.Street,
			City:       req.BillingInfo.City,
			State:      req.BillingInfo.State,
			Country:    req.BillingInfo.Country,
			PostalCode: req.BillingInfo.PostalCode,
		},
		Payment: entity.Payment{
			Method: req.Payment.Method,
			Amount: req.Payment.Amount,
		},
	}

	if req.Notes != nil {
		order.Notes = *req.Notes
	}

	// Call usecase to update order
	updatedOrder, err := s.orderUsecase.UpdateOrder(ctx, req.Id, order)
	if err != nil {
		s.logger.Error("Failed to update order", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	return convertOrderToProto(updatedOrder), nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC UpdateOrderStatus request received", "id", req.Id, "status", req.Status)

	// Validate required fields
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}
	if req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	// Parse status
	orderStatus, err := valueobject.ParseOrderStatus(req.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order status")
	}

	// Call usecase to update order status
	updatedOrder, err := s.orderUsecase.UpdateOrderStatus(ctx, req.Id, orderStatus, req.Comment)
	if err != nil {
		s.logger.Error("Failed to update order status", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	return convertOrderToProto(updatedOrder), nil
}

// CancelOrder cancels an order
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC CancelOrder request received", "id", req.Id)

	// Validate required fields
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID is required")
	}

	// Call usecase to cancel order
	cancelledOrder, err := s.orderUsecase.CancelOrder(ctx, req.Id, req.Reason)
	if err != nil {
		s.logger.Error("Failed to cancel order", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	return convertOrderToProto(cancelledOrder), nil
}

// GetOrdersByUser gets orders for a specific user
func (s *OrderServer) GetOrdersByUser(ctx context.Context, req *pb.GetOrdersByUserRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("gRPC GetOrdersByUser request received", "user_id", req.UserId)

	// Validate required fields
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Set default pagination values
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get orders by user ID
	orders, total, err := s.orderUsecase.GetOrdersByUserID(ctx, req.UserId, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to get orders by user", "error", err)
		return nil, handleError(err)
	}

	// Convert to proto response
	orderResponses := make([]*pb.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = convertOrderToProto(order)
	}

	// Calculate total pages
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListOrdersResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Orders:     orderResponses,
	}, nil
}

// Helper function to convert domain entity to protobuf
func convertOrderToProto(order *entity.Order) *pb.OrderResponse {
	// Convert items
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    int32(item.Quantity),
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	// Convert status history
	statusHistory := make([]*pb.StatusHistoryItem, len(order.StatusHistory))
	for i, history := range order.StatusHistory {
		statusHistory[i] = &pb.StatusHistoryItem{
			Status:    history.Status.String(),
			Timestamp: timestamppb.New(history.Timestamp),
			Comment:   history.Comment,
		}
	}

	// Create payment with optional fields
	payment := &pb.Payment{
		Method: order.Payment.Method,
		Amount: order.Payment.Amount,
	}
	if order.Payment.TransactionID != "" {
		payment.TransactionId = &order.Payment.TransactionID
	}
	if order.Payment.Status != "" {
		payment.Status = &order.Payment.Status
	}
	if order.Payment.PaidAt != nil {
		payment.PaidAt = timestamppb.New(*order.Payment.PaidAt)
	}

	// Return complete order response
	return &pb.OrderResponse{
		Id:          order.ID,
		UserId:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      order.Status.String(),
		ShippingInfo: &pb.Address{
			Street:     order.ShippingInfo.Street,
			City:       order.ShippingInfo.City,
			State:      order.ShippingInfo.State,
			Country:    order.ShippingInfo.Country,
			PostalCode: order.ShippingInfo.PostalCode,
		},
		BillingInfo: &pb.Address{
			Street:     order.BillingInfo.Street,
			City:       order.BillingInfo.City,
			State:      order.BillingInfo.State,
			Country:    order.BillingInfo.Country,
			PostalCode: order.BillingInfo.PostalCode,
		},
		Payment:       payment,
		Notes:         order.Notes,
		CreatedAt:     timestamppb.New(order.CreatedAt),
		UpdatedAt:     timestamppb.New(order.UpdatedAt),
		StatusHistory: statusHistory,
	}
}

// handleError maps domain errors to appropriate gRPC status errors
func handleError(err error) error {
	var statusCode codes.Code
	var message string

	switch {
	case errors.Is(err, entity.ErrOrderNotFound):
		statusCode = codes.NotFound
		message = "Order not found"
	case errors.Is(err, entity.ErrInvalidOrderData):
		statusCode = codes.InvalidArgument
		message = "Invalid order data"
	case errors.Is(err, entity.ErrInvalidOrderStatus):
		statusCode = codes.InvalidArgument
		message = "Invalid order status"
	case errors.Is(err, entity.ErrInvalidStatusTransition):
		statusCode = codes.FailedPrecondition
		message = "Invalid status transition"
	case errors.Is(err, entity.ErrOrderAlreadyExists):
		statusCode = codes.AlreadyExists
		message = "Order already exists"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = codes.ResourceExhausted
		message = "Insufficient stock"
	case errors.Is(err, entity.ErrPaymentFailed):
		statusCode = codes.Aborted
		message = "Payment failed"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = codes.Internal
		message = "Internal server error"
	default:
		statusCode = codes.Internal
		message = "Something went wrong"
	}

	return status.Error(statusCode, message)
}
