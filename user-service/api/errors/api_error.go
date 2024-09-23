package errors

import (
	"fmt"
	"github.com/devbenho/bazar-user-service/internal/utils"
	"net/http"
)

type APIError struct {
	Code    int `json:"code"`
	Message any `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %v", e.Code, e.Message)
}
func NewAPIError(err error, code int) APIError {
	return APIError{Code: code, Message: err.Error()}
}

func InvalidRequestData(errors map[string]string) APIError {
	return APIError{Message: errors, Code: http.StatusBadRequest}
}

func InvalidJSON() APIError {
	return APIError{Message: fmt.Errorf("invalid JSON Request Data"), Code: http.StatusBadRequest}
}

type APIFunc func(http.ResponseWriter, *http.Request) error

func Make(apiFunc APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := apiFunc(w, r); err != nil {
			if apiErr, ok := err.(*APIError); ok {
				utils.WriteJson(w, apiErr.Code, apiErr)
			} else {
				errResp := map[string]any{"error": err.Error()}
				utils.WriteJson(w, http.StatusInternalServerError, errResp)
			}
		}
	}
}
