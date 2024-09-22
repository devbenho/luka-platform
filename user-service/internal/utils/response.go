package utils

type Response struct {
	Success bool        `json:"success"`
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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
