package log

import (
	"context"
	"io"
	"os"
)

// WriteSyncer is the interface implemented by writers that support syncing
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// syncWriter wraps an io.Writer to implement WriteSyncer
type syncWriter struct {
	io.Writer
}

func (w syncWriter) Sync() error {
	if syncer, ok := w.Writer.(WriteSyncer); ok {
		return syncer.Sync()
	}
	return nil
}

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


// NewWriteSyncer creates a WriteSyncer from an io.Writer
func NewWriteSyncer(w io.Writer) WriteSyncer {
	if ws, ok := w.(WriteSyncer); ok {
		return ws
	}
	return syncWriter{Writer: w}
}

// NewFileWriteSyncer creates a WriteSyncer for a file
func NewFileWriteSyncer(filename string) (WriteSyncer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// NewStdoutWriteSyncer creates a WriteSyncer for stdout
func NewStdoutWriteSyncer() WriteSyncer {
	return NewWriteSyncer(os.Stdout)
}

// NewStderrWriteSyncer creates a WriteSyncer for stderr
func NewStderrWriteSyncer() WriteSyncer {
	return NewWriteSyncer(os.Stderr)
}

// NewLogger creates a new logger instance using the specified adapter
func NewLogger(loggerType LoggerType, config Config, writer WriteSyncer, opts ...Option) Logger {
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
