package errors

import (
	"fmt"
	"strings"
)

type ErrorType string

const (
	ValidationErrorType ErrorType = "VALIDATION_ERROR"
	NotFoundErrorType   ErrorType = "NOT_FOUND"
	UnauthorizedType    ErrorType = "UNAUTHORIZED"
	InternalServerType  ErrorType = "INTERNAL_SERVER_ERROR"
	ConflictType        ErrorType = "CONFLICT"
	BadRequestType      ErrorType = "BAD_REQUEST"
	InvalidCredentials  ErrorType = "INVALID_CREDENTIALS"
)

// AppError is the base error type for the application
type AppError struct {
	Type     ErrorType              `json:"type"`
	Message  string                 `json:"message"`
	Field    string                 `json:"field,omitempty"`
	Code     int                    `json:"code"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Cause    error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewError creates a new AppError
func NewError(errType ErrorType, code int, message string, opts ...ErrorOption) *AppError {
	err := &AppError{
		Type:    errType,
		Message: message,
		Code:    code,
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

// ErrorOption is a function that configures an AppError
type ErrorOption func(*AppError)

// WithField adds a field to the error
func WithField(field string) ErrorOption {
	return func(e *AppError) {
		e.Field = field
	}
}

// WithMetadata adds metadata to the error
func WithMetadata(metadata map[string]interface{}) ErrorOption {
	return func(e *AppError) {
		e.Metadata = metadata
	}
}

// WithCause adds a cause to the error
func WithCause(cause error) ErrorOption {
	return func(e *AppError) {
		e.Cause = cause
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []*AppError

func (e ValidationErrors) Error() string {
	var errMessages []string
	for _, err := range e {
		errMessages = append(errMessages, err.Error())
	}
	return strings.Join(errMessages, "; ")
}

// Common error constructors
func NewValidationError(field, message string) *AppError {
	return NewError(ValidationErrorType, 400, message, WithField(field))
}

func NewNotFoundError(entity, value string) *AppError {
	return NewError(NotFoundErrorType, 404, fmt.Sprintf("%s not found: %s", entity, value))
}

func NewUnauthorizedError(message string) *AppError {
	return NewError(UnauthorizedType, 401, message)
}

func NewInternalError(message string, cause error) *AppError {
	return NewError(InternalServerType, 500, message, WithCause(cause))
}

func NewConflictError(message string) *AppError {
	return NewError(ConflictType, 409, message)
}

func NewBadRequestError(message string) *AppError {
	return NewError(BadRequestType, 400, message)
}

// Wrap wraps an error with a message and returns an AppError
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return NewError(
			appErr.Type,
			appErr.Code,
			fmt.Sprintf("%s: %s", message, appErr.Message),
			WithCause(err),
			WithField(appErr.Field),
			WithMetadata(appErr.Metadata),
		)
	}

	return NewError(
		InternalServerType,
		500,
		fmt.Sprintf("%s: %s", message, err.Error()),
		WithCause(err),
	)
}

// Wrapf wraps an error with a formatted message and returns an AppError
func Wrapf(err error, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}
