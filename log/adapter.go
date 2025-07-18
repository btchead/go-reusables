package log

import "io"

// LoggerAdapter defines the interface for logger adapters
type LoggerAdapter interface {
	New(config Config, writer io.Writer) Logger
}
