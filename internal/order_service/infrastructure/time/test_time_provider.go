package time

import (
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/order_service/domain/valueobject"
)

// TestTimeProvider provides a fixed time for testing
type TestTimeProvider struct {
	fixedTime time.Time
}

// NewTestTimeProvider creates a new test time provider with the given fixed time
func NewTestTimeProvider(fixedTime time.Time) *TestTimeProvider {
	return &TestTimeProvider{
		fixedTime: fixedTime,
	}
}

// Now returns the fixed time as a domain timestamp
func (p *TestTimeProvider) Now() valueobject.Timestamp {
	return valueobject.NewTimestamp(p.fixedTime)
}

// SetFixedTime changes the fixed time
func (p *TestTimeProvider) SetFixedTime(fixedTime time.Time) {
	p.fixedTime = fixedTime
}

// AdvanceTime adds the specified duration to the fixed time
func (p *TestTimeProvider) AdvanceTime(duration time.Duration) {
	p.fixedTime = p.fixedTime.Add(duration)
}
