package log_test

import (
	"os"
	"testing"

	"github.com/go-reusables/log"
)

func Test_Logger(t *testing.T) {
	t.Run("Test JSON slog logger", func(t *testing.T) {
		config := log.Config{
			Level:   "debug",
			Format:  "json",
			Colored: false,
		}

		slogLogger := log.NewLogger(log.SlogType, config, os.Stdout, log.WithAppName("test"), log.WithAppVersion("v0.0.1"))
		slogLogger.Debug("Debug message")
		slogLogger.Warn("Warn message")
		slogLogger.Error("Error message")
		slogLogger.Info("Info message")
	})

	t.Run("Test JSON zerolog logger", func(t *testing.T) {
		config := log.Config{
			Level:  "debug",
			Format: "json",
		}

		zerologLogger := log.NewLogger(log.ZeroLogType, config, os.Stdout)
		zerologLogger.Debug("Debug message")
		zerologLogger.Warn("Warn message")
		zerologLogger.Error("Error message")
		zerologLogger.Info("Info message")
	})

	t.Run("Test console slog logger", func(t *testing.T) {
		config := log.Config{
			Level:   "debug",
			Format:  "console",
			Colored: true,
		}

		slogLogger := log.NewLogger(log.SlogType, config, os.Stdout)
		slogLogger.Debug("Debug message")
		slogLogger.Warn("Warn message")
		slogLogger.Error("Error message")
		slogLogger.Info("Info message")
	})

	t.Run("Test console zerolog logger", func(t *testing.T) {
		config := log.Config{
			Level:   "debug",
			Format:  "console",
			Colored: true,
		}

		zerologLogger := log.NewLogger(log.ZeroLogType, config, os.Stdout)
		zerologLogger.Debug("Debug message")
		zerologLogger.Warn("Warn message")
		zerologLogger.Error("Error message")
		zerologLogger.Info("Info message")
	})
}
