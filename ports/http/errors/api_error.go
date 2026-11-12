package errors

import (
	stderrors "errors"
	"log"
	"net/http"

	"github.com/devbenho/luka-platform/pkg/errors"
	"github.com/gin-gonic/gin"
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

// MapErrorToHTTP converts a service error into the API error body. Client
// errors (4xx) keep their message; server errors get a generic message so
// internal details (database errors, causes) are never sent to clients.
func MapErrorToHTTP(err error) *APIError {
	var validationErr errors.ValidationErrors
	if stderrors.As(err, &validationErr) {
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
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) && appErr.Code >= 400 && appErr.Code < 500 {
		return &APIError{
			Success:  false,
			Status:   appErr.Code,
			Message:  appErr.Message,
			Type:     string(appErr.Type),
			Metadata: appErr.Metadata,
		}
	}
	return &APIError{
		Success: false,
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Type:    string(errors.InternalServerType),
	}
}

// Respond writes err as a JSON API error and logs server errors, which are
// otherwise hidden from the client.
func Respond(c *gin.Context, err error) {
	apiError := MapErrorToHTTP(err)
	if apiError.Status >= http.StatusInternalServerError {
		log.Printf("%s %s: %v", c.Request.Method, c.FullPath(), err)
	}
	c.JSON(apiError.Status, apiError)
}
