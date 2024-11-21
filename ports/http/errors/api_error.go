package errors

import (
	"net/http"

	"github.com/devbenho/luka-platform/pkg/database"
	"github.com/devbenho/luka-platform/pkg/errors"
)

type APIError struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(status int, message string) *APIError {
	return &APIError{Success: false, Status: status, Message: message}
}

func MapErrorToHTTP(err error) *APIError {
	switch e := err.(type) {
	case *errors.ValidationError, errors.ValidationErrors:
		return NewAPIError(http.StatusBadRequest, e.Error())
	case *errors.NotFoundError:
		return NewAPIError(http.StatusNotFound, e.Error())
	case *errors.UnauthorizedError:
		return NewAPIError(http.StatusForbidden, e.Error())
	case *errors.InternalServerError:
		return NewAPIError(http.StatusInternalServerError, e.Error())
	case *errors.ConflictError, *database.DBDuplicateError:
		return NewAPIError(http.StatusConflict, e.Error())
	case *errors.BadRequestError:
		return NewAPIError(http.StatusBadRequest, e.Error())
	case *errors.InvalidCredentialsError:
		return NewAPIError(http.StatusUnauthorized, e.Error())
	default:
		return NewAPIError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}
