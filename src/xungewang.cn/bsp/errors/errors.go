package errors

import (
	"github.com/go-ozzo/ozzo-validation"
	"net/http"
)

type (
	// Params is used to replace placeholders in an error template with the corresponding values.
	Params map[string]interface{}
)

// InternalServerError creates a new API error representing an internal server error (HTTP 500)
func InternalServerError(err error) *APIError {
	return NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
}

// NotFound creates a new API error representing a resource-not-found error (HTTP 404)
func NotFound(message string) *APIError {
	return NewAPIError(http.StatusNotFound, "NOT_FOUND", message)
}

// InvalidData converts a data validation error into an API error (HTTP 400)
func InvalidData(errs validation.Errors) *APIError {
	return NewAPIError(http.StatusBadRequest, "INVALID_DATA", errs.Error())
}

// SimpleInvalidData converts a plain error message into an API error (HTTP 400)
func SimpleInvalidData(message string) *APIError {
	return NewAPIError(http.StatusBadRequest, "INVALID_DATA", message)
}
