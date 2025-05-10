// internal/order_service/domain/valueobject/order_status.go
package valueobject

import (
	"errors"
	"strings"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusFailed     OrderStatus = "failed"
	OrderStatusReturned   OrderStatus = "returned"
)

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) IsValid() bool {
	statuses := [...]OrderStatus{
		OrderStatusPending,
		OrderStatusProcessing,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCompleted,
		OrderStatusCancelled,
		OrderStatusFailed,
		OrderStatusReturned,
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

// IsTerminal returns whether this status represents a terminal state
func (s OrderStatus) IsTerminal() bool {
	terminalStatuses := [...]OrderStatus{
		OrderStatusCompleted,
		OrderStatusCancelled,
		OrderStatusReturned,
	}
	for _, status := range terminalStatuses {
		if s == status {
			return true
		}
	}
	return false
}
