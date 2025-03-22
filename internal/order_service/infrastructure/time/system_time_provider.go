package time

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// SystemTimeProvider provides the current system time
type SystemTimeProvider struct{}

// NewSystemTimeProvider creates a new system time provider
func NewSystemTimeProvider() *SystemTimeProvider {
	return &SystemTimeProvider{}
}

// Now returns the current time as a domain timestamp
func (p *SystemTimeProvider) Now() valueobject.Timestamp {
	return valueobject.NewTimestamp(time.Now())
}
