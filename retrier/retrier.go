package retrier

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"syscall"
	"time"
)

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryPolicy defines how retries should be performed
type RetryPolicy interface {
	// ShouldRetry determines if a retry should be attempted
	ShouldRetry(attempt int, err error) bool
	// NextDelay calculates the delay before the next retry
	NextDelay(attempt int) time.Duration
}

// RetryCondition determines if an error should trigger a retry
type RetryCondition func(error) bool

// Result contains the result of a retry operation
type Result struct {
	attempts  atomic.Int64
	LastErr   error
	Success   bool
	Duration  time.Duration
	StartTime time.Time
}

// Attempts returns the number of attempts made (thread-safe)
func (o *Result) Attempts() int {
	return int(o.attempts.Load())
}

// String returns a string representation of the result
func (o *Result) String() string {
	if o.Success {
		return fmt.Sprintf("Success after %d attempts in %v", o.Attempts(), o.Duration)
	}
	return fmt.Sprintf("Failed after %d attempts in %v: %v", o.Attempts(), o.Duration, o.LastErr)
}

// IsSuccess returns true if the operation was successful
func (o *Result) IsSuccess() bool {
	return o.Success
}

// Error returns the last error if the operation failed
func (o *Result) Error() error {
	if o.Success {
		return nil
	}
	return o.LastErr
}

// config holds retry configuration
type config struct {
	maxAttempts    int
	timeout        time.Duration
	retryCondition RetryCondition
	onRetry        func(attempt int, err error, delay time.Duration)
	policy         RetryPolicy
}

// Common retry conditions
var (
	// RetryAlways always retries (except on context cancellation)
	RetryAlways = func(err error) bool {
		return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	}

	// RetryOnAny retries on any error
	RetryOnAny = func(err error) bool {
		return err != nil
	}

	// RetryNever never retries
	RetryNever = func(err error) bool {
		return false
	}
)

// IsNetworkError checks if an error is a network error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network operation errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// Check for address errors
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		return true
	}

	// Check for parse errors
	var parseErr *net.ParseError
	if errors.As(err, &parseErr) {
		return true
	}

	// Check for common syscall errors
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ETIMEDOUT) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}

	// Check for context timeout (often network related)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

// IsTemporaryError checks if an error is temporary and should be retried
func IsTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// First check if it's a network error - most network errors should be retried
	if IsNetworkError(err) {
		return true
	}

	// For non-network errors, check if error implements Temporary() method
	type temporary interface {
		Temporary() bool
	}

	if te, ok := err.(temporary); ok {
		return te.Temporary()
	}

	return false
}

// defaultConfig returns default retry configuration
func defaultConfig() *config {
	return &config{
		maxAttempts:    3,
		timeout:        30 * time.Second,
		retryCondition: RetryAlways,
		policy:         NewExponentialBackoffPolicy(100*time.Millisecond, 2.0, 0, 5*time.Second),
	}
}

// Do executes a function with retry logic
func Do(ctx context.Context, fn RetryableFunc, options ...Option) *Result {
	cfg := defaultConfig()
	for _, opt := range options {
		opt(cfg)
	}

	result := &Result{
		StartTime: time.Now(),
	}

	// Create context with timeout if specified
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}

	for attempt := 0; attempt < cfg.maxAttempts; attempt++ {
		result.attempts.Store(int64(attempt + 1))

		err := fn()
		if err == nil {
			result.Success = true
			result.Duration = time.Since(result.StartTime)
			return result
		}

		result.LastErr = err

		// Check if we should retry
		if !cfg.retryCondition(err) {
			break
		}

		// Check if policy allows retry
		if cfg.policy != nil && !cfg.policy.ShouldRetry(attempt, err) {
			break
		}

		// Don't retry on the last attempt
		if attempt == cfg.maxAttempts-1 {
			break
		}

		// Calculate delay
		var delay time.Duration
		if cfg.policy != nil {
			delay = cfg.policy.NextDelay(attempt)
		}

		// Call retry callback
		if cfg.onRetry != nil {
			cfg.onRetry(attempt+1, err, delay)
		}

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			result.LastErr = ctx.Err()
			result.Duration = time.Since(result.StartTime)
			return result
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	result.Duration = time.Since(result.StartTime)
	return result
}

// DoAsync executes a function with retry logic asynchronously
func DoAsync(ctx context.Context, fn RetryableFunc, options ...Option) <-chan *Result {
	resultChan := make(chan *Result, 1)

	go func() {
		defer close(resultChan)
		result := Do(ctx, fn, options...)
		resultChan <- result
	}()

	return resultChan
}

// Retry is a convenience function that returns only the error
func Retry(ctx context.Context, fn RetryableFunc, options ...Option) error {
	result := Do(ctx, fn, options...)
	return result.Error()
}
