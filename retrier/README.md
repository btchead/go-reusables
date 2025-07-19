# Retrier Package

Flexible retry package with multiple strategies, error classification, and execution modes.

## Features

- **Multiple Retry Policies**: Fixed, exponential, linear, and custom backoff strategies
- **Smart Error Classification**: Comprehensive network error detection
- **Execution Modes**: Sync, async, and simple error-only APIs
- **Jitter Support**: Prevent thundering herd problems
- **Context Support**: Timeout and cancellation handling
- **Thread-Safe**: Atomic operations for concurrent usage
- **Robust Network Detection**: Handles Go network error types and syscalls
- **Functional Options**: Clean, extensible configuration

## Quick Start

```go
import "github.com/go-reusables/retrier"

// Simple retry
err := retrier.Retry(ctx, func() error {
    return doSomething()
})

// Advanced configuration
result := retrier.Do(ctx, func() error { 
    return ioOperation()
    }, 
    retrier.WithExponentialBackoff(100*time.Millisecond, 2.0),
    retrier.WithMaxAttempts(5),
    retrier.WithJitter(0.1),
    retrier.WithRetryIf(retrier.IsNetworkError),
)
```

## Retry Policies

### Fixed Backoff
```go
retrier.WithFixedBackoff(500*time.Millisecond)
```

### Exponential Backoff
```go
retrier.WithExponentialBackoff(100*time.Millisecond, 2.0) // base=100ms, multiplier=2.0
```

### Linear Backoff
```go
retrier.WithLinearBackoff(100*time.Millisecond) // 100ms, 200ms, 300ms...
```

### Custom Policy
```go
policy := retrier.NewCustomPolicy(
    func(attempt int, err error) bool { return attempt < 5 },
    func(attempt int) time.Duration { return time.Second },
)
retrier.WithPolicy(policy)
```

## Error Classification

### Built-in Error Conditions

```go
// Retry only on network errors (connection refused, timeouts, DNS errors, etc.)
retrier.WithRetryIf(retrier.IsNetworkError)

// Retry on temporary errors (includes all network errors)
retrier.WithRetryIf(retrier.IsTemporaryError)

// Always retry (except context cancellation)
retrier.WithRetryIf(retrier.RetryAlways)

// Never retry
retrier.WithRetryIf(retrier.RetryNever)
```

### Network Error Detection

The `IsNetworkError` function detects:
- **Network operations**: `*net.OpError` (dial, connect, etc.)
- **DNS errors**: `*net.DNSError` (hostname resolution)
- **Address errors**: `*net.AddrError` (invalid addresses)
- **Syscall errors**: Connection refused, reset, timeout, unreachable
- **Context timeouts**: `context.DeadlineExceeded`

### Custom Conditions

```go
// Custom condition based on error content
retrier.WithRetryIf(func(err error) bool {
    return strings.Contains(err.Error(), "rate limit")
})

// Retry on specific HTTP status codes
retrier.WithRetryIf(func(err error) bool {
    var httpErr *HTTPError
    return errors.As(err, &httpErr) && httpErr.StatusCode >= 500
})
```

## Options

- `WithMaxAttempts(n)` - Maximum retry attempts
- `WithTimeout(duration)` - Total timeout for all attempts
- `WithJitter(factor)` - Add randomness (0.0-1.0)
- `WithOnRetry(callback)` - Retry notifications
- `WithRetryIf(condition)` - Error-based retry conditions

## Execution Modes

### Synchronous
```go
result := retrier.Do(ctx, fn, options...)
```

### Asynchronous
```go
resultChan := retrier.DoAsync(ctx, fn, options...)
result := <-resultChan
```

### Simple Error Return
```go
err := retrier.Retry(ctx, fn, options...)
```

## Result Information

The `Result` type provides detailed information about retry attempts:

```go
result := retrier.Do(ctx, fn, options...)

fmt.Printf("Success: %v\n", result.IsSuccess())
fmt.Printf("Attempts: %d\n", result.Attempts()) // Thread-safe
fmt.Printf("Duration: %v\n", result.Duration)
fmt.Printf("Error: %v\n", result.Error())
```

**Thread-Safety**: The `Attempts()` method uses atomic operations, making it safe to call from multiple goroutines (important for async usage).