package retrier

import (
	"math"
	"math/rand"
	"time"
)

// FixedBackoffPolicy implements a fixed delay retry policy
type FixedBackoffPolicy struct {
	delay       time.Duration
	maxAttempts int
}

// NewFixedBackoffPolicy creates a new fixed backoff policy
func NewFixedBackoffPolicy(delay time.Duration, maxAttempts int) *FixedBackoffPolicy {
	return &FixedBackoffPolicy{
		delay:       delay,
		maxAttempts: maxAttempts,
	}
}

func (p *FixedBackoffPolicy) ShouldRetry(attempt int, err error) bool {
	if p.maxAttempts <= 0 {
		return true // No limit
	}
	return attempt < p.maxAttempts
}

func (p *FixedBackoffPolicy) NextDelay(attempt int) time.Duration {
	return p.delay
}

// ExponentialBackoffPolicy implements exponential backoff with optional jitter and max delay
type ExponentialBackoffPolicy struct {
	baseDelay   time.Duration
	multiplier  float64
	jitter      float64
	maxDelay    time.Duration
	maxAttempts int
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(baseDelay time.Duration, multiplier float64, jitter float64, maxDelay time.Duration) *ExponentialBackoffPolicy {
	if multiplier <= 0 {
		multiplier = 2.0
	}
	return &ExponentialBackoffPolicy{
		baseDelay:  baseDelay,
		multiplier: multiplier,
		jitter:     jitter,
		maxDelay:   maxDelay,
	}
}

// WithMaxAttempts sets the maximum number of attempts for exponential backoff
func (p *ExponentialBackoffPolicy) WithMaxAttempts(maxAttempts int) *ExponentialBackoffPolicy {
	p.maxAttempts = maxAttempts
	return p
}

func (p *ExponentialBackoffPolicy) ShouldRetry(attempt int, err error) bool {
	if p.maxAttempts <= 0 {
		return true // No limit
	}
	return attempt < p.maxAttempts
}

func (p *ExponentialBackoffPolicy) NextDelay(attempt int) time.Duration {
	delay := time.Duration(float64(p.baseDelay) * math.Pow(p.multiplier, float64(attempt)))
	
	// Apply jitter if specified
	if p.jitter > 0 {
		jitterAmount := float64(delay) * p.jitter * (rand.Float64()*2 - 1) // Random between -jitter and +jitter
		delay = time.Duration(float64(delay) + jitterAmount)
	}
	
	// Cap at max delay if specified
	if p.maxDelay > 0 && delay > p.maxDelay {
		delay = p.maxDelay
	}
	
	// Ensure positive delay
	if delay < 0 {
		delay = p.baseDelay
	}
	
	return delay
}

// LinearBackoffPolicy implements linear backoff
type LinearBackoffPolicy struct {
	baseDelay   time.Duration
	increment   time.Duration
	maxDelay    time.Duration
	maxAttempts int
}

// NewLinearBackoffPolicy creates a new linear backoff policy
func NewLinearBackoffPolicy(baseDelay time.Duration, maxDelay time.Duration) *LinearBackoffPolicy {
	return &LinearBackoffPolicy{
		baseDelay: baseDelay,
		increment: baseDelay,
		maxDelay:  maxDelay,
	}
}

// WithIncrement sets the increment amount for each retry
func (p *LinearBackoffPolicy) WithIncrement(increment time.Duration) *LinearBackoffPolicy {
	p.increment = increment
	return p
}

// WithMaxAttempts sets the maximum number of attempts for linear backoff
func (p *LinearBackoffPolicy) WithMaxAttempts(maxAttempts int) *LinearBackoffPolicy {
	p.maxAttempts = maxAttempts
	return p
}

func (p *LinearBackoffPolicy) ShouldRetry(attempt int, err error) bool {
	if p.maxAttempts <= 0 {
		return true // No limit
	}
	return attempt < p.maxAttempts
}

func (p *LinearBackoffPolicy) NextDelay(attempt int) time.Duration {
	delay := p.baseDelay + time.Duration(attempt)*p.increment
	
	// Cap at max delay if specified
	if p.maxDelay > 0 && delay > p.maxDelay {
		delay = p.maxDelay
	}
	
	return delay
}

// JitterPolicy wraps another policy to add jitter
type JitterPolicy struct {
	policy RetryPolicy
	jitter float64
}

// NewJitterPolicy creates a new jitter policy that wraps another policy
func NewJitterPolicy(policy RetryPolicy, jitter float64) *JitterPolicy {
	if jitter < 0 {
		jitter = 0
	}
	if jitter > 1 {
		jitter = 1
	}
	return &JitterPolicy{
		policy: policy,
		jitter: jitter,
	}
}

func (p *JitterPolicy) ShouldRetry(attempt int, err error) bool {
	return p.policy.ShouldRetry(attempt, err)
}

func (p *JitterPolicy) NextDelay(attempt int) time.Duration {
	delay := p.policy.NextDelay(attempt)
	
	if p.jitter > 0 {
		// Add random jitter: delay * (1 ± jitter)
		jitterAmount := float64(delay) * p.jitter * (rand.Float64()*2 - 1)
		delay = time.Duration(float64(delay) + jitterAmount)
	}
	
	// Ensure positive delay
	if delay < 0 {
		delay = time.Millisecond
	}
	
	return delay
}

// ConditionalPolicy wraps another policy with additional retry conditions
type ConditionalPolicy struct {
	policy    RetryPolicy
	condition RetryCondition
}

// NewConditionalPolicy creates a policy that only retries if the condition is met
func NewConditionalPolicy(policy RetryPolicy, condition RetryCondition) *ConditionalPolicy {
	return &ConditionalPolicy{
		policy:    policy,
		condition: condition,
	}
}

func (p *ConditionalPolicy) ShouldRetry(attempt int, err error) bool {
	return p.condition(err) && p.policy.ShouldRetry(attempt, err)
}

func (p *ConditionalPolicy) NextDelay(attempt int) time.Duration {
	return p.policy.NextDelay(attempt)
}

// CustomPolicy allows for custom retry logic
type CustomPolicy struct {
	shouldRetryFunc func(attempt int, err error) bool
	nextDelayFunc   func(attempt int) time.Duration
}

// NewCustomPolicy creates a policy with custom retry and delay logic
func NewCustomPolicy(
	shouldRetry func(attempt int, err error) bool,
	nextDelay func(attempt int) time.Duration,
) *CustomPolicy {
	return &CustomPolicy{
		shouldRetryFunc: shouldRetry,
		nextDelayFunc:   nextDelay,
	}
}

func (p *CustomPolicy) ShouldRetry(attempt int, err error) bool {
	if p.shouldRetryFunc == nil {
		return true
	}
	return p.shouldRetryFunc(attempt, err)
}

func (p *CustomPolicy) NextDelay(attempt int) time.Duration {
	if p.nextDelayFunc == nil {
		return time.Second
	}
	return p.nextDelayFunc(attempt)
}