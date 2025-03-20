package grpcctl

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/command"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/query"
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
)

// OrderServer implements the gRPC OrderService interface
type OrderServer struct {
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

// NewOrderServer creates a new OrderServer instance
func NewOrderServer(
	createOrderUsecase command.CreateOrderUsecase,
	updateOrderUsecase command.UpdateOrderUsecase,
	cancelOrderUsecase command.CancelOrderUsecase,
	processPaymentUsecase command.ProcessPaymentUsecase,
	updateShippingUsecase command.UpdateShippingUsecase,
	getOrderUsecase query.GetOrderUsecase,
	listOrdersUsecase query.ListOrdersUsecase,
	orderHistoryUsecase query.OrderHistoryUsecase,
	logger logger.Logger,
) *OrderServer {
	return &OrderServer{
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
func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC CreateOrder request received", "user_id", req.UserId)

	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductID:    item.ProductId,
			Name:         item.Name,
			SKU:          item.Sku,
			Quantity:     int(item.Quantity),
			Price:        item.Price,
			TotalPrice:   item.Price * float64(item.Quantity),
			CurrencyCode: item.CurrencyCode,
		}
	}

	input := command.CreateOrderInput{
		UserID: req.UserId,
		Items:  items,
		ShippingAddress: entity.Address{
			FirstName:    req.ShippingAddress.FirstName,
			LastName:     req.ShippingAddress.LastName,
			AddressLine1: req.ShippingAddress.AddressLine1,
			AddressLine2: req.ShippingAddress.AddressLine2,
			City:         req.ShippingAddress.City,
			State:        req.ShippingAddress.State,
			PostalCode:   req.ShippingAddress.PostalCode,
			Country:      req.ShippingAddress.Country,
			Phone:        req.ShippingAddress.Phone,
			Email:        req.ShippingAddress.Email,
		},
		BillingAddress: entity.Address{
			FirstName:    req.BillingAddress.FirstName,
			LastName:     req.BillingAddress.LastName,
			AddressLine1: req.BillingAddress.AddressLine1,
			AddressLine2: req.BillingAddress.AddressLine2,
			City:         req.BillingAddress.City,
			State:        req.BillingAddress.State,
			PostalCode:   req.BillingAddress.PostalCode,
			Country:      req.BillingAddress.Country,
			Phone:        req.BillingAddress.Phone,
			Email:        req.BillingAddress.Email,
		},
		Notes:          req.Notes,
		PromotionCodes: req.PromotionCodes,
	}

	order, err := s.createOrderUsecase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to create order", "error", err)
		return nil, handleError(err)
	}

	return convertOrderToProto(order), nil
}

// GetOrder handles retrieving an order by ID
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC GetOrder request received", "id", req.Id)

	order, err := s.getOrderUsecase.Execute(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get order", "error", err)
		return nil, handleError(err)
	}

	return convertOrderToProto(order), nil
}

// UpdateOrder handles updating an existing order
func (s *OrderServer) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.OrderResponse, error) {
	s.logger.Info("gRPC UpdateOrder request received", "id", req.Id)

	var shippingAddress *entity.Address
	var billingAddress *entity.Address
	var notes *string

	if req.ShippingAddress != nil {
		address := entity.Address{
			FirstName:    req.ShippingAddress.FirstName,
			LastName:     req.ShippingAddress.LastName,
			AddressLine1: req.ShippingAddress.AddressLine1,
			AddressLine2: req.ShippingAddress.AddressLine2,
			City:         req.ShippingAddress.City,
			State:        req.ShippingAddress.State,
			PostalCode:   req.ShippingAddress.PostalCode,
			Country:      req.ShippingAddress.Country,
			Phone:        req.ShippingAddress.Phone,
			Email:        req.ShippingAddress.Email,
		}
		shippingAddress = &address
	}

	if req.BillingAddress != nil {
		address := entity.Address{
			FirstName:    req.BillingAddress.FirstName,
			LastName:     req.BillingAddress.LastName,
			AddressLine1: req.BillingAddress.AddressLine1,
			AddressLine2: req.BillingAddress.AddressLine2,
			City:         req.BillingAddress.City,
			State:        req.BillingAddress.State,
			PostalCode:   req.BillingAddress.PostalCode,
			Country:      req.BillingAddress.Country,
			Phone:        req.BillingAddress.Phone,
			Email:        req.BillingAddress.Email,
		}
		billingAddress = &address
	}

	if req.Notes != "" {
		notes = &req.Notes
	}

	input := command.UpdateOrderInput{
		Notes:           notes,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
	}

	order, err := s.updateOrderUsecase.Execute(ctx, req.Id, input)
	if err != nil {
		s.logger.Error("Failed to update order", "error", err)
		return nil, handleError(err)
	}

	return convertOrderToProto(order), nil
}

