package apaas

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return &APIError{StatusCode: 500}
		}
		return nil
	}

	ctx := context.Background()
	err := Retry(ctx, config, fn)

	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return &APIError{StatusCode: 400}
	}

	ctx := context.Background()
	err := Retry(ctx, config, fn)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetry_ContextCanceled(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return &APIError{StatusCode: 500}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Retry(ctx, config, fn)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context deadline exceeded, got %v", err)
	}
	if attempts < 1 {
		t.Errorf("expected at least 1 attempt, got %d", attempts)
	}
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return &NetworkError{Operation: "test", Err: errors.New("network error")}
	}

	ctx := context.Background()
	err := Retry(ctx, config, fn)

	if err == nil {
		t.Error("expected error after max retries, got nil")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // capped at MaxDelay
	}

	for _, tt := range tests {
		got := calculateBackoff(tt.attempt, config)
		if got != tt.want {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestCalculateBackoff_WithJitter(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Test multiple times to ensure jitter is applied
	delays := make(map[time.Duration]bool)
	for i := 0; i < 10; i++ {
		delay := calculateBackoff(1, config)
		delays[delay] = true
		
		// Should be at least 200ms (base) and at most 250ms (base + 25% jitter)
		if delay < 200*time.Millisecond || delay > 250*time.Millisecond {
			t.Errorf("delay out of expected range: %v", delay)
		}
	}
}
