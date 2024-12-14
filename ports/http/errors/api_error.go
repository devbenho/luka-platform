package errors

import (
	"net/http"

	"github.com/devbenho/luka-platform/pkg/errors"
)

type APIError struct {
	Success  bool                   `json:"success"`
	Status   int                    `json:"status"`
	Message  string                 `json:"message"`
	Type     string                 `json:"type"`
	Field    string                 `json:"field,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(status int, message string) *APIError {
	return &APIError{Success: false, Status: status, Message: message}
}

func MapErrorToHTTP(err error) *APIError {
	if appErr, ok := err.(*errors.AppError); ok {
		return &APIError{
			Success:  false,
			Status:   appErr.Code,
			Message:  appErr.Message,
			Type:     string(appErr.Type),
			Field:    appErr.Field,
			Metadata: appErr.Metadata,
		}
	}

	// Default to internal server error for unknown error types
	return &APIError{
		Success: false,
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Type:    string(errors.InternalServerType),
	}
}
