package middlewares

import (
	"encoding/json"
	"net/http"

	"github.com/devbenho/luka-platform/pkg/errors"
)

// ErrorResponse is the structure for error responses.
type ErrorResponse struct {
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorMiddleware is a middleware for handling errors.
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function to recover from panics and handle errors
		defer func() {
			if err := recover(); err != nil {
				// Log the error (optional)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Check for errors set in the context (if using context-based error handling)
		if err, ok := r.Context().Value("error").(error); ok {
			var statusCode int
			var response ErrorResponse

			// Handle different error types
			switch {
			case err.Error() == "User not found":
				statusCode = http.StatusNotFound
				response.Message = "User not found"
			case isValidationError(err): // Custom function to check validation errors
				statusCode = http.StatusBadRequest
				response.Message = "Validation errors"
				response.Details = err // Include details if necessary
			default:
				statusCode = http.StatusInternalServerError
				response.Message = "An unexpected error occurred"
			}

			// Set the response status and encode the response
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
		}
	})
}

// isValidationError is a custom function to determine if an error is a validation error.
func isValidationError(err error) bool {
	// Logic to determine if the error is a validation error (e.g., type assertion)
	_, ok := err.(errors.ValidationErrors) // Assuming you have defined this type
	return ok
}
