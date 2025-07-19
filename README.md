# Go Reusables

A collection of reusable Go packages for common programming patterns and utilities.

## Packages

### 📝 [log](./log/) - Unified Logging Interface

A unified logging interface that supports multiple logging backends (zerolog, slog) with a consistent API.

**Features:**
- **Multiple backends**: Support for zerolog and slog
- **Unified API**: Consistent interface across all backends
- **WriteSyncer support**: Guaranteed log flushing for crash-safe logging
- **Configurable formatting**: JSON and console output with optional colors
- **App metadata**: Built-in support for app name and version
- **Context support**: Logger context propagation

## Installation

```bash
go get github.com/btchead/go-reusables/log
```

**Quick Start:**
```go
// Create a logger with zerolog backend
logger := log.NewLogger(log.ZeroLogType, log.Config{
    Level:   "info",
    Format:  "console",
    Colored: true,
}, os.Stdout,
    log.WithAppName("myapp"),
    log.WithAppVersion("v1.0.0"),
)

userLogger := logger.With("userID", 123, "session", "abc123")
userLogger.Info("Processing request")
```

### 🔄 [retrier](./retrier/) - Flexible Retry Logic

A comprehensive retry library with multiple backoff strategies, error classification, and execution modes.

**Features:**
- **Multiple retry policies**: Fixed, exponential, linear backoff with jitter support
- **Error classification**: Smart error detection for network, temporary, and custom errors
- **Execution modes**: Synchronous, asynchronous, and simple retry patterns
- **Context support**: Timeout and cancellation handling
- **Thread-safe**: Atomic operations for concurrent usage
- **Configurable**: Functional options pattern for flexible configuration

**Retry Policies:**
- **Fixed Backoff**: Constant delay between retries
- **Exponential Backoff**: Exponentially increasing delays (2x, 4x, 8x...)
- **Linear Backoff**: Linearly increasing delays (1x, 2x, 3x...)
- **Jitter Policy**: Adds randomization to prevent thundering herd
- **Custom Policies**: Implement your own retry logic
- **Conditional Policies**: Retry based on custom conditions

## Installation

```bash
go get github.com/btchead/go-reusables/retrier
```

**Quick Start:**
```go
// Simple retry with exponential backoff
err := retrier.Retry(ctx, func() error {
    return makeHTTPRequest()
}, 
    retrier.WithMaxAttempts(5),
    retrier.WithPolicy(retrier.NewExponentialBackoffPolicy(
        100*time.Millisecond, // base delay
        2.0,                  // multiplier
        0.1,                  // jitter factor
        5*time.Second,        // max delay
    )),
    retrier.WithRetryCondition(retrier.IsNetworkError),
)
```

**Error Classification:**
```go
// Built-in error conditions
retrier.RetryAlways        // Retry on any error (except context cancellation)
retrier.RetryOnAny         // Retry on any error
retrier.RetryNever         // Never retry

// Smart error detection
retrier.IsNetworkError     // Detects network-related errors
retrier.IsTemporaryError   // Detects temporary errors that should be retried

// Custom conditions
customCondition := func(err error) bool {
    return strings.Contains(err.Error(), "rate limit")
}
```

## License

MIT License - see [LICENSE](./LICENSE) file for details.
