package entity

import (
	"time"

	"github.com/google/uuid"
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/valueobject"
	"github.com/shopspring/decimal"
)

// Payment เป็นโครงสร้างข้อมูลหลักสำหรับการชำระเงิน
type Payment struct {
	ID                   uuid.UUID        `json:"id"`
	OrderID              uuid.UUID        `json:"order_id"`
	UserID               uuid.UUID        `json:"user_id"`
	Amount               decimal.Decimal  `json:"amount"`
	Status               vo.PaymentStatus `json:"status"`
	PaymentMethod        vo.PaymentMethod `json:"payment_method"`
	GatewayTransactionID string           `json:"gateway_transaction_id"`
	Transactions         []*Transaction   `json:"transactions,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

// PaymentMethod เป็นโครงสร้างข้อมูลสำหรับวิธีการชำระเงิน
type PaymentMethod struct {
	ID            uuid.UUID        `json:"id"`
	UserID        uuid.UUID        `json:"user_id"`
	Type          vo.PaymentMethod `json:"type"`
	TokenizedData string           `json:"tokenized_data"`
	IsDefault     bool             `json:"is_default"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// Transaction เป็นโครงสร้างข้อมูลสำหรับธุรกรรมการชำระเงิน
type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	PaymentID       uuid.UUID       `json:"payment_id"`
	Type            string          `json:"type"`
	Amount          decimal.Decimal `json:"amount"`
	Status          string          `json:"status"`
	GatewayResponse interface{}     `json:"gateway_response"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// GatewayResponse เป็นโครงสร้างข้อมูลสำหรับการตอบกลับจาก payment gateway
type GatewayResponse struct {
	TransactionID string      `json:"transaction_id"`
	Status        string      `json:"status"`
	RawResponse   interface{} `json:"raw_response"`
}

// OrderDetails เป็นโครงสร้างข้อมูลสำหรับรายละเอียดคำสั่งซื้อ
type OrderDetails struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	TotalAmount   decimal.Decimal `json:"total_amount"`
	Status        string          `json:"status"`
	PaymentStatus string          `json:"payment_status"`
}

// DeterminePaymentMethodFromToken วิเคราะห์ประเภทวิธีการชำระเงินจาก token
func DeterminePaymentMethodFromToken(token string) string {
	// ตัวอย่างการวิเคราะห์ประเภทวิธีการชำระเงินจาก token
	// ในการใช้งานจริง อาจต้องตรวจสอบรูปแบบหรือข้อมูลเพิ่มเติม

	if len(token) < 4 {
		return "UNKNOWN"
	}

	// สมมติว่า token มีรูปแบบที่บ่งบอกประเภทได้
	prefix := token[0:4]

	switch prefix {
	case "ccrd":
		return "CREDIT_CARD"
	case "bktr":
		return "BANK_TRANSFER"
	case "wllt":
		return "WALLET"
	case "qrcd":
		return "QR_CODE"
	default:
		return "UNKNOWN"
	}
}
