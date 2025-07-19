# 📝 Logger Package

Unified logging interface supporting zerolog and slog backends.

## Usage

```go
config := log.Config{
    Level:   "debug",
    Format:  "console", 
    Colored: true,
}

logger := log.NewLogger(
    log.ZeroLogType, 
    config, 
    os.Stdout,
    log.WithAppName("my-app"),
    log.WithAppVersion("1.0.0"),
)

logger.Info("Message", "key", "value")
```

## Config

```go
type Config struct {
    Level   string // debug, info, warn, error
    Format  string // json, console
    Colored bool   // colored console output
}
```

## Logger Types

- `log.ZeroLogType` - [zerolog](https://github.com/rs/zerolog) backend
- `log.SlogType` - [slog](https://pkg.go.dev/log/slog) backend

## Options

- `log.WithAppName(name)` - adds app name to all logs
- `log.WithAppVersion(version)` - adds app version to all logs