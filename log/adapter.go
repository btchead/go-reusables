package log

// LoggerAdapter defines the interface for logger adapters
type LoggerAdapter interface {
	New(config Config, writer WriteSyncer) Logger
}
