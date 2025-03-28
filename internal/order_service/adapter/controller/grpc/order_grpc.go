package grpc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase"
)

// OrderServer implements the OrderService gRPC server interface
type OrderServer struct {
	proto.UnimplementedOrderServiceServer
	orderUseCase *usecase.OrderUseCase
}

// NewOrderServer creates a new order gRPC server
func NewOrderServer(orderUseCase *usecase.OrderUseCase) *OrderServer {
	return &OrderServer{
		orderUseCase: orderUseCase,
	}
}

// CreateOrder creates a new order
func (s *OrderServer) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.OrderResponse, error) {
	// Convert request to domain model
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID:  item.ProductId,
			Quantity:   int(item.Quantity),
			Price:      item.Price,
			TotalPrice: item.Price * float64(item.Quantity),
		}
	}
	
	// Create order
	order, err := s.orderUseCase.CreateOrder(ctx, req.UserId, items, req.ShippingAddress)
	if err != nil {
		switch err {
		case entity.ErrInvalidUserID, entity.ErrEmptyOrderItems:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case entity.ErrOrderAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to create order")
		}
	}
	
	// Convert to response
	return s.orderToProto(order), nil
}

// GetOrder gets an order by ID
func (s *OrderServer) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.OrderResponse, error) {
	order, err := s.orderUseCase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		if err == entity.ErrOrderNotFound {
			return nil, status.Error(codes.NotFound, "Order not found")
		}
		return nil, status.Error(codes.Internal, "Failed to get order")
	}
	
	return s.orderToProto(order), nil
}

// GetUserOrders gets orders for a specific user
func (s *OrderServer) GetUserOrders(ctx context.Context, req *proto.GetUserOrdersRequest) (*proto.OrderListResponse, error) {
	orders, err := s.orderUseCase.GetOrdersByUserID(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get user orders")
	}
	
	// Convert to response
	orderResponses := make([]*proto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = s.orderToProto(order)
	}
	
	return &proto.OrderListResponse{
		Orders:     orderResponses,
		TotalCount: int32(len(orderResponses)),
		Page:       1,
		PageSize:   int32(len(orderResponses)),
		TotalPages: 1,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.OrderResponse, error) {
	err := s.orderUseCase.UpdateOrderStatus(ctx, req.OrderId, valueobject.OrderStatus(req.Status))
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			return nil, status.Error(codes.NotFound, "Order not found")
		case entity.ErrInvalidOrderStatus:
			return nil, status.Error(codes.InvalidArgument, "Invalid order status")
		default:
			return nil, status.Error(codes.Internal, "Failed to update order status")
		}
	}
	
	// Get updated order
	order, err := s.orderUseCase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get updated order")
	}
	
	return s.orderToProto(order), nil
}

// ProcessPayment processes payment for an order
func (s *OrderServer) ProcessPayment(ctx context.Context, req *proto.ProcessPaymentRequest) (*proto.OrderResponse, error) {
	err := s.orderUseCase.AddPaymentToOrder(ctx, req.OrderId, req.PaymentId)
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			return nil, status.Error(codes.NotFound, "Order not found")
		default:
			return nil, status.Error(codes.Internal, "Failed to process payment")
		}
	}
	
	// Get updated order
	order, err := s.orderUseCase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get updated order")
	}
	
	return s.orderToProto(order), nil
}

// CancelOrder cancels an order
func (s *OrderServer) CancelOrder(ctx context.Context, req *proto.CancelOrderRequest) (*proto.OrderResponse, error) {
	err := s.orderUseCase.CancelOrder(ctx, req.OrderId)
	if err != nil {
		switch err {
		case entity.ErrOrderNotFound:
			return nil, status.Error(codes.NotFound, "Order not found")
		default:
			return nil, status.Error(codes.Internal, "Failed to cancel order")
		}
	}
	
	// Get updated order
	order, err := s.orderUseCase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get updated order")
	}
	
	return s.orderToProto(order), nil
}

// ListOrders lists orders with pagination and filtering
func (s *OrderServer) ListOrders(ctx context.Context, req *proto.ListOrdersRequest) (*proto.OrderListResponse, error) {
	var orders []*entity.Order
	var totalCount int
	var err error
	
	if req.Status != "" {
		// Filter by status if provided
		orderStatus := valueobject.OrderStatus(req.Status)
		if !orderStatus.IsValid() {
			return nil, status.Error(codes.InvalidArgument, "Invalid order status")
		}
		
		orders, err = s.orderUseCase.ListOrdersByStatus(ctx, orderStatus)
		totalCount = len(orders)
	} else {
		// Get paginated orders
		orders, totalCount, err = s.orderUseCase.GetOrdersPaginated(ctx, int(req.Page), int(req.Limit))
	}
	
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to list orders")
	}
	
	// Convert to response
	orderResponses := make([]*proto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = s.orderToProto(order)
	}
	
	return &proto.OrderListResponse{
		Orders:     orderResponses,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		PageSize:   req.Limit,
		TotalPages: int32((totalCount + int(req.Limit) - 1) / int(req.Limit)),
	}, nil
}

// Helper function to convert domain order to proto order
func (s *OrderServer) orderToProto(order *entity.Order) *proto.OrderResponse {
	// Convert items
	items := make([]*proto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = &proto.OrderItemResponse{
			ProductId:  item.ProductID,
			Quantity:   int32(item.Quantity),
			Price:      item.Price,
			TotalPrice: item.TotalPrice,
		}
	}
	
	// Convert timestamps
	createdAt := timestamppb.New(order.CreatedAt)
	updatedAt := timestamppb.New(order.UpdatedAt)
	
	return &proto.OrderResponse{
		Id:              order.ID,
		UserId:          order.UserID,
		Items:           items,
		TotalAmount:     order.TotalAmount,
		Status:          string(order.Status),
		ShippingAddress: order.ShippingAddress,
		PaymentId:       order.PaymentID,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}
