package identifier

import (
	"github.com/google/uuid"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// UUIDGenerator generates UUIDs for domain identifiers
type UUIDGenerator struct{}

// NewUUIDGenerator creates a new UUID generator
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// NewID generates a new UUID and returns it as a domain ID
func (g *UUIDGenerator) NewID() valueobject.ID {
	return valueobject.ID(uuid.New().String())
}
