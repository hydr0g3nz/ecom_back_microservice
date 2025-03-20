package valueobject

import (
	"errors"
	"strings"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

func (s PaymentStatus) String() string {
	return string(s)
}

func (s PaymentStatus) IsValid() bool {
	statuses := [...]PaymentStatus{
		PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted,
		PaymentStatusFailed, PaymentStatusRefunded, PaymentStatusCancelled,
	}
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func ParsePaymentStatus(status string) (PaymentStatus, error) {
	status = strings.ToLower(status)
	if !PaymentStatus(status).IsValid() {
		return "", errors.New("invalid payment status")
	}
	return PaymentStatus(status), nil
}
