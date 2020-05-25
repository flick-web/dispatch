package dispatch

import (
	"errors"
	"net/http"
)

// APIError is an error that contains status code information as well as error text.
type APIError struct {
	StatusCode int
	ErrorText  string
}

// NewAPIErrorFromStatus creates an APIError with text from an HTTP status code.
func NewAPIErrorFromStatus(statusCode int) *APIError {
	return &APIError{
		StatusCode: statusCode,
		ErrorText:  http.StatusText(statusCode),
	}
}

// NewAPIError returns an APIError from the given HTTP status code and error string.
func NewAPIError(statusCode int, errorText string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		ErrorText:  errorText,
	}
}

func (apiErr *APIError) Error() string {
	return apiErr.ErrorText
}

// ErrorBadRequest represents an error from a malformed request.
var ErrorBadRequest = errors.New("Bad request")

// ErrorNotFound represents a 404 error.
var ErrorNotFound = errors.New("Path not found")

// ErrorInternal represents some unexpected internal error.
var ErrorInternal = errors.New("Internal error")
