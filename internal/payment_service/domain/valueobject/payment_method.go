package vo

import "fmt"

type PaymentMethod string

const (
	CreditCard   PaymentMethod = "CREDIT_CARD"
	BankTransfer PaymentMethod = "BANK_TRANSFER"
	Wallet       PaymentMethod = "WALLET"
	QRCode       PaymentMethod = "QR_CODE"
	Unknown      PaymentMethod = "UNKNOWN"
)

func NewPaymentMethod(pmethod string) (PaymentMethod, error) {
	switch pmethod {
	case "ccrd":
		return CreditCard, nil
	case "bktr":
		return BankTransfer, nil
	case "wllt":
		return Wallet, nil
	case "qrcd":
		return QRCode, nil
	default:
		return Unknown, fmt.Errorf("invalid payment method: %s", pmethod)
	}
}

func (p PaymentMethod) IsValid() bool {
	switch p {
	case CreditCard, BankTransfer, Wallet, QRCode:
		return true
	default:
		return false
	}
}