// CancelOrder handles cancelling an order
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*emptypb.Empty, error) {
	s.logger.Info("gRPC CancelOrder request received", "id", req.Id)

	err := s.cancelOrderUsecase.Execute(ctx, req.Id, req.Reason)
	if err != nil {
		s.logger.Error("Failed to cancel order", "error", err)
		return nil, handleError(err)
	}

	return &emptypb.Empty{}, nil
}

// ListOrdersByUser handles retrieving orders for a user
func (s *OrderServer) ListOrdersByUser(ctx context.Context, req *pb.ListOrdersByUserRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("gRPC ListOrdersByUser request received", "user_id", req.UserId)

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orders, total, err := s.listOrdersUsecase.ListByUser(ctx, req.UserId, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list orders by user", "error", err)
		return nil, handleError(err)
	}

	return createListOrdersResponse(orders, total, page, pageSize), nil
}

// ListOrdersByStatus handles retrieving orders with a specific status
func (s *OrderServer) ListOrdersByStatus(ctx context.Context, req *pb.ListOrdersByStatusRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("gRPC ListOrdersByStatus request received", "status", req.Status)

	status, err := valueobject.ParseOrderStatus(req.Status)
	if err != nil {
		return nil, handleError(err)
	}

	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orders, total, err := s.listOrdersUsecase.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to list orders by status", "error", err)
		return nil, handleError(err)
	}

	return createListOrdersResponse(orders, total, page, pageSize), nil
}

// SearchOrders handles searching orders based on criteria
func (s *OrderServer) SearchOrders(ctx context.Context, req *pb.SearchOrdersRequest) (*pb.ListOrdersResponse, error) {
	s.logger.Info("gRPC SearchOrders request received")

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

	orders, total, err := s.listOrdersUsecase.Search(ctx, criteria, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to search orders", "error", err)
		return nil, handleError(err)
	}

	return createListOrdersResponse(orders, total, page, pageSize), nil
}

// GetOrderHistory handles retrieving the event history for an order
func (s *OrderServer) GetOrderHistory(ctx context.Context, req *pb.GetOrderHistoryRequest) (*pb.OrderHistoryResponse, error) {
	s.logger.Info("gRPC GetOrderHistory request received", "order_id", req.OrderId)

	events, err := s.orderHistoryUsecase.GetEvents(ctx, req.OrderId)
	if err != nil {
		s.logger.Error("Failed to get order history", "error", err)
		return nil, handleError(err)
	}

	protoEvents := make([]*pb.OrderEvent, len(events))
	for i, event := range events {
		protoEvents[i] = &pb.OrderEvent{
			Id:        event.ID,
			OrderId:   event.OrderID,
			Type:      string(event.Type),
			Data:      event.Data,
			Version:   int32(event.Version),
			Timestamp: timestamppb.New(event.Timestamp),
			UserId:    event.UserID,
		}
	}

	return &pb.OrderHistoryResponse{
		OrderId: req.OrderId,
		Events:  protoEvents,
	}, nil
}

// ProcessPayment handles processing a payment for an order
func (s *OrderServer) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
	s.logger.Info("gRPC ProcessPayment request received", "order_id", req.OrderId)

	input := command.ProcessPaymentInput{
		OrderID:         req.OrderId,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Method:          req.Method,
		TransactionID:   req.TransactionId,
		GatewayResponse: req.GatewayResponse,
	}

	payment, err := s.processPaymentUsecase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to process payment", "error", err)
		return nil, handleError(err)
	}

	return convertPaymentToProto(payment), nil
}

