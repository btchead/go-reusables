# ⚙️ Service Management Package

A comprehensive service lifecycle management package for Go applications that provides service registration, start/stop operations, and graceful shutdown handling.

## Features

- **Service Registration**: Register multiple services with unique names
- **Lifecycle Management**: Start and stop services individually or collectively
- **Graceful Shutdown**: Handle OS signals for clean application shutdown
- **Signal Handling**: Configurable signals for graceful and force shutdown
- **Concurrent Safe**: Thread-safe operations with proper synchronization
- **Status Monitoring**: Query service states and get comprehensive status information
- **Error Handling**: Robust error handling with detailed error messages
- **Configurable Timeouts**: Customizable shutdown timeout periods
- **Structured Logging**: Integrated logging support compatible with go-reusables/log

## Installation

```bash
go get github.com/btchead/go-reusables/service
```

## Quick Start

### Basic Service Implementation

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/btchead/go-reusables/service"
)

// ExampleService implements the Service interface
type ExampleService struct {
    name string
}

func (o *ExampleService) Name() string {
    return o.name
}

func (o *ExampleService) Start(ctx context.Context) error {
    fmt.Printf("Starting service: %s\n", o.name)
    // Your service startup logic here
    return nil
}

func (o *ExampleService) Stop(ctx context.Context) error {
    fmt.Printf("Stopping service: %s\n", o.name)
    // Your service cleanup logic here
    return nil
}

func main() {
    // Create service manager
    manager := service.NewManager(
        service.WithShutdownTimeout(30 * time.Second),
    )

    // Register services
    webService := &ExampleService{name: "web-server"}
    dbService := &ExampleService{name: "database"}
    
    manager.Register(webService)
    manager.Register(dbService)

    // Run with graceful shutdown
    ctx := context.Background()
    if err := manager.RunWithGracefulShutdown(ctx); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Manual Service Management

```go
// Create manager
manager := service.NewManager()

// Register services
manager.Register(&MyWebService{})
manager.Register(&MyDatabaseService{})

// Start all services
ctx := context.Background()
if err := manager.Start(ctx); err != nil {
    log.Fatal(err)
}

// Start specific service
if err := manager.StartService(ctx, "web-server"); err != nil {
    log.Printf("Failed to start web-server: %v", err)
}

// Check service status
if manager.IsRunning("web-server") {
    fmt.Println("Web server is running")
}

// Get all service statuses
statuses := manager.GetStatus()
for _, status := range statuses {
    fmt.Printf("Service %s: %v\n", status.Name, status.State)
}

// Stop specific service
if err := manager.StopService(ctx, "web-server"); err != nil {
    log.Printf("Failed to stop web-server: %v", err)
}

// Stop all services
if err := manager.Stop(ctx); err != nil {
    log.Printf("Error stopping services: %v", err)
}
```

## Configuration Options

### Shutdown Timeout

```go
manager := service.NewManager(
    service.WithShutdownTimeout(45 * time.Second),
)
```

### Custom Signal Handling

```go
import "syscall"

manager := service.NewManager(
    service.WithGracefulSignals(syscall.SIGTERM, syscall.SIGINT),
    service.WithForceSignals(syscall.SIGKILL, syscall.SIGQUIT),
)
```

### Logger Integration

The service package supports structured logging compatible with the `go-reusables/log` package:

```go
import (
    "github.com/go-reusables/log"
    "github.com/go-reusables/service"
)

// Create logger (compatible with go-reusables/log)
logger := log.NewLogger(log.ZeroLogType, log.Config{
    Level:   "info",
    Format:  "json",
}, os.Stdout)

// Create manager with logger
manager := service.NewManager(
    service.WithLogger(logger),
    service.WithShutdownTimeout(30 * time.Second),
)

// All service operations will now be logged
manager.Register(&MyService{})
manager.Start(ctx) // Logs service start/stop operations
```

**Custom Logger Implementation:**

```go
// Implement the Logger interface for custom logging
type MyLogger struct{}

func (l MyLogger) Debug(msg string, keysAndValues ...any) { /* implementation */ }
func (l MyLogger) Info(msg string, keysAndValues ...any)  { /* implementation */ }
func (l MyLogger) Warn(msg string, keysAndValues ...any)  { /* implementation */ }
func (l MyLogger) Error(msg string, keysAndValues ...any) { /* implementation */ }
func (l MyLogger) Fatal(msg string, keysAndValues ...any) { /* implementation */ }
func (l MyLogger) With(keysAndValues ...any) service.Logger { return l }
func (l MyLogger) WithContext(ctx context.Context) service.Logger { return l }

// Use custom logger
manager := service.NewManager(service.WithLogger(MyLogger{}))
```

## Service Interface

Services must implement the `Service` interface:

```go
type Service interface {
    Name() string                           // Unique service identifier
    Start(ctx context.Context) error        // Start the service
    Stop(ctx context.Context) error         // Stop the service
}
```

## Service States

The package tracks the following service states:

- `StateStopped`: Service is not running
- `StateStarting`: Service is in the process of starting
- `StateRunning`: Service is running normally
- `StateStopping`: Service is in the process of stopping
- `StateError`: Service encountered an error

## Error Handling

The package provides detailed error information:

```go
// Service registration errors
err := manager.Register(duplicateService)
if err != nil {
    // Handle duplicate service name error
}

// Service operation errors
err = manager.Start(ctx)
if err != nil {
    // Handle startup failures
}
```

## Graceful Shutdown

The manager handles graceful shutdown automatically:

1. **Signal Reception**: Listens for configured OS signals
2. **Graceful Stop**: Attempts to stop all services within timeout
3. **Force Stop**: If timeout exceeded, forces immediate shutdown
4. **Error Collection**: Aggregates and reports any shutdown errors

## Best Practices

1. **Service Dependencies**: Register services in dependency order (dependencies first)
2. **Error Handling**: Always handle errors from Start/Stop operations
3. **Context Usage**: Respect context cancellation in service implementations
4. **Resource Cleanup**: Ensure proper resource cleanup in Stop methods
5. **Unique Names**: Use descriptive, unique names for services
6. **Timeout Configuration**: Set appropriate shutdown timeouts for your services

## Thread Safety

All manager operations are thread-safe and can be called concurrently from multiple goroutines.

## License

MIT License - see [LICENSE](../LICENSE) file for details.