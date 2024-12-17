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
			Metadata: appErr.Metadata,
		}
	}
	if validationErr, ok := err.(errors.ValidationErrors); ok {
		return &APIError{
			Success: false,
			Status:  http.StatusBadRequest,
			Message: "Validation error",
			Type:    string(errors.ValidationErrorType),
			Metadata: map[string]interface{}{
				"errors": validationErr,
			},
		}
	}
	return &APIError{
		Success: false,
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Type:    string(errors.InternalServerType),
	}
}
