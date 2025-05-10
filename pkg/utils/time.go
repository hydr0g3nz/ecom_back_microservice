package utils

import "time"

// TimePtr returns a pointer to the given time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// NowPtr returns a pointer to the current time
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// ParseTimePtr parses a time string and returns a pointer to the resulting time
func ParseTimePtr(layout, value string) (*time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
