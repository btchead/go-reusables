package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Service represents a service that can be started and stopped
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ServiceFunc represents the main service logic function
type ServiceFunc func(ctx context.Context) error

// BaseService provides a clean service implementation that handles common patterns
type BaseService struct {
	name      string
	startFunc ServiceFunc
	stopFunc  ServiceFunc
	done      chan struct{}
	running   atomic.Bool
}

// NewService creates a service with just a name and start function
// The start function should block until the service should stop
func NewService(name string, startFunc ServiceFunc) *BaseService {
	return &BaseService{
		name:      name,
		startFunc: startFunc,
		done:      make(chan struct{}),
	}
}

// WithStopFunc adds a custom stop function (optional)
func (o *BaseService) WithStopFunc(stopFunc ServiceFunc) *BaseService {
	o.stopFunc = stopFunc
	return o
}

// Name returns the service name
func (o *BaseService) Name() string {
	return o.name
}

// Start runs the service until stopped or context is cancelled
func (o *BaseService) Start(ctx context.Context) error {
	if !o.running.CompareAndSwap(false, true) {
		return fmt.Errorf("service '%s' is already running", o.name)
	}

	defer o.running.Store(false)

	// Create a context that gets cancelled when Stop is called
	serviceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Monitor for stop signal in background
	go func() {
		select {
		case <-o.done:
			cancel()
		case <-serviceCtx.Done():
		}
	}()

	// Run the user's service logic
	return o.startFunc(serviceCtx)
}

// Stop gracefully stops the service
func (o *BaseService) Stop(ctx context.Context) error {
	if o.stopFunc != nil {
		// Use custom stop function if provided
		if err := o.stopFunc(ctx); err != nil {
			return err
		}
	}

	// Signal the service to stop
	if o.running.Load() {
		select {
		case <-o.done:
			// Already closed
		default:
			close(o.done)
		}
	}
	return nil
}

// IsRunning returns true if the service is currently running
func (o *BaseService) IsRunning() bool {
	return o.running.Load()
}

// serviceState represents the atomic state of a service
type serviceState struct {
	service   Service
	state     atomic.Int32 // ServiceState as int32
	ctx       context.Context
	cancel    context.CancelFunc
	lastError error
	mu        sync.RWMutex   // protects lastError
	wg        sync.WaitGroup // tracks service goroutines
}

// Manager manages the lifecycle of multiple services
type Manager struct {
	services        []*serviceState
	serviceMap      map[string]*serviceState
	shutdownTimeout time.Duration
	gracefulSignals []os.Signal
	forceSignals    []os.Signal
	logger          Logger
	mu              sync.RWMutex
	waitGroup       sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
}

// ServiceState represents the current state of a service
type ServiceState int

const (
	StateStopped ServiceState = iota
	StateStarting
	StateRunning
	StateStopping
	StateError
)

// ServiceInfo contains information about a service's current state
type ServiceInfo struct {
	Name  string
	State ServiceState
	Error error
}

// NewManager creates a new service manager with default configuration
func NewManager(options ...Option) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		services:        make([]*serviceState, 0),
		serviceMap:      make(map[string]*serviceState),
		shutdownTimeout: 30 * time.Second,
		gracefulSignals: []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		forceSignals:    []os.Signal{syscall.SIGKILL},
		logger:          NoOpLogger{},
		ctx:             ctx,
		cancel:          cancel,
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

// setState atomically sets the service state
func (s *serviceState) setState(state ServiceState) {
	s.state.Store(int32(state))
}

// getState atomically gets the service state
func (s *serviceState) getState() ServiceState {
	return ServiceState(s.state.Load())
}

// setError safely sets the last error
func (s *serviceState) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = err
}

// getError safely gets the last error
func (s *serviceState) getError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

// Register adds a service to the manager
func (o *Manager) Register(service Service) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Check for duplicate service names
	if _, exists := o.serviceMap[service.Name()]; exists {
		o.logger.Error("Service registration failed: duplicate name", "service", service.Name())
		return fmt.Errorf("service with name '%s' already registered", service.Name())
	}

	// Create a child context from the manager's application context
	ctx, cancel := context.WithCancel(o.ctx)
	state := &serviceState{
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}
	state.setState(StateStopped)

	o.services = append(o.services, state)
	o.serviceMap[service.Name()] = state
	o.logger.Debug("Service registered", "service", service.Name())
	return nil
}

