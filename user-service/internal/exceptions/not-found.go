package exceptions

// NotFound is a struct that represents the not found exception.
type NotFound struct {
	Message string `json:"message"`
}

// NewNotFound is a function that creates a new not found exception.
func NewNotFound(message string) *NotFound {
	return &NotFound{
		Message: message,
	}

}

// Error is a function that returns the error message.
func (e *NotFound) Error() string {
	return e.Message
}
