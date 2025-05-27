package vo

import "fmt"

// เกี่ยวกับสถานะการชำระเงิน
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusCompleted  PaymentStatus = "COMPLETED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
)

// เกี่ยวกับประเภทธุรกรรม
type TransactionType string

const (
	TransactionTypeCharge TransactionType = "CHARGE"
	TransactionTypeRefund TransactionType = "REFUND"
	TransactionTypeVoid   TransactionType = "VOID"
)

// เกี่ยวกับสถานะธุรกรรม
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusRefunded:
		return true
	}
	return false
}

func (s TransactionType) IsValid() bool {
	switch s {
	case TransactionTypeCharge, TransactionTypeRefund, TransactionTypeVoid:
		return true
	}
	return false
}

func (s TransactionStatus) IsValid() bool {
	switch s {
	case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed:
		return true
	}
	return false
}

func NewPaymentStatus(status string) (PaymentStatus, error) {
	s := PaymentStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid payment status: %s", status)
	}
	return s, nil
}

func NewTransactionType(ttype string) (TransactionType, error) {
	s := TransactionType(ttype)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid transaction type: %s", ttype)
	}
	return s, nil
}

func NewTransactionStatus(status string) (TransactionStatus, error) {
	s := TransactionStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid transaction status: %s", status)
	}
	return s, nil
}

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "Pending"
	case PaymentStatusProcessing:
		return "Processing"
	case PaymentStatusCompleted:
		return "Completed"
	case PaymentStatusFailed:
		return "Failed"
	case PaymentStatusRefunded:
		return "Refunded"
	default:
		return "Unknown"
	}
}

func (s TransactionType) String() string {
	switch s {
	case TransactionTypeCharge:
		return "Charge"
	case TransactionTypeRefund:
		return "Refund"
	case TransactionTypeVoid:
		return "Void"
	default:
		return "Unknown"
	}
}

func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "Pending"
	case TransactionStatusCompleted:
		return "Completed"
	case TransactionStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}
