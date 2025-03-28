package dto

// PaymentDTO represents a payment data transfer object
type PaymentDTO struct {
	ID              string  `json:"id"`
	OrderID         string  `json:"order_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Method          string  `json:"method"`
	Status          string  `json:"status"`
	TransactionID   string  `json:"transaction_id"`
	GatewayResponse string  `json:"gateway_response"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	FailedAt        *string `json:"failed_at,omitempty"`
}

// ProcessPaymentInput represents the input for processing a payment
type ProcessPaymentInput struct {
	OrderID         string  `json:"order_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Method          string  `json:"method"`
	TransactionID   string  `json:"transaction_id"`
	GatewayResponse string  `json:"gateway_response"`
}

// PaymentStatusUpdateInput represents the input for updating a payment status
type PaymentStatusUpdateInput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
