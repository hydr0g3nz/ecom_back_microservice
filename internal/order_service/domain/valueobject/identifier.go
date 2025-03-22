package valueobject

// ID represents a domain identifier
type ID string

// NewID creates a new domain identifier
type IDGenerator interface {
	NewID() ID
}

// String returns the string representation of the ID
func (id ID) String() string {
	return string(id)
}
