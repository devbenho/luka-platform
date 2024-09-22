package errors

import (
	"fmt"
)

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

func NewValidationError(field string, tag string, value string) ValidationError {
	return ValidationError{
		Field: field,
		Tag:   tag,
		Value: value,
	}
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	return fmt.Sprintf("Validation failed: %v", ve)
}
