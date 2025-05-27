package model

import (
	"database/sql"  // Required for NullString or similar if needed, though direct string is often sufficient with GORM
	"encoding/json" // To handle marshaling/unmarshaling GatewayResponse
	"time"

	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
	"github.com/shopspring/decimal"
	// GORM library
)

// Payment represents the GORM model for a Payment
type Payment struct {
	ID                   string          `gorm:"type:uuid;primaryKey"` // Store uuid.UUID as string (UUID format)
	OrderID              string          `gorm:"type:uuid;index;not null"`
	UserID               string          `gorm:"type:uuid;index;not null"`
	Amount               decimal.Decimal `gorm:"type:decimal(18,2);not null"` // Store decimal.Decimal
	Status               string          `gorm:"not null"`
	PaymentMethod        string          `gorm:"not null"`
	GatewayTransactionID sql.NullString  `gorm:"index"`                // Use NullString for potentially nullable field
	Transactions         []Transaction   `gorm:"foreignKey:PaymentID"` // HasMany relationship
	CreatedAt            time.Time       `gorm:"not null"`
	UpdatedAt            time.Time       `gorm:"not null"`
	// DeletedAt gorm.DeletedAt `gorm:"index"` // Uncomment for soft delete
}

// TableName specifies the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// ToEntity converts a GORM Payment model to a domain entity
func (m *Payment) ToEntity() (*entity.Payment, error) {
	paymentID, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	orderID, err := uuid.Parse(m.OrderID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}

	transactions := make([]*entity.Transaction, len(m.Transactions))
	for i, tx := range m.Transactions {
		txEntity, err := tx.ToEntity()
		if err != nil {
			return nil, err // Or handle partial conversion
		}
		transactions[i] = txEntity
	}

	return &entity.Payment{
		ID:                   paymentID,
		OrderID:              orderID,
		UserID:               userID,
		Amount:               m.Amount,
		Status:               m.Status,
		PaymentMethod:        m.PaymentMethod,
		GatewayTransactionID: m.GatewayTransactionID.String,
		Transactions:         transactions,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}, nil
}

// NewPaymentModel creates a new GORM Payment model from a domain entity
func NewPaymentModel(entity *entity.Payment) *Payment {
	transactions := make([]Transaction, len(entity.Transactions))
	for i, txEntity := range entity.Transactions {
		transactions[i] = *NewTransactionModel(txEntity)
	}

	return &Payment{
		ID:                   entity.ID.String(),
		OrderID:              entity.OrderID.String(),
		UserID:               entity.UserID.String(),
		Amount:               entity.Amount,
		Status:               entity.Status,
		PaymentMethod:        entity.PaymentMethod,
		GatewayTransactionID: sql.NullString{String: entity.GatewayTransactionID, Valid: entity.GatewayTransactionID != ""},
		Transactions:         transactions,
		CreatedAt:            entity.CreatedAt,
		UpdatedAt:            entity.UpdatedAt,
	}
}

// PaymentMethod represents the GORM model for a Payment Method
type PaymentMethod struct {
	ID            string    `gorm:"type:uuid;primaryKey"` // Store uuid.UUID as string
	UserID        string    `gorm:"type:uuid;index;not null"`
	Type          string    `gorm:"not null"`
	TokenizedData string    `gorm:"type:text;not null"` // Store tokenized data as text
	IsDefault     bool      `gorm:"default:false;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
	// DeletedAt gorm.DeletedAt `gorm:"index"` // Uncomment for soft delete
}

// TableName specifies the table name for the PaymentMethod model
func (PaymentMethod) TableName() string {
	return "payment_methods"
}

// ToEntity converts a GORM PaymentMethod model to a domain entity
func (m *PaymentMethod) ToEntity() (*entity.PaymentMethod, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}

	return &entity.PaymentMethod{
		ID:            id,
		UserID:        userID,
		Type:          m.Type,
		TokenizedData: m.TokenizedData,
		IsDefault:     m.IsDefault,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}

// NewPaymentMethodModel creates a new GORM PaymentMethod model from a domain entity
func NewPaymentMethodModel(entity *entity.PaymentMethod) *PaymentMethod {
	return &PaymentMethod{
		ID:            entity.ID.String(),
		UserID:        entity.UserID.String(),
		Type:          entity.Type,
		TokenizedData: entity.TokenizedData,
		IsDefault:     entity.IsDefault,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}

// Transaction represents the GORM model for a Transaction
type Transaction struct {
	ID              string          `gorm:"type:uuid;primaryKey"`     // Store uuid.UUID as string
	PaymentID       string          `gorm:"type:uuid;index;not null"` // Foreign key to Payment
	Type            string          `gorm:"not null"`
	Amount          decimal.Decimal `gorm:"type:decimal(18,2);not null"`
	Status          string          `gorm:"not null"`
	GatewayResponse []byte          `gorm:"type:jsonb"` // Store GatewayResponse as JSONB (or text for JSON)
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`
	// DeletedAt gorm.DeletedAt `gorm:"index"` // Uncomment for soft delete

	// Relationship to Payment (belongs to) - GORM automatically handles this with the foreign key
	// Payment Payment `gorm:"foreignKey:PaymentID"`
}

// TableName specifies the table name for the Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// ToEntity converts a GORM Transaction model to a domain entity
func (m *Transaction) ToEntity() (*entity.Transaction, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	paymentID, err := uuid.Parse(m.PaymentID)
	if err != nil {
		return nil, err
	}

	var gatewayResponse interface{}
	if len(m.GatewayResponse) > 0 {
		// Attempt to unmarshal the JSON stored in GatewayResponse
		if err := json.Unmarshal(m.GatewayResponse, &gatewayResponse); err != nil {
			// Handle the unmarshalling error, maybe return a partial entity or an error
			// For now, returning the error
			return nil, err
		}
	}

	return &entity.Transaction{
		ID:              id,
		PaymentID:       paymentID,
		Type:            m.Type,
		Amount:          m.Amount,
		Status:          m.Status,
		GatewayResponse: gatewayResponse,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}, nil
}

// NewTransactionModel creates a new GORM Transaction model from a domain entity
func NewTransactionModel(entity *entity.Transaction) *Transaction {
	var gatewayResponseJSON []byte
	if entity.GatewayResponse != nil {
		// Attempt to marshal the interface{} into JSON
		jsonBytes, err := json.Marshal(entity.GatewayResponse)
		if err == nil {
			gatewayResponseJSON = jsonBytes
		} else {
			// Handle marshalling error - perhaps log it or store nil/empty JSON
			// For now, if marshalling fails, gatewayResponseJSON will be nil
			// A more robust implementation might handle specific error cases
		}
	}

	return &Transaction{
		ID:              entity.ID.String(),
		PaymentID:       entity.PaymentID.String(),
		Type:            entity.Type,
		Amount:          entity.Amount,
		Status:          entity.Status,
		GatewayResponse: gatewayResponseJSON,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

// Assuming the entity package exists and contains the definitions
// package entity
// import (
// 	"time"
// 	"github.com/google/uuid"
// 	"github.com/shopspring/decimal"
// )
// type Payment struct { ... }
// type PaymentMethod struct { ... }
// type Transaction struct { ... }
// type GatewayResponse struct { ... } // If needed as a specific type for marshaling/unmarshaling
