package apperror

import (
	"errors"
	"net/http"
	"strings"

	"github.com/devbenho/luka-platform/internal/shared/constants"
)

var (
	ErrEmailAlreadyExist         = errors.New(constants.ErrEmailAlreadyExist)
	ErrorUsernameAlreadyExist    = errors.New(constants.ErrUsernameAlreadyExist)
	ErrInvalidUserType           = errors.New(constants.ErrInvalidUserType)
	ErrInvalidPassword           = errors.New(constants.ErrInvalidPassword)
	ErrFailedGenerateJWT         = errors.New(constants.ErrFailedGenerateJWT)
	ErrInvalidIsActive           = errors.New(constants.ErrInvalidIsActive)
	ErrStatusValue               = errors.New(constants.ErrStatusValue)
	ErrFailedGetTokenInformation = errors.New(constants.ErrFailedGetTokenInformation)
)

type AppError struct {
	Code    int
	Err     error
	Message string
}

func Equals(err error, expectedErr error) bool {
	return strings.EqualFold(err.Error(), expectedErr.Error())
}

func (h AppError) Error() string {
	return h.Err.Error()
}

func BadRequest(err error) error {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: constants.ErrBadRequest,
		Err:     err,
	}
}

func InternalServerError(err error) error {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: constants.ErrInternalServerError,
		Err:     err,
	}
}

func Unauthorized(err error) error {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: constants.ErrUnauthorized,
		Err:     err,
	}
}

func Forbidden(err error) error {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: constants.ErrForbidden,
		Err:     err,
	}
}

func NotFound(err error) error {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: constants.ErrNotFound,
		Err:     err,
	}
}

func Conflict(err error) error {
	return &AppError{
		Code:    http.StatusConflict,
		Message: constants.ErrConflict,
		Err:     err,
	}
}

func GatewayTimeout(err error) error {
	return &AppError{
		Code:    http.StatusGatewayTimeout,
		Message: constants.ErrGatewayTimeout,
		Err:     err,
	}
}

func ServiceUnavailable(err error) error {
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Message: constants.ErrGatewayTimeout,
		Err:     err,
	}
}