// UpdateShipping handles updating shipping information for an order
func (s *OrderServer) UpdateShipping(ctx context.Context, req *pb.UpdateShippingRequest) (*pb.ShippingResponse, error) {
	s.logger.Info("gRPC UpdateShipping request received", "order_id", req.OrderId)

	status, err := valueobject.ParseShippingStatus(req.Status)
	if err != nil {
		return nil, handleError(err)
	}

	var estimatedDelivery *time.Time
	if req.EstimatedDelivery != nil {
		t := req.EstimatedDelivery.AsTime()
		estimatedDelivery = &t
	}

	input := command.UpdateShippingInput{
		OrderID:           req.OrderId,
		Carrier:           req.Carrier,
		TrackingNumber:    req.TrackingNumber,
		Status:            status,
		EstimatedDelivery: estimatedDelivery,
		ShippingMethod:    req.ShippingMethod,
		ShippingCost:      req.ShippingCost,
		Notes:             req.Notes,
	}

	shipping, err := s.updateShippingUsecase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to update shipping", "error", err)
		return nil, handleError(err)
	}

	return convertShippingToProto(shipping), nil
}

// GetShipping handles retrieving shipping information for an order
func (s *OrderServer) GetShipping(ctx context.Context, req *pb.GetShippingRequest) (*pb.ShippingResponse, error) {
	s.logger.Info("gRPC GetShipping request received", "order_id", req.OrderId)

	// This should use a proper query usecase
	// We'd need to add a GetShippingUsecase to the OrderServer struct
	// For now, we'll modify this to retrieve the order first and then check its shipping ID

	order, err := s.getOrderUsecase.Execute(ctx, req.OrderId)
	if err != nil {
		s.logger.Error("Failed to get order for shipping information", "error", err)
		return nil, handleError(err)
	}

	if order.ShippingID == "" {
		return nil, status.Error(codes.NotFound, "Shipping information not found for this order")
	}

	// In a real implementation, this would call a ShippingUsecase
	// For now, return a placeholder response
	return &pb.ShippingResponse{
		Id:        "shipping-not-implemented",
		OrderId:   req.OrderId,
		Status:    "pending",
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

// Helper functions
func convertOrderToProto(order *entity.Order) *pb.OrderResponse {
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			Id:           item.ID,
			ProductId:    item.ProductID,
			Name:         item.Name,
			Sku:          item.SKU,
			Quantity:     int32(item.Quantity),
			Price:        item.Price,
			TotalPrice:   item.TotalPrice,
			CurrencyCode: item.CurrencyCode,
		}
	}

	discounts := make([]*pb.Discount, len(order.Discounts))
	for i, discount := range order.Discounts {
		discounts[i] = &pb.Discount{
			Code:        discount.Code,
			Description: discount.Description,
			Type:        discount.Type,
			Amount:      discount.Amount,
		}
	}

	response := &pb.OrderResponse{
		Id:             order.ID,
		UserId:         order.UserID,
		Items:          items,
		TotalAmount:    order.TotalAmount,
		Status:         order.Status.String(),
		PaymentId:      order.PaymentID,
		ShippingId:     order.ShippingID,
		Notes:          order.Notes,
		PromotionCodes: order.PromotionCodes,
		Discounts:      discounts,
		TaxAmount:      order.TaxAmount,
		CreatedAt:      timestamppb.New(order.CreatedAt),
		UpdatedAt:      timestamppb.New(order.UpdatedAt),
		Version:        int32(order.Version),
	}

	// Add shipping address
	response.ShippingAddress = &pb.Address{
		FirstName:    order.ShippingAddress.FirstName,
		LastName:     order.ShippingAddress.LastName,
		AddressLine1: order.ShippingAddress.AddressLine1,
		AddressLine2: order.ShippingAddress.AddressLine2,
		City:         order.ShippingAddress.City,
		State:        order.ShippingAddress.State,
		PostalCode:   order.ShippingAddress.PostalCode,
		Country:      order.ShippingAddress.Country,
		Phone:        order.ShippingAddress.Phone,
		Email:        order.ShippingAddress.Email,
	}

	// Add billing address
	response.BillingAddress = &pb.Address{
		FirstName:    order.BillingAddress.FirstName,
		LastName:     order.BillingAddress.LastName,
		AddressLine1: order.BillingAddress.AddressLine1,
		AddressLine2: order.BillingAddress.AddressLine2,
		City:         order.BillingAddress.City,
		State:        order.BillingAddress.State,
		PostalCode:   order.BillingAddress.PostalCode,
		Country:      order.BillingAddress.Country,
		Phone:        order.BillingAddress.Phone,
		Email:        order.BillingAddress.Email,
	}

	// Add optional timestamps
	if order.CompletedAt != nil {
		response.CompletedAt = timestamppb.New(*order.CompletedAt)
	}

	if order.CancelledAt != nil {
		response.CancelledAt = timestamppb.New(*order.CancelledAt)
	}

	return response
}

