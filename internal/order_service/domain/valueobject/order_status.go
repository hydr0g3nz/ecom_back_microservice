package valueobject

import (
	"errors"
	"strings"
)

type OrderStatus string

const (
	Created          OrderStatus = "created"
	Pending          OrderStatus = "pending"
	PaymentPending   OrderStatus = "payment_pending"
	PaymentCompleted OrderStatus = "payment_completed"
	PaymentFailed    OrderStatus = "payment_failed"
	Processing       OrderStatus = "processing"
	Shipped          OrderStatus = "shipped"
	Delivered        OrderStatus = "delivered"
	Completed        OrderStatus = "completed"
	Cancelled        OrderStatus = "cancelled"
	Refunded         OrderStatus = "refunded"
	Returned         OrderStatus = "returned"
)

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) IsValid() bool {
	statuses := [...]OrderStatus{
		Created, Pending, PaymentPending, PaymentCompleted, PaymentFailed,
		Processing, Shipped, Delivered, Completed, Cancelled, Refunded, Returned,
	}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	status = strings.ToLower(status)
	if !OrderStatus(status).IsValid() {
		return "", errors.New("invalid order status")
	}
	return OrderStatus(status), nil
}

// IsValidTransition checks if transitioning from the current status to the new status is valid
func IsValidTransition(current, new OrderStatus) bool {
	// Define valid transitions
	validTransitions := map[OrderStatus][]OrderStatus{
		Created:          {Pending, PaymentPending, Cancelled},
		Pending:          {PaymentPending, Processing, Cancelled},
		PaymentPending:   {PaymentCompleted, PaymentFailed, Cancelled},
		PaymentFailed:    {PaymentPending, Cancelled},
		PaymentCompleted: {Processing, Shipped, Cancelled},
		Processing:       {Shipped, Cancelled},
		Shipped:          {Delivered, Returned},
		Delivered:        {Completed, Returned},
		Completed:        {Refunded, Returned},
		Returned:         {Refunded},
		Cancelled:        {}, // Terminal state
		Refunded:         {}, // Terminal state
	}

	// Check if the transition is valid
	validNextStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, validStatus := range validNextStatuses {
		if validStatus == new {
			return true
		}
	}

	return false
}
