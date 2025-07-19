package retrier

import "time"

// Option configures retry behavior
type Option func(*config)

// WithMaxAttempts sets the maximum number of retry attempts
func WithMaxAttempts(attempts int) Option {
	return func(c *config) {
		c.maxAttempts = attempts
	}
}

// WithTimeout sets the total timeout for all retry attempts
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}

// WithRetryCondition sets the condition for retrying on errors
func WithRetryCondition(condition RetryCondition) Option {
	return func(c *config) {
		c.retryCondition = condition
	}
}

// WithOnRetry sets a callback function called on each retry attempt
func WithOnRetry(callback func(attempt int, err error, delay time.Duration)) Option {
	return func(c *config) {
		c.onRetry = callback
	}
}

// WithPolicy sets the retry policy
func WithPolicy(policy RetryPolicy) Option {
	return func(c *config) {
		c.policy = policy
	}
}

// WithFixedBackoff sets a fixed delay between retries
func WithFixedBackoff(delay time.Duration) Option {
	return WithPolicy(NewFixedBackoffPolicy(delay, 0))
}

// WithExponentialBackoff sets exponential backoff with base delay and multiplier
func WithExponentialBackoff(baseDelay time.Duration, multiplier float64) Option {
	return WithPolicy(NewExponentialBackoffPolicy(baseDelay, multiplier, 0, 0))
}

// WithLinearBackoff sets linear backoff with base delay
func WithLinearBackoff(baseDelay time.Duration) Option {
	return WithPolicy(NewLinearBackoffPolicy(baseDelay, 0))
}

// WithJitter adds jitter to the current policy
func WithJitter(jitterFactor float64) Option {
	return func(c *config) {
		if c.policy != nil {
			c.policy = NewJitterPolicy(c.policy, jitterFactor)
		}
	}
}
