package valueobject

import "time"

// Timestamp represents a point in time in the domain
type Timestamp struct {
	time time.Time
}

// TimeProvider provides current time for the domain
type TimeProvider interface {
	Now() Timestamp
}

// NewTimestamp creates a new timestamp from a Go time.Time
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{time: t}
}

// Now creates a new timestamp with the current time
func Now() Timestamp {
	return Timestamp{time: time.Now()}
}

// Time returns the underlying Go time.Time
func (t Timestamp) Time() time.Time {
	return t.time
}

// Add adds a duration to the timestamp
func (t Timestamp) Add(d time.Duration) Timestamp {
	return Timestamp{time: t.time.Add(d)}
}

// Before returns true if this timestamp is before the other
func (t Timestamp) Before(other Timestamp) bool {
	return t.time.Before(other.time)
}

// After returns true if this timestamp is after the other
func (t Timestamp) After(other Timestamp) bool {
	return t.time.After(other.time)
}

// Equal returns true if this timestamp is equal to the other
func (t Timestamp) Equal(other Timestamp) bool {
	return t.time.Equal(other.time)
}

// IsZero returns true if this timestamp represents the zero value
func (t Timestamp) IsZero() bool {
	return t.time.IsZero()
}

// Format formats the timestamp according to the layout
func (t Timestamp) Format(layout string) string {
	return t.time.Format(layout)
}

// String returns the string representation of the timestamp
func (t Timestamp) String() string {
	return t.time.String()
}
