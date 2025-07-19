# Go Reusables

A collection of reusable Go packages for common programming patterns and utilities.

## Packages

### 📝 [log](./log/) - Unified Logging Interface

A unified logging interface that supports multiple logging backends (zerolog, slog) with a consistent API.

**Features:**
- Multiple backends (zerolog, slog)
- Unified API across all backends
- WriteSyncer support for crash-safe logging
- Configurable formatting (JSON/console with colors)
- Built-in app metadata support
- Context propagation

### 🔄 [retrier](./retrier/) - Flexible Retry Logic

A comprehensive retry library with multiple backoff strategies, error classification, and execution modes.

**Features:**
- Multiple retry policies (fixed, exponential, linear backoff)
- Smart error classification for network/temporary errors
- Execution modes (sync, async, simple)
- Jitter support to prevent thundering herd
- Context timeout and cancellation handling
- Thread-safe atomic operations
- Functional options configuration

### ⚙️ [service](./service/) - Service Lifecycle Management

A comprehensive service management package that handles service registration, lifecycle operations, and graceful shutdown for Go applications.

**Features:**
- Service registration with unique names
- Individual and collective lifecycle management
- Graceful shutdown with OS signal handling
- Configurable signals for graceful and force shutdown
- Thread-safe operations with proper synchronization
- Status monitoring and comprehensive service information
- Robust error handling with detailed messages
- Customizable shutdown timeout periods
- Clean `NewService()` constructor to minimize boilerplate

## License

MIT License - see [LICENSE](./LICENSE) file for details.