// Start starts all registered services
func (o *Manager) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Info("Starting all services", "count", len(o.services))

	for _, state := range o.services {
		if state.getState() == StateRunning {
			o.logger.Debug("Service already running, skipping", "service", state.service.Name())
			continue
		}

		o.logger.Debug("Starting service", "service", state.service.Name())
		state.setState(StateStarting)

		// Start service in a goroutine so it can run independently
		state.wg.Add(1)
		o.waitGroup.Add(1)
		go func(s *serviceState) {
			defer s.wg.Done()
			defer o.waitGroup.Done()

			if err := s.service.Start(s.ctx); err != nil {
				o.logger.Error("Service failed during execution", "service", s.service.Name(), "error", err)
				s.setError(err)
				s.setState(StateError)
				return
			}

			// Service.Start should block until the service stops
			// When it returns without error, the service has stopped cleanly
			s.setState(StateStopped)
			o.logger.Info("Service stopped cleanly", "service", s.service.Name())
		}(state)

		// Give the service a moment to start up
		time.Sleep(10 * time.Millisecond)

		// Check if service failed to start
		if state.getState() == StateError {
			o.logger.Error("Service start failed, stopping all services", "service", state.service.Name())
			o.stopAllServices(ctx)
			return fmt.Errorf("failed to start service '%s': %w", state.service.Name(), state.getError())
		}

		state.setState(StateRunning)
		o.logger.Info("Service started successfully", "service", state.service.Name())
	}

	o.logger.Info("All services started successfully")
	return nil
}

// Stop stops all running services in reverse order
func (o *Manager) Stop(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.stopAllServices(ctx)
}

// stopAllServices stops all services (internal helper, assumes lock is held)
func (o *Manager) stopAllServices(ctx context.Context) error {
	var errors []error

	o.logger.Info("Stopping all services", "count", len(o.services))

	// Stop services in reverse order
	for i := len(o.services) - 1; i >= 0; i-- {
		state := o.services[i]
		if state.getState() == StateStopped {
			o.logger.Debug("Service already stopped, skipping", "service", state.service.Name())
			continue
		}

		o.logger.Debug("Stopping service", "service", state.service.Name())
		state.setState(StateStopping)

		// Cancel the service context
		state.cancel()

		if err := state.service.Stop(ctx); err != nil {
			o.logger.Error("Service stop failed", "service", state.service.Name(), "error", err)
			state.setError(err)
			state.setState(StateError)
			errors = append(errors, fmt.Errorf("failed to stop service '%s': %w", state.service.Name(), err))
		} else {
			o.logger.Info("Service stop initiated", "service", state.service.Name())
		}

		// Wait for service goroutines to complete
		o.logger.Debug("Waiting for service goroutines to complete", "service", state.service.Name())
		state.wg.Wait()
		o.logger.Debug("Service goroutines completed", "service", state.service.Name())
	}

	if len(errors) > 0 {
		o.logger.Error("Errors occurred while stopping services", "errors", len(errors))
		return fmt.Errorf("errors stopping services: %v", errors)
	}

	o.logger.Info("All services stopped successfully")
	return nil
}

// StartService starts a specific service by name
func (o *Manager) StartService(ctx context.Context, name string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Debug("Starting specific service", "service", name)

	state, exists := o.serviceMap[name]
	if !exists {
		o.logger.Error("Service not found", "service", name)
		return fmt.Errorf("service '%s' not found", name)
	}

	if state.getState() == StateRunning {
		o.logger.Warn("Attempted to start already running service", "service", name)
		return fmt.Errorf("service '%s' is already running", name)
	}

	state.setState(StateStarting)

	// Start service in a goroutine
	state.wg.Add(1)
	o.waitGroup.Add(1)
	go func() {
		defer state.wg.Done()
		defer o.waitGroup.Done()

		if err := state.service.Start(state.ctx); err != nil {
			o.logger.Error("Service failed during execution", "service", name, "error", err)
			state.setError(err)
			state.setState(StateError)
			return
		}

		state.setState(StateStopped)
		o.logger.Info("Service stopped cleanly", "service", name)
	}()

	// Give the service a moment to start
	time.Sleep(10 * time.Millisecond)

	if state.getState() == StateError {
		return fmt.Errorf("failed to start service '%s': %w", name, state.getError())
	}

	state.setState(StateRunning)
	o.logger.Info("Service started successfully", "service", name)
	return nil
}

// StopService stops a specific service by name
func (o *Manager) StopService(ctx context.Context, name string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Debug("Stopping specific service", "service", name)

	state, exists := o.serviceMap[name]
	if !exists {
		o.logger.Error("Service not found", "service", name)
		return fmt.Errorf("service '%s' not found", name)
	}

	if state.getState() == StateStopped {
		o.logger.Warn("Attempted to stop already stopped service", "service", name)
		return fmt.Errorf("service '%s' is not running", name)
	}

	state.setState(StateStopping)
	state.cancel()

	if err := state.service.Stop(ctx); err != nil {
		o.logger.Error("Failed to stop service", "service", name, "error", err)
		state.setError(err)
		state.setState(StateError)
		return fmt.Errorf("failed to stop service '%s': %w", name, err)
	}

	// Wait for service goroutines to complete
	o.logger.Debug("Waiting for service goroutines to complete", "service", name)
	state.wg.Wait()
	o.logger.Debug("Service goroutines completed", "service", name)
	o.logger.Info("Service stopped successfully", "service", name)
	return nil
}

