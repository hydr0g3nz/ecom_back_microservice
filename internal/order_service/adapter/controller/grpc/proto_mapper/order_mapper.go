package proto_mapper

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/adapter/controller/grpc/proto"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
)

// CreateOrderInputFromProto converts a gRPC CreateOrderRequest to a CreateOrderInput DTO
func CreateOrderInputFromProto(req *pb.CreateOrderRequest) dto.CreateOrderInput {
	// Convert items
	items := make([]dto.OrderItemDTO, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.OrderItemDTO{
			ProductID:    item.ProductId,
			Name:         item.Name,
			SKU:          item.Sku,
			Quantity:     int(item.Quantity),
			Price:        item.Price,
			TotalPrice:   item.Price * float64(item.Quantity),
			CurrencyCode: item.CurrencyCode,
		}
	}

	// Convert shipping address
	shippingAddress := dto.AddressDTO{
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

	// Convert billing address
	billingAddress := dto.AddressDTO{
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

	return dto.CreateOrderInput{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
		Notes:           req.Notes,
		PromotionCodes:  req.PromotionCodes,
	}
}

// UpdateOrderInputFromProto converts a gRPC UpdateOrderRequest to an UpdateOrderInput DTO
func UpdateOrderInputFromProto(req *pb.UpdateOrderRequest) dto.UpdateOrderInput {
	var updateInput dto.UpdateOrderInput

	// Notes
	if req.Notes != "" {
		notes := req.Notes
		updateInput.Notes = &notes
	}

	// Shipping address
	if req.ShippingAddress != nil {
		address := dto.AddressDTO{
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
		updateInput.ShippingAddress = &address
	}

	// Billing address
	if req.BillingAddress != nil {
		address := dto.AddressDTO{
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
		updateInput.BillingAddress = &address
	}

	return updateInput
}

// ProcessPaymentInputFromProto converts a gRPC ProcessPaymentRequest to a ProcessPaymentInput DTO
func ProcessPaymentInputFromProto(req *pb.ProcessPaymentRequest) dto.ProcessPaymentInput {
	return dto.ProcessPaymentInput{
		OrderID:         req.OrderId,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Method:          req.Method,
		TransactionID:   req.TransactionId,
		GatewayResponse: req.GatewayResponse,
	}
}

// OrderResponseFromDTO converts an OrderDTO to a gRPC OrderResponse
func OrderResponseFromDTO(orderDTO dto.OrderDTO) *pb.OrderResponse {
	items := make([]*pb.OrderItem, len(orderDTO.Items))
	for i, item := range orderDTO.Items {
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

	discounts := make([]*pb.Discount, len(orderDTO.Discounts))
	for i, discount := range orderDTO.Discounts {
		discounts[i] = &pb.Discount{
			Code:        discount.Code,
			Description: discount.Description,
			Type:        discount.Type,
			Amount:      discount.Amount,
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, orderDTO.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, orderDTO.UpdatedAt)

	response := &pb.OrderResponse{
		Id:             orderDTO.ID,
		UserId:         orderDTO.UserID,
		Items:          items,
		TotalAmount:    orderDTO.TotalAmount,
		Status:         orderDTO.Status,
		PaymentId:      orderDTO.PaymentID,
		ShippingId:     orderDTO.ShippingID,
		Notes:          orderDTO.Notes,
		PromotionCodes: orderDTO.PromotionCodes,
		Discounts:      discounts,
		TaxAmount:      orderDTO.TaxAmount,
		CreatedAt:      timestamppb.New(createdAt),
		UpdatedAt:      timestamppb.New(updatedAt),
		Version:        int32(orderDTO.Version),
	}

	// Add shipping address
	response.ShippingAddress = &pb.Address{
		FirstName:    orderDTO.ShippingAddress.FirstName,
		LastName:     orderDTO.ShippingAddress.LastName,
		AddressLine1: orderDTO.ShippingAddress.AddressLine1,
		AddressLine2: orderDTO.ShippingAddress.AddressLine2,
		City:         orderDTO.ShippingAddress.City,
		State:        orderDTO.ShippingAddress.State,
		PostalCode:   orderDTO.ShippingAddress.PostalCode,
		Country:      orderDTO.ShippingAddress.Country,
		Phone:        orderDTO.ShippingAddress.Phone,
		Email:        orderDTO.ShippingAddress.Email,
	}

	// Add billing address
	response.BillingAddress = &pb.Address{
		FirstName:    orderDTO.BillingAddress.FirstName,
		LastName:     orderDTO.BillingAddress.LastName,
		AddressLine1: orderDTO.BillingAddress.AddressLine1,
		AddressLine2: orderDTO.BillingAddress.AddressLine2,
		City:         orderDTO.BillingAddress.City,
		State:        orderDTO.BillingAddress.State,
		PostalCode:   orderDTO.BillingAddress.PostalCode,
		Country:      orderDTO.BillingAddress.Country,
		Phone:        orderDTO.BillingAddress.Phone,
		Email:        orderDTO.BillingAddress.Email,
	}

	// Add optional timestamps
	if orderDTO.CompletedAt != nil {
		completedAt, _ := time.Parse(time.RFC3339, *orderDTO.CompletedAt)
		response.CompletedAt = timestamppb.New(completedAt)
	}

	if orderDTO.CancelledAt != nil {
		cancelledAt, _ := time.Parse(time.RFC3339, *orderDTO.CancelledAt)
		response.CancelledAt = timestamppb.New(cancelledAt)
	}

	return response
}

// ListOrdersResponseFromDTO converts a ListOrdersOutput DTO to a gRPC ListOrdersResponse
func ListOrdersResponseFromDTO(output dto.ListOrdersOutput) *pb.ListOrdersResponse {
	protoOrders := make([]*pb.OrderResponse, len(output.Orders))
	for i, orderDTO := range output.Orders {
		protoOrders[i] = OrderResponseFromDTO(orderDTO)
	}

	return &pb.ListOrdersResponse{
		Total:      int32(output.Total),
		Page:       int32(output.Page),
		PageSize:   int32(output.PageSize),
		TotalPages: int32(output.TotalPages),
		Orders:     protoOrders,
	}
}

// PaymentResponseFromDTO converts a PaymentDTO to a gRPC PaymentResponse
func PaymentResponseFromDTO(paymentDTO dto.PaymentDTO) *pb.PaymentResponse {
	createdAt, _ := time.Parse(time.RFC3339, paymentDTO.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, paymentDTO.UpdatedAt)

	response := &pb.PaymentResponse{
		Id:              paymentDTO.ID,
		OrderId:         paymentDTO.OrderID,
		Amount:          paymentDTO.Amount,
		Currency:        paymentDTO.Currency,
		Method:          paymentDTO.Method,
		Status:          paymentDTO.Status,
		TransactionId:   paymentDTO.TransactionID,
		GatewayResponse: paymentDTO.GatewayResponse,
		CreatedAt:       timestamppb.New(createdAt),
		UpdatedAt:       timestamppb.New(updatedAt),
	}

	if paymentDTO.CompletedAt != nil {
		completedAt, _ := time.Parse(time.RFC3339, *paymentDTO.CompletedAt)
		response.CompletedAt = timestamppb.New(completedAt)
	}

	if paymentDTO.FailedAt != nil {
		failedAt, _ := time.Parse(time.RFC3339, *paymentDTO.FailedAt)
		response.FailedAt = timestamppb.New(failedAt)
	}

	return response
}

// ShippingResponseFromDTO converts a ShippingDTO to a gRPC ShippingResponse
func ShippingResponseFromDTO(shippingDTO dto.ShippingDTO) *pb.ShippingResponse {
	createdAt, _ := time.Parse(time.RFC3339, shippingDTO.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, shippingDTO.UpdatedAt)

	response := &pb.ShippingResponse{
		Id:             shippingDTO.ID,
		OrderId:        shippingDTO.OrderID,
		Carrier:        shippingDTO.Carrier,
		TrackingNumber: shippingDTO.TrackingNumber,
		Status:         shippingDTO.Status,
		ShippingMethod: shippingDTO.ShippingMethod,
		ShippingCost:   shippingDTO.ShippingCost,
		Notes:          shippingDTO.Notes,
		CreatedAt:      timestamppb.New(createdAt),
		UpdatedAt:      timestamppb.New(updatedAt),
	}

	if shippingDTO.EstimatedDelivery != nil {
		estimatedDelivery, _ := time.Parse(time.RFC3339, *shippingDTO.EstimatedDelivery)
		response.EstimatedDelivery = timestamppb.New(estimatedDelivery)
	}

	if shippingDTO.ShippedAt != nil {
		shippedAt, _ := time.Parse(time.RFC3339, *shippingDTO.ShippedAt)
		response.ShippedAt = timestamppb.New(shippedAt)
	}

	if shippingDTO.DeliveredAt != nil {
		deliveredAt, _ := time.Parse(time.RFC3339, *shippingDTO.DeliveredAt)
		response.DeliveredAt = timestamppb.New(deliveredAt)
	}

	return response
}
