package apaas

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// LimiterOptions configures the request rate limiter.
type LimiterOptions struct {
	RequestsPerInterval int
	Interval            time.Duration
	Burst               int
}

// DefaultLimiterOptions returns a conservative limiter aligned with the Node.js SDK defaults.
func DefaultLimiterOptions() LimiterOptions {
	return LimiterOptions{
		RequestsPerInterval: 5,
		Interval:            time.Second,
		Burst:               20,
	}
}

// RateLimiter wraps golang.org/x/time/rate limiter.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter constructs a rate limiter using the provided options.
func NewRateLimiter(opts LimiterOptions) *RateLimiter {
	if opts.RequestsPerInterval <= 0 {
		opts.RequestsPerInterval = DefaultLimiterOptions().RequestsPerInterval
	}
	if opts.Interval <= 0 {
		opts.Interval = DefaultLimiterOptions().Interval
	}
	if opts.Burst <= 0 {
		opts.Burst = DefaultLimiterOptions().Burst
	}

	limit := rate.Every(opts.Interval / time.Duration(opts.RequestsPerInterval))
	return &RateLimiter{
		limiter: rate.NewLimiter(limit, opts.Burst),
	}
}

// Do waits for the next available slot and executes the provided function.
func (r *RateLimiter) Do(ctx context.Context, fn func() error) error {
	if r == nil || r.limiter == nil {
		return fn()
	}

	if err := r.limiter.Wait(ctx); err != nil {
		return err
	}
	return fn()
}
