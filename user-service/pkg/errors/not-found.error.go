package errors

import (
	"fmt"
)

type NotFound struct {
	Resource string
}

// Error returns the error message.
func (e *NotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFound creates a new NotFound error.
func NewNotFound(resource string) *NotFound {
	return &NotFound{Resource: resource}
}
