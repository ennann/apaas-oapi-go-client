package apaas

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// RetryConfig configures the retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the initial backoff delay
	InitialDelay time.Duration
	// MaxDelay is the maximum backoff delay
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier
	Multiplier float64
	// Jitter enables random jitter to prevent thundering herd
	Jitter bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func() error

// Retry executes a function with exponential backoff retry logic.
func Retry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if the error is retryable
		if !IsRetryableError(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, config)

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	return lastErr
}

// calculateBackoff calculates the backoff delay for a given attempt.
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: initialDelay * (multiplier ^ attempt)
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter if enabled
	if config.Jitter {
		// Add random jitter between 0% and 25% of the delay
		jitter := rand.Float64() * delay * 0.25
		delay += jitter
	}

	return time.Duration(delay)
}
