package apaas

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	ErrInvalidConfig      = errors.New("invalid client configuration")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrNotFound           = errors.New("resource not found")
	ErrBadRequest         = errors.New("bad request")
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout            = errors.New("request timeout")
	ErrCanceled           = errors.New("request canceled")
)

// APIError represents an error from the aPaaS API with detailed context.
type APIError struct {
	// StatusCode is the HTTP status code
	StatusCode int
	// Code is the API error code
	Code string
	// Message is the human-readable error message
	Message string
	// RequestID helps track the request for debugging
	RequestID string
	// Endpoint is the API endpoint that was called
	Endpoint string
	// Method is the HTTP method used
	Method string
	// Err is the underlying error, if any
	Err error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("api error [%d]: code=%s, msg=%s, request_id=%s, endpoint=%s %s",
			e.StatusCode, e.Code, e.Message, e.RequestID, e.Method, e.Endpoint)
	}
	return fmt.Sprintf("api error [%d]: code=%s, msg=%s, endpoint=%s %s",
		e.StatusCode, e.Code, e.Message, e.Method, e.Endpoint)
}

// Unwrap returns the underlying error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is potentially retryable.
func (e *APIError) IsRetryable() bool {
	// Rate limit, timeout, and server errors are retryable
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusRequestTimeout ||
		e.StatusCode == http.StatusInternalServerError ||
		e.StatusCode == http.StatusBadGateway ||
		e.StatusCode == http.StatusServiceUnavailable ||
		e.StatusCode == http.StatusGatewayTimeout
}

// ValidationError represents input validation errors.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field=%s, msg=%s", e.Field, e.Message)
}

// NetworkError represents network-level errors.
type NetworkError struct {
	Operation string
	Err       error
}

// Error implements the error interface.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the network error is retryable.
func (e *NetworkError) IsRetryable() bool {
	return true
}

// newAPIError creates a new API error with context.
func newAPIError(statusCode int, code, message, method, endpoint string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Method:     method,
		Endpoint:   endpoint,
		Err:        err,
	}
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}

	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.IsRetryable()
	}

	return false
}

// ErrorCode extracts the API error code from an error.
func ErrorCode(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return ""
}

// StatusCode extracts the HTTP status code from an error.
func StatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return 0
}
