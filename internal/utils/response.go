package utils

import (
	"fmt"
)

type Response struct {
	Success  bool                   `json:"success"`
	Status   int                    `json:"status"`
	Message  string                 `json:"message"`
	Data     interface{}            `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func NewSuccessResponse(status int, message string, data interface{}) *Response {
	return &Response{
		Success: true,
		Status:  status,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(status int, message string, err string) *Response {
	return &Response{
		Success: false,
		Status:  status,
		Message: message,
		Error:   err,
	}
}

func NewValidationErrorResponse(errors interface{}) *Response {
	return &Response{
		Success: false,
		Status:  400,
		Message: "Validation error",
		Type:    "VALIDATION_ERROR",
		Metadata: map[string]interface{}{
			"errors": errors,
		},
	}
}

func NewNotFoundResponse(entity, id string) *Response {
	return &Response{
		Success: false,
		Status:  404,
		Message: fmt.Sprintf("%s not found: %s", entity, id),
		Type:    "NOT_FOUND",
	}
}

func NewUnauthorizedResponse(message string) *Response {
	return &Response{
		Success: false,
		Status:  401,
		Message: message,
		Type:    "UNAUTHORIZED",
	}
}

func NewForbiddenResponse(message string) *Response {
	return &Response{
		Success: false,
		Status:  403,
		Message: message,
		Type:    "FORBIDDEN",
	}
}

func NewInternalErrorResponse(message string) *Response {
	return &Response{
		Success: false,
		Status:  500,
		Message: message,
		Type:    "INTERNAL_SERVER_ERROR",
	}
}
