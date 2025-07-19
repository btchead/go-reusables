package retrier

import (
	"context"
	"net"
	"syscall"
	"testing"
)

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network operation error",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED},
			expected: true,
		},
		{
			name:     "DNS error",
			err:      &net.DNSError{Err: "no such host", Name: "example.invalid"},
			expected: true,
		},
		{
			name:     "address error",
			err:      &net.AddrError{Err: "missing port in address", Addr: "localhost"},
			expected: true,
		},
		{
			name:     "parse error",
			err:      &net.ParseError{Type: "IP address", Text: "invalid"},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      syscall.ECONNREFUSED,
			expected: true,
		},
		{
			name:     "connection reset",
			err:      syscall.ECONNRESET,
			expected: true,
		},
		{
			name:     "connection aborted",
			err:      syscall.ECONNABORTED,
			expected: true,
		},
		{
			name:     "timeout",
			err:      syscall.ETIMEDOUT,
			expected: true,
		},
		{
			name:     "host unreachable",
			err:      syscall.EHOSTUNREACH,
			expected: true,
		},
		{
			name:     "network unreachable",
			err:      syscall.ENETUNREACH,
			expected: true,
		},
		{
			name:     "broken pipe",
			err:      syscall.EPIPE,
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "non-network error",
			err:      syscall.ENOENT, // File not found
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNetworkError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsTemporaryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network error should be temporary",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED},
			expected: true,
		},
		{
			name:     "context timeout should be temporary",
			err:      context.DeadlineExceeded,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTemporaryError(tt.err)
			if result != tt.expected {
				t.Errorf("IsTemporaryError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
