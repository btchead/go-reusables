package service

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (o NoOpLogger) Debug(msg string, keysAndValues ...any) {}
func (o NoOpLogger) Info(msg string, keysAndValues ...any)  {}
func (o NoOpLogger) Warn(msg string, keysAndValues ...any)  {}
func (o NoOpLogger) Error(msg string, keysAndValues ...any) {}
func (o NoOpLogger) Fatal(msg string, keysAndValues ...any) {}
