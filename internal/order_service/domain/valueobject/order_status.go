package valueobject

// OrderStatus represents the possible states of an order
type OrderStatus string

const (
	// OrderStatusCreated represents a newly created order
	OrderStatusCreated OrderStatus = "CREATED"
	
	// OrderStatusPaid represents an order that has been paid
	OrderStatusPaid OrderStatus = "PAID"
	
	// OrderStatusProcessing represents an order that is being processed
	OrderStatusProcessing OrderStatus = "PROCESSING"
	
	// OrderStatusShipped represents an order that has been shipped
	OrderStatusShipped OrderStatus = "SHIPPED"
	
	// OrderStatusDelivered represents an order that has been delivered
	OrderStatusDelivered OrderStatus = "DELIVERED"
	
	// OrderStatusCancelled represents an order that has been cancelled
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// IsValid checks if the order status is a valid one
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusCreated, OrderStatusPaid, OrderStatusProcessing, 
		 OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled:
		return true
	}
	return false
}

// String returns the string representation of the order status
func (s OrderStatus) String() string {
	return string(s)
}
