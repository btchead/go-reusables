package log

import (
	"context"

	"github.com/rs/zerolog"
)

type ZerologAdapter struct {
	options *options
}

// New creates a new zerolog-based logger
func (o *ZerologAdapter) New(config Config, writer WriteSyncer) Logger {
	if writer == nil {
		writer = NewStdoutWriteSyncer()
	}

	// Set log level
	level := zerolog.InfoLevel
	switch config.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	var zl zerolog.Logger
	if config.Format == "console" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:     writer,
			NoColor: !config.Colored,
		}
		ctx := zerolog.New(consoleWriter).
			Level(level).
			With().
			Timestamp()

		if o.options != nil {
			if o.options.appName != "" {
				ctx = ctx.Str("appName", o.options.appName)
			}
			if o.options.appVersion != "" {
				ctx = ctx.Str("appVersion", o.options.appVersion)
			}
		}

		zl = ctx.Logger()
	} else {
		ctx := zerolog.New(writer).
			Level(level).
			With().
			Timestamp()

		if o.options != nil {
			if o.options.appName != "" {
				ctx = ctx.Str("appName", o.options.appName)
			}
			if o.options.appVersion != "" {
				ctx = ctx.Str("appVersion", o.options.appVersion)
			}
		}

		zl = ctx.Logger()
	}

	return &zerologLogger{logger: zl}
}

// zerologLogger wraps zerolog.Logger to implement our Logger interface
type zerologLogger struct {
	logger zerolog.Logger
}

func (l *zerologLogger) Debug(msg string, keysAndValues ...any) {
	event := l.logger.Debug()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			event = event.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *zerologLogger) Info(msg string, keysAndValues ...any) {
	event := l.logger.Info()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			event = event.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *zerologLogger) Warn(msg string, keysAndValues ...any) {
	event := l.logger.Warn()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			event = event.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *zerologLogger) Error(msg string, keysAndValues ...any) {
	event := l.logger.Error()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			event = event.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *zerologLogger) Fatal(msg string, keysAndValues ...any) {
	event := l.logger.Fatal()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			event = event.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	event.Msg(msg)
}

func (l *zerologLogger) With(keysAndValues ...any) Logger {
	ctx := l.logger.With()
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			ctx = ctx.Interface(keysAndValues[i].(string), keysAndValues[i+1])
		}
	}
	return &zerologLogger{logger: ctx.Logger()}
}

func (l *zerologLogger) WithContext(ctx context.Context) Logger {
	return &zerologLogger{logger: l.logger.With().Ctx(ctx).Logger()}
}
