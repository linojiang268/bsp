package errors

// APIError represents an error that can be sent in an error response.
type APIError struct {
	// Status represents the HTTP status code
	Status int `json:"-"`
	// ErrorCode is the code uniquely identifying an error
	ErrorCode string `json:"error_code"`
	// Message is the error message that may be displayed to end users
	Message string `json:"message"`
}

// Error returns the error message.
func (e APIError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code.
func (e APIError) StatusCode() int {
	return e.Status
}

// NewAPIError creates a new APIError with the given HTTP status code, error code, and additional message.
func NewAPIError(status int, code string, message string) *APIError {
	return &APIError{
		Status:    status,
		ErrorCode: code,
		Message:   message,
	}
}
