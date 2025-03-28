package grpcctl

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto_mapper"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/query"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderController implements the gRPC OrderService interface
type OrderController struct {
	pb.UnimplementedOrderServiceServer
	createOrderUsecase    command.CreateOrderUsecase
	updateOrderUsecase    command.UpdateOrderUsecase
	cancelOrderUsecase    command.CancelOrderUsecase
	processPaymentUsecase command.ProcessPaymentUsecase
	updateShippingUsecase command.UpdateShippingUsecase
	getOrderUsecase       query.GetOrderUsecase
	listOrdersUsecase     query.ListOrdersUsecase
	orderHistoryUsecase   query.OrderHistoryUsecase
	logger                logger.Logger
}

// NewOrderController creates a new OrderController instance
func NewOrderController(
	createOrderUsecase command.CreateOrderUsecase,
	updateOrderUsecase command.UpdateOrderUsecase,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
	getOrderUsecase query.GetOrderUsecase,
	listOrdersUsecase query.ListOrdersUsecase,
	orderHistoryUsecase query.OrderHistoryUsecase,
	logger logger.Logger,
) *OrderController {
	return &OrderController{
		createOrderUsecase:    createOrderUsecase,
		updateOrderUsecase:    updateOrderUsecase,
		cancelOrderUsecase:    cancelOrderUsecase,
		processPaymentUsecase: processPaymentUsecase,
		updateShippingUsecase: updateShippingUsecase,
		getOrderUsecase:       getOrderUsecase,
		listOrdersUsecase:     listOrdersUsecase,
		orderHistoryUsecase:   orderHistoryUsecase,
		logger:                logger,
	}
}

// CreateOrder handles creating a new order
func (c *OrderController) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	c.logger.Info("gRPC CreateOrder request received", "user_id", req.UserId)

	// Convert request to DTO
	createOrderInput := proto_mapper.CreateOrderInputFromProto(req)

	// Execute the use case
	orderDTO, err := c.createOrderUsecase.Execute(ctx, createOrderInput)
	if err != nil {
		c.logger.Error("Failed to create order", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.OrderResponseFromDTO(orderDTO)
	return response, nil
}

// GetOrder handles retrieving an order by ID
func (c *OrderController) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	c.logger.Info("gRPC GetOrder request received", "id", req.Id)

	// Execute the use case
	orderDTO, err := c.getOrderUsecase.Execute(ctx, req.Id)
	if err != nil {
		c.logger.Error("Failed to get order", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.OrderResponseFromDTO(orderDTO)
	return response, nil
}

// UpdateOrder handles updating an existing order
func (c *OrderController) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.OrderResponse, error) {
	c.logger.Info("gRPC UpdateOrder request received", "id", req.Id)

	// Convert request to DTO
	updateOrderInput := proto_mapper.UpdateOrderInputFromProto(req)

	// Execute the use case
	orderDTO, err := c.updateOrderUsecase.Execute(ctx, req.Id, updateOrderInput)
	if err != nil {
		c.logger.Error("Failed to update order", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.OrderResponseFromDTO(orderDTO)
	return response, nil
}

// CancelOrder handles cancelling an order
func (c *OrderController) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*emptypb.Empty, error) {
	c.logger.Info("gRPC CancelOrder request received", "id", req.Id)

	// Execute the use case
	err := c.cancelOrderUsecase.Execute(ctx, req.Id, req.Reason)
	if err != nil {
		c.logger.Error("Failed to cancel order", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// ListOrdersByUser handles retrieving orders for a user
func (c *OrderController) ListOrdersByUser(ctx context.Context, req *pb.ListOrdersByUserRequest) (*pb.ListOrdersResponse, error) {
	c.logger.Info("gRPC ListOrdersByUser request received", "user_id", req.UserId)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Execute the use case
	ordersOutput, err := c.listOrdersUsecase.ListByUser(ctx, req.UserId, page, pageSize)
	if err != nil {
		c.logger.Error("Failed to list orders by user", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.ListOrdersResponseFromDTO(ordersOutput)
	return response, nil
}

// ListOrdersByStatus handles retrieving orders with a specific status
func (c *OrderController) ListOrdersByStatus(ctx context.Context, req *pb.ListOrdersByStatusRequest) (*pb.ListOrdersResponse, error) {
	c.logger.Info("gRPC ListOrdersByStatus request received", "status", req.Status)

	status, err := valueobject.ParseOrderStatus(req.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid order status")
	}

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Execute the use case
	ordersOutput, err := c.listOrdersUsecase.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		c.logger.Error("Failed to list orders by status", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.ListOrdersResponseFromDTO(ordersOutput)
	return response, nil
}

// SearchOrders handles searching orders based on criteria
func (c *OrderController) SearchOrders(ctx context.Context, req *pb.SearchOrdersRequest) (*pb.ListOrdersResponse, error) {
	c.logger.Info("gRPC SearchOrders request received")

	// Convert search criteria
	criteria := make(map[string]interface{})

	if req.UserId != "" {
		criteria["user_id"] = req.UserId
	}

	if req.Status != "" {
		criteria["status"] = req.Status
	}

	if req.ProductId != "" {
		criteria["product_id"] = req.ProductId
	}

	if req.StartDate != nil {
		criteria["start_date"] = req.StartDate.AsTime()
	}

	if req.EndDate != nil {
		criteria["end_date"] = req.EndDate.AsTime()
	}

	if req.MinAmount > 0 {
		criteria["min_amount"] = req.MinAmount
	}

	if req.MaxAmount > 0 {
		criteria["max_amount"] = req.MaxAmount
	}

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Execute the use case
	ordersOutput, err := c.listOrdersUsecase.Search(ctx, criteria, page, pageSize)
	if err != nil {
		c.logger.Error("Failed to search orders", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.ListOrdersResponseFromDTO(ordersOutput)
	return response, nil
}

// GetOrderHistory handles retrieving the event history for an order
func (c *OrderController) GetOrderHistory(ctx context.Context, req *pb.GetOrderHistoryRequest) (*pb.OrderHistoryResponse, error) {
	c.logger.Info("gRPC GetOrderHistory request received", "order_id", req.OrderId)

	// Execute the use case
	events, err := c.orderHistoryUsecase.GetEvents(ctx, req.OrderId)
	if err != nil {
		c.logger.Error("Failed to get order history", "error", err)
		return nil, handleError(err)
	}

	// Convert events to proto response
	protoEvents := make([]*pb.OrderEvent, len(events))
	for i, event := range events {
		protoEvents[i] = &pb.OrderEvent{
			Id:        event.ID.String(),
			OrderId:   event.OrderID.String(),
			Type:      string(event.Type),
			Data:      event.Data,
			Version:   int32(event.Version),
			Timestamp: timestamppb.New(event.Timestamp.Time()),
			UserId:    event.UserID.String(),
		}
	}

	return &pb.OrderHistoryResponse{
		OrderId: req.OrderId,
		Events:  protoEvents,
	}, nil
}

// ProcessPayment handles processing a payment for an order
func (c *OrderController) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
	c.logger.Info("gRPC ProcessPayment request received", "order_id", req.OrderId)

	// Convert request to DTO
	processPaymentInput := proto_mapper.ProcessPaymentInputFromProto(req)

	// Execute the use case
	paymentDTO, err := c.processPaymentUsecase.Execute(ctx, processPaymentInput)
	if err != nil {
		c.logger.Error("Failed to process payment", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.PaymentResponseFromDTO(paymentDTO)
	return response, nil
}

// UpdateShipping handles updating shipping information for an order
func (c *OrderController) UpdateShipping(ctx context.Context, req *pb.UpdateShippingRequest) (*pb.ShippingResponse, error) {
	c.logger.Info("gRPC UpdateShipping request received", "order_id", req.OrderId)

	// Convert request to DTO
	var estimatedDelivery *string
	if req.EstimatedDelivery != nil {
		timeStr := req.EstimatedDelivery.AsTime().Format(time.RFC3339)
		estimatedDelivery = &timeStr
	}

	updateShippingInput := dto.UpdateShippingInput{
		OrderID:           req.OrderId,
		Carrier:           req.Carrier,
		TrackingNumber:    req.TrackingNumber,
		Status:            req.Status,
		EstimatedDelivery: estimatedDelivery,
		ShippingMethod:    req.ShippingMethod,
		ShippingCost:      req.ShippingCost,
		Notes:             req.Notes,
	}

	// Execute the use case
	shippingDTO, err := c.updateShippingUsecase.Execute(ctx, updateShippingInput)
	if err != nil {
		c.logger.Error("Failed to update shipping", "error", err)
		return nil, handleError(err)
	}

	// Convert DTO to response
	response := proto_mapper.ShippingResponseFromDTO(shippingDTO)
	return response, nil
}

// GetShipping handles retrieving shipping information for an order
func (c *OrderController) GetShipping(ctx context.Context, req *pb.GetShippingRequest) (*pb.ShippingResponse, error) {
	c.logger.Info("gRPC GetShipping request received", "order_id", req.OrderId)

	// Get order first to validate
	orderDTO, err := c.getOrderUsecase.Execute(ctx, req.OrderId)
	if err != nil {
		c.logger.Error("Failed to get order for shipping information", "error", err)
		return nil, handleError(err)
	}

	if orderDTO.ShippingID == "" {
		return nil, status.Error(codes.NotFound, "Shipping information not found for this order")
	}

	// In a real implementation, we would call a GetShippingUsecase
	// For now, return a placeholder shipping response
	return &pb.ShippingResponse{
		Id:        "shipping-not-implemented",
		OrderId:   req.OrderId,
		Status:    "pending",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

// handleError maps domain errors to appropriate gRPC status errors
func handleError(err error) error {
	var statusCode codes.Code
	var message string

	switch {
	case err == entity.ErrOrderNotFound:
		statusCode = codes.NotFound
		message = "Order not found"
	case err == entity.ErrInvalidOrderData:
		statusCode = codes.InvalidArgument
		message = "Invalid order data"
	case err == entity.ErrInvalidOrderStatus:
		statusCode = codes.FailedPrecondition
		message = "Invalid order status transition"
	case err == entity.ErrOrderCancelled:
		statusCode = codes.FailedPrecondition
		message = "Order is already cancelled"
	case err == entity.ErrOrderCompleted:
		statusCode = codes.FailedPrecondition
		message = "Order is already completed"
	case err == entity.ErrPaymentNotFound:
		statusCode = codes.NotFound
		message = "Payment not found"
	case err == entity.ErrPaymentFailed:
		statusCode = codes.FailedPrecondition
		message = "Payment failed"
	case err == entity.ErrShippingNotFound:
		statusCode = codes.NotFound
		message = "Shipping not found"
	case err == entity.ErrItemNotFound:
		statusCode = codes.NotFound
		message = "Item not found in order"
	case err == entity.ErrInsufficientStock:
		statusCode = codes.ResourceExhausted
		message = "Insufficient stock for item"
	case err == entity.ErrInternalServerError:
		statusCode = codes.Internal
		message = "Internal server error"
	default:
		statusCode = codes.Unknown
		message = err.Error()
	}

	return status.Error(statusCode, message)
}
