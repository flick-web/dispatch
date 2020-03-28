package dispatch

import "errors"

// APIError is an error that contains status code information as well as error text.
type APIError struct {
	StatusCode int
	ErrorText  string
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
