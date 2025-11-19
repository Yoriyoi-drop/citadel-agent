package utils

// Response helpers for API responses
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}) Response {
	return Response{
		Status: "success",
		Data:   data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string) Response {
	return Response{
		Status: "error",
		Error:  message,
	}
}

// MessageResponse creates a message-only response
func MessageResponse(message string) Response {
	return Response{
		Status:  "success",
		Message: message,
	}
}