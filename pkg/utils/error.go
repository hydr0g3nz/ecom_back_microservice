package utils

import "fmt"

type ErrorBuilder struct {
	service string
}

func NewErrorBuilder(service string) *ErrorBuilder {
	return &ErrorBuilder{service: service}
}
func (e *ErrorBuilder) Err(err error) error {
	return fmt.Errorf("%s: %w", e.service, err)
}
