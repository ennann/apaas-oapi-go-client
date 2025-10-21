package apaas

import (
	"errors"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "with request ID",
			err: &APIError{
				StatusCode: http.StatusBadRequest,
				Code:       "400001",
				Message:    "invalid parameter",
				RequestID:  "req-123",
				Endpoint:   "/api/test",
				Method:     "POST",
			},
			want: "api error [400]: code=400001, msg=invalid parameter, request_id=req-123, endpoint=POST /api/test",
		},
		{
			name: "without request ID",
			err: &APIError{
				StatusCode: http.StatusNotFound,
				Code:       "404001",
				Message:    "resource not found",
				Endpoint:   "/api/resource",
				Method:     "GET",
			},
			want: "api error [404]: code=404001, msg=resource not found, endpoint=GET /api/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"too many requests", http.StatusTooManyRequests, true},
		{"request timeout", http.StatusRequestTimeout, true},
		{"internal server error", http.StatusInternalServerError, true},
		{"bad gateway", http.StatusBadGateway, true},
		{"service unavailable", http.StatusServiceUnavailable, true},
		{"gateway timeout", http.StatusGatewayTimeout, true},
		{"bad request", http.StatusBadRequest, false},
		{"not found", http.StatusNotFound, false},
		{"unauthorized", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsRetryable(); got != tt.want {
				t.Errorf("APIError.IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "retryable API error",
			err:  &APIError{StatusCode: http.StatusTooManyRequests},
			want: true,
		},
		{
			name: "non-retryable API error",
			err:  &APIError{StatusCode: http.StatusBadRequest},
			want: false,
		},
		{
			name: "network error",
			err:  &NetworkError{Operation: "request", Err: errors.New("timeout")},
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "API error with code",
			err:  &APIError{Code: "ERR001"},
			want: "ERR001",
		},
		{
			name: "non-API error",
			err:  errors.New("some error"),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorCode(tt.err); got != tt.want {
				t.Errorf("ErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "API error with status",
			err:  &APIError{StatusCode: http.StatusNotFound},
			want: http.StatusNotFound,
		},
		{
			name: "non-API error",
			err:  errors.New("some error"),
			want: 0,
		},
		{
			name: "nil error",
			err:  nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCode(tt.err); got != tt.want {
				t.Errorf("StatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "username",
		Message: "required field",
	}
	want := "validation error: field=username, msg=required field"
	if got := err.Error(); got != want {
		t.Errorf("ValidationError.Error() = %v, want %v", got, want)
	}
}

func TestNetworkError_Error(t *testing.T) {
	err := &NetworkError{
		Operation: "connect",
		Err:       errors.New("connection refused"),
	}
	want := "network error during connect: connection refused"
	if got := err.Error(); got != want {
		t.Errorf("NetworkError.Error() = %v, want %v", got, want)
	}
}

func TestNetworkError_IsRetryable(t *testing.T) {
	err := &NetworkError{
		Operation: "request",
		Err:       errors.New("timeout"),
	}
	if !err.IsRetryable() {
		t.Error("NetworkError should be retryable")
	}
}
