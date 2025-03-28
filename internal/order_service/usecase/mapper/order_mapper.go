package mapper

import (
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/entity"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/usecase/dto"
)

// ToOrderDTO converts an order entity to an order DTO
func ToOrderDTO(order *entity.Order) dto.OrderDTO {
	itemDTOs := make([]dto.OrderItemDTO, len(order.Items))
	for i, item := range order.Items {
		itemDTOs[i] = ToOrderItemDTO(item)
	}

	discountDTOs := make([]dto.DiscountDTO, len(order.Discounts))
	for i, discount := range order.Discounts {
		discountDTOs[i] = ToDiscountDTO(discount)
	}

	orderDTO := dto.OrderDTO{
		ID:              order.ID.String(),
		UserID:          order.UserID.String(),
		Items:           itemDTOs,
		TotalAmount:     order.TotalAmount,
		Status:          order.Status.String(),
		ShippingAddress: ToAddressDTO(order.ShippingAddress),
		BillingAddress:  ToAddressDTO(order.BillingAddress),
		PaymentID:       order.PaymentID.String(),
		ShippingID:      order.ShippingID.String(),
		Notes:           order.Notes,
		PromotionCodes:  order.PromotionCodes,
		Discounts:       discountDTOs,
		TaxAmount:       order.TaxAmount,
		CreatedAt:       order.CreatedAt.String(),
		UpdatedAt:       order.UpdatedAt.String(),
		Version:         order.Version,
	}

	if order.CompletedAt != nil {
		completedAt := order.CompletedAt.String()
		orderDTO.CompletedAt = &completedAt
	}

	if order.CancelledAt != nil {
		cancelledAt := order.CancelledAt.String()
		orderDTO.CancelledAt = &cancelledAt
	}

	return orderDTO
}

// ToOrderItemDTO converts an order item entity to an order item DTO
func ToOrderItemDTO(item entity.OrderItem) dto.OrderItemDTO {
	return dto.OrderItemDTO{
		ID:           item.ID.String(),
		ProductID:    item.ProductID.String(),
		Name:         item.Name,
		SKU:          item.SKU,
		Quantity:     item.Quantity,
		Price:        item.Price,
		TotalPrice:   item.TotalPrice,
		CurrencyCode: item.CurrencyCode,
	}
}

// ToAddressDTO converts an address entity to an address DTO
func ToAddressDTO(address entity.Address) dto.AddressDTO {
	return dto.AddressDTO{
		FirstName:    address.FirstName,
		LastName:     address.LastName,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		PostalCode:   address.PostalCode,
		Country:      address.Country,
		Phone:        address.Phone,
		Email:        address.Email,
	}
}

// ToDiscountDTO converts a discount entity to a discount DTO
func ToDiscountDTO(discount entity.Discount) dto.DiscountDTO {
	return dto.DiscountDTO{
		Code:        discount.Code,
		Description: discount.Description,
		Type:        discount.Type,
		Amount:      discount.Amount,
	}
}

// ToOrderItem converts an order item DTO to an order item entity
func ToOrderItem(itemDTO dto.OrderItemDTO, idGenerator valueobject.IDGenerator) (entity.OrderItem, error) {
	var id valueobject.ID
	if itemDTO.ID == "" {
		id = idGenerator.NewID()
	} else {
		id = valueobject.ID(itemDTO.ID)
	}

	productID := valueobject.ID(itemDTO.ProductID)

	return entity.NewOrderItem(
		id,
		productID,
		itemDTO.Name,
		itemDTO.SKU,
		itemDTO.Quantity,
		itemDTO.Price,
		itemDTO.CurrencyCode,
	)
}

// ToAddress converts an address DTO to an address entity
func ToAddress(addressDTO dto.AddressDTO) (entity.Address, error) {
	return entity.NewAddress(
		addressDTO.FirstName,
		addressDTO.LastName,
		addressDTO.AddressLine1,
		addressDTO.AddressLine2,
		addressDTO.City,
		addressDTO.State,
		addressDTO.PostalCode,
		addressDTO.Country,
		addressDTO.Phone,
		addressDTO.Email,
	)
}

// ToDiscount converts a discount DTO to a discount entity
func ToDiscount(discountDTO dto.DiscountDTO) (entity.Discount, error) {
	return entity.NewDiscount(
		discountDTO.Code,
		discountDTO.Description,
		discountDTO.Type,
		discountDTO.Amount,
	)
}

// CreateOrderFromDTO creates an order entity from DTOs
func CreateOrderFromDTO(
	input dto.CreateOrderInput,
	idGenerator valueobject.IDGenerator,
	timeProvider valueobject.TimeProvider,
) (*entity.Order, error) {
	// Convert order items
	items := make([]entity.OrderItem, len(input.Items))
	for i, itemDTO := range input.Items {
		item, err := ToOrderItem(itemDTO, idGenerator)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	// Convert addresses
	shippingAddress, err := ToAddress(input.ShippingAddress)
	if err != nil {
		return nil, err
	}

	billingAddress, err := ToAddress(input.BillingAddress)
	if err != nil {
		return nil, err
	}

	// Create the order
	return entity.NewOrder(
		idGenerator.NewID(),
		valueobject.ID(input.UserID),
		items,
		shippingAddress,
		billingAddress,
		input.Notes,
		input.PromotionCodes,
		timeProvider,
	)
}

// ToOrderListOutput converts a list of orders to a ListOrdersOutput DTO
func ToOrderListOutput(orders []*entity.Order, total, page, pageSize int) dto.ListOrdersOutput {
	orderDTOs := make([]dto.OrderDTO, len(orders))
	for i, order := range orders {
		orderDTOs[i] = ToOrderDTO(order)
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return dto.ListOrdersOutput{
		Orders:     orderDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