// IsRunning checks if a service is currently running
// This method is lock-free for better performance
func (o *Manager) IsRunning(name string) bool {
	o.mu.RLock()
	state, exists := o.serviceMap[name]
	o.mu.RUnlock()

	if !exists {
		return false
	}

	return state.getState() == StateRunning
}

// GetStatus returns the status of all registered services
func (o *Manager) GetStatus() []ServiceInfo {
	o.mu.RLock()
	defer o.mu.RUnlock()

	status := make([]ServiceInfo, 0, len(o.services))
	for _, state := range o.services {
		info := ServiceInfo{
			Name: state.service.Name(),
		}

		info.State = state.getState()
		info.Error = state.getError()

		status = append(status, info)
	}

	return status
}

// RunWithGracefulShutdown runs all services and handles graceful shutdown
func (o *Manager) RunWithGracefulShutdown(ctx context.Context) error {
	o.logger.Info("Starting service manager with graceful shutdown")

	// Start all services
	if err := o.Start(ctx); err != nil {
		return err
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, o.gracefulSignals...)
	o.logger.Debug("Graceful shutdown signals registered", "signals", o.gracefulSignals)

	// Setup force shutdown handler
	forceChan := make(chan os.Signal, 1)
	signal.Notify(forceChan, o.forceSignals...)
	o.logger.Debug("Force shutdown signals registered", "signals", o.forceSignals)

	o.logger.Info("Service manager running, waiting for shutdown signal")

	select {
	case <-ctx.Done():
		o.logger.Info("Context cancelled, initiating graceful shutdown")
		return o.Shutdown(ctx)
	case sig := <-sigChan:
		o.logger.Info("Graceful shutdown signal received", "signal", sig)
		return o.gracefulShutdown()
	case sig := <-forceChan:
		o.logger.Warn("Force shutdown signal received", "signal", sig)
		return o.Shutdown(context.Background())
	}
}

// gracefulShutdown performs a graceful shutdown with timeout
func (o *Manager) gracefulShutdown() error {
	o.logger.Info("Starting graceful shutdown", "timeout", o.shutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), o.shutdownTimeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- o.Shutdown(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			o.logger.Error("Graceful shutdown completed with errors", "error", err)
		} else {
			o.logger.Info("Graceful shutdown completed successfully")
		}
		return err
	case <-ctx.Done():
		o.logger.Warn("Graceful shutdown timeout reached, forcing shutdown", "timeout", o.shutdownTimeout)
		return o.Shutdown(context.Background())
	}
}

// Services returns a copy of all registered services
func (o *Manager) Services() []Service {
	o.mu.RLock()
	defer o.mu.RUnlock()

	services := make([]Service, len(o.services))
	for i, state := range o.services {
		services[i] = state.service
	}
	return services
}

// Shutdown gracefully shuts down the manager and all services
func (o *Manager) Shutdown(ctx context.Context) error {
	o.logger.Info("Shutting down service manager")

	// Cancel the manager context
	o.cancel()

	// Stop all services
	err := o.Stop(ctx)

	// Wait for all service goroutines to complete
	o.logger.Debug("Waiting for all service goroutines to complete")
	o.waitGroup.Wait()
	o.logger.Debug("All service goroutines completed")

	o.logger.Info("Service manager shutdown complete")
	return err
}

// HealthCheck returns the health status of all services
func (o *Manager) HealthCheck() map[string]bool {
	o.mu.RLock()
	defer o.mu.RUnlock()

	health := make(map[string]bool)
	for _, state := range o.services {
		health[state.service.Name()] = state.getState() == StateRunning
	}
	return health
}

// WaitForService waits for a specific service to complete
func (o *Manager) WaitForService(name string) error {
	o.mu.RLock()
	state, exists := o.serviceMap[name]
	o.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service '%s' not found", name)
	}

	o.logger.Debug("Waiting for service to complete", "service", name)
	state.wg.Wait()
	o.logger.Debug("Service completed", "service", name)
	return state.getError()
}

// WaitForAllServices waits for all services to complete
func (o *Manager) WaitForAllServices() {
	o.logger.Debug("Waiting for all services to complete")
	o.waitGroup.Wait()
	o.logger.Info("All services completed")
}
