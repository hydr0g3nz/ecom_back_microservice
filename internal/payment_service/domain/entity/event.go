package entity

import (
	"time"

	"github.com/google/uuid"
)

type PaymentFailed struct {
	PaymentID uuid.UUID `json:"payment_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}
