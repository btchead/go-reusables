package service

import (
	"context"
	"os"
	"time"
)

// Option configures the service manager
type Option func(*Manager)

// WithShutdownTimeout sets the timeout for graceful shutdown
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(m *Manager) {
		m.shutdownTimeout = timeout
	}
}

// WithGracefulSignals sets the signals that trigger graceful shutdown
func WithGracefulSignals(signals ...os.Signal) Option {
	return func(m *Manager) {
		m.gracefulSignals = signals
	}
}

// WithForceSignals sets the signals that trigger immediate shutdown
func WithForceSignals(signals ...os.Signal) Option {
	return func(m *Manager) {
		m.forceSignals = signals
	}
}

// WithLogger sets the logger for the service manager
func WithLogger(logger Logger) Option {
	return func(m *Manager) {
		m.logger = logger
	}
}

// WithContext sets the application context for the service manager
func WithContext(ctx context.Context) Option {
	return func(m *Manager) {
		if m.cancel != nil {
			m.cancel() // Cancel the default context
		}
		m.ctx, m.cancel = context.WithCancel(ctx)
	}
}

// WithServiceSequence sets the service startup/shutdown sequence
func WithServiceSequence(sequence ServiceSequence) Option {
	return func(m *Manager) {
		m.serviceSequence = sequence
	}
}