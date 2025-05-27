package entity

import "errors"

// ข้อผิดพลาดที่เกี่ยวข้อง
var (
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrInvalidCallback      = errors.New("invalid gateway callback")
	ErrInvalidCallbackData  = errors.New("invalid callback data")
	ErrUnknownPaymentStatus = errors.New("unknown payment status")
	ErrCannotRefundPayment  = errors.New("cannot refund payment with current status")
	ErrRefundAmountTooLarge = errors.New("refund amount exceeds original payment amount")
	ErrUnauthorized         = errors.New("unauthorized action")
)
