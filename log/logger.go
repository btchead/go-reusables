package log

import (
	"context"
	"io"
)

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)
	With(keysAndValues ...any) Logger
	WithContext(ctx context.Context) Logger
}

type logger struct {
	config Config
}

type LoggerType string

const (
	ZeroLogType LoggerType = "zerolog"
	SlogType    LoggerType = "slog"
)

// NewLogger creates a new logger instance using the specified adapter
func NewLogger(loggerType LoggerType, config Config, writer io.Writer, opts ...Option) Logger {
	var adapter LoggerAdapter

	// Apply options
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	switch loggerType {
	case ZeroLogType:
		adapter = &ZerologAdapter{options: options}
	case SlogType:
		adapter = &SlogAdapter{options: options}
	default:
		adapter = &ZerologAdapter{options: options}
	}

	return adapter.New(config, writer)
}
