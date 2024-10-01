package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s' with tag '%s with value %s'", e.Field, e.Tag, e.Value)
}

func NewValidationError(field, tag, value string) *ValidationError {
	return &ValidationError{Field: field, Tag: tag, Value: value}
}

type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	var errMessages []string
	for _, err := range e {
		errMessages = append(errMessages, err.Error())
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(errMessages, ", "))
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Entity string
	Field  string
	Value  string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found with %s %s", e.Entity, e.Field, e.Value)
}

// UnauthorizedError represents an unauthorized access error
type UnauthorizedError struct {
	Action string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized to perform action '%s'", e.Action)
}

// InternalServerError represents a generic internal server error
type InternalServerError struct {
	Message string
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("internal server error: %s", e.Message)
}

// ConflictError represents a conflict error
type ConflictError struct {
	Entity string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict", e.Entity)
}

// BadRequestError represents a bad request error
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request: %s", e.Message)
}

type InvalidCredentialsError struct{}

func (e *InvalidCredentialsError) Error() string {
	return "invalid credentials"
}
