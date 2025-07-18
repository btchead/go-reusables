package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type SlogAdapter struct {
	options *options
}

func (o *SlogAdapter) New(config Config, writer WriteSyncer) Logger {
	if writer == nil {
		writer = NewStdoutWriteSyncer()
	}

	// Set log level
	level := slog.LevelInfo
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}

	if config.Format == "console" {
		if config.Colored {
			handler = newColoredTextHandler(writer, handlerOpts)
		} else {
			handler = slog.NewTextHandler(writer, handlerOpts)
		}
	} else {
		handler = slog.NewJSONHandler(writer, handlerOpts)
	}

	logger := slog.New(handler)

	// Add app metadata if provided
	if o.options != nil {
		if o.options.appName != "" || o.options.appVersion != "" {
			attrs := make([]any, 0, 4)
			if o.options.appName != "" {
				attrs = append(attrs, "appName", o.options.appName)
			}
			if o.options.appVersion != "" {
				attrs = append(attrs, "appVersion", o.options.appVersion)
			}
			logger = logger.With(attrs...)
		}
	}

	return &slogLogger{logger: logger}
}

// slogLogger wraps slog.Logger to implement our Logger interface
type slogLogger struct {
	logger *slog.Logger
}

func (o *slogLogger) Debug(msg string, keysAndValues ...any) {
	attrs := make([]slog.Attr, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			attrs[i/2] = slog.Any(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	o.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs[:len(keysAndValues)/2]...)
}

func (o *slogLogger) Info(msg string, keysAndValues ...any) {
	attrs := make([]slog.Attr, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			attrs[i/2] = slog.Any(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	o.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs[:len(keysAndValues)/2]...)
}

func (o *slogLogger) Warn(msg string, keysAndValues ...any) {
	attrs := make([]slog.Attr, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			attrs[i/2] = slog.Any(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	o.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs[:len(keysAndValues)/2]...)
}

func (o *slogLogger) Error(msg string, keysAndValues ...any) {
	attrs := make([]slog.Attr, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			attrs[i/2] = slog.Any(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	o.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs[:len(keysAndValues)/2]...)
}

func (o *slogLogger) Fatal(msg string, keysAndValues ...any) {
	attrs := make([]slog.Attr, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			attrs[i/2] = slog.Any(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	o.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs[:len(keysAndValues)/2]...)
	os.Exit(1)
}

func (o *slogLogger) With(keysAndValues ...any) Logger {
	return &slogLogger{logger: o.logger.With(keysAndValues...)}
}

func (o *slogLogger) WithContext(ctx context.Context) Logger {
	return &slogLogger{logger: o.logger.With()}
}

// coloredTextHandler is a custom handler that adds colors to text output
type coloredTextHandler struct {
	*slog.TextHandler
	writer io.Writer
}

func newColoredTextHandler(w io.Writer, opts *slog.HandlerOptions) *coloredTextHandler {
	return &coloredTextHandler{
		TextHandler: slog.NewTextHandler(w, opts),
		writer:      w,
	}
}

func (o *coloredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Get level color
	levelColor := getLevelColor(r.Level)

	// Format time
	timeStr := r.Time.Format(time.RFC3339)

	// Build colored output
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("time=%s level=%s%s\033[0m msg=\"%s\"",
		timeStr, levelColor, r.Level.String(), r.Message))

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})

	buf.WriteString("\n")

	_, err := o.writer.Write([]byte(buf.String()))
	return err
}

func getLevelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\033[36m" // Cyan
	case slog.LevelInfo:
		return "\033[32m" // Green
	case slog.LevelWarn:
		return "\033[33m" // Yellow
	case slog.LevelError:
		return "\033[31m" // Red
	default:
		return ""
	}
}