func convertPaymentToProto(payment *entity.Payment) *pb.PaymentResponse {
	response := &pb.PaymentResponse{
		Id:              payment.ID,
		OrderId:         payment.OrderID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Method:          payment.Method,
		Status:          string(payment.Status),
		TransactionId:   payment.TransactionID,
		GatewayResponse: payment.GatewayResponse,
		CreatedAt:       timestamppb.New(payment.CreatedAt),
		UpdatedAt:       timestamppb.New(payment.UpdatedAt),
	}

	if payment.CompletedAt != nil {
		response.CompletedAt = timestamppb.New(*payment.CompletedAt)
	}

	if payment.FailedAt != nil {
		response.FailedAt = timestamppb.New(*payment.FailedAt)
	}

	return response
}

func convertShippingToProto(shipping *entity.Shipping) *pb.ShippingResponse {
	response := &pb.ShippingResponse{
		Id:             shipping.ID,
		OrderId:        shipping.OrderID,
		Carrier:        shipping.Carrier,
		TrackingNumber: shipping.TrackingNumber,
		Status:         string(shipping.Status),
		ShippingMethod: shipping.ShippingMethod,
		ShippingCost:   shipping.ShippingCost,
		Notes:          shipping.Notes,
		CreatedAt:      timestamppb.New(shipping.CreatedAt),
		UpdatedAt:      timestamppb.New(shipping.UpdatedAt),
	}

	if shipping.EstimatedDelivery != nil {
		response.EstimatedDelivery = timestamppb.New(*shipping.EstimatedDelivery)
	}

	if shipping.ShippedAt != nil {
		response.ShippedAt = timestamppb.New(*shipping.ShippedAt)
	}

	if shipping.DeliveredAt != nil {
		response.DeliveredAt = timestamppb.New(*shipping.DeliveredAt)
	}

	return response
}

func createListOrdersResponse(orders []*entity.Order, total, page, pageSize int) *pb.ListOrdersResponse {
	protoOrders := make([]*pb.OrderResponse, len(orders))
	for i, order := range orders {
		protoOrders[i] = convertOrderToProto(order)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &pb.ListOrdersResponse{
		Total:      int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: int32(totalPages),
		Orders:     protoOrders,
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
		statusCode = codes.FailedPrecondition
		message = "Invalid order status transition"
	case errors.Is(err, entity.ErrOrderCancelled):
		statusCode = codes.FailedPrecondition
		message = "Order is already cancelled"
	case errors.Is(err, entity.ErrOrderCompleted):
		statusCode = codes.FailedPrecondition
		message = "Order is already completed"
	case errors.Is(err, entity.ErrPaymentNotFound):
		statusCode = codes.NotFound
		message = "Payment not found"
	case errors.Is(err, entity.ErrPaymentFailed):
		statusCode = codes.FailedPrecondition
		message = "Payment failed"
	case errors.Is(err, entity.ErrShippingNotFound):
		statusCode = codes.NotFound
		message = "Shipping not found"
	case errors.Is(err, entity.ErrItemNotFound):
		statusCode = codes.NotFound
		message = "Item not found in order"
	case errors.Is(err, entity.ErrInsufficientStock):
		statusCode = codes.ResourceExhausted
		message = "Insufficient stock for item"
	case errors.Is(err, entity.ErrInternalServerError):
		statusCode = codes.Internal
		message = "Internal server error"
	default:
		statusCode = codes.Unknown
		message = "Unknown error"
	}

	return status.Error(statusCode, message)
}
