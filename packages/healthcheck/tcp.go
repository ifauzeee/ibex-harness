package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	ErrTCPAddressNotConfigured = errors.New("tcp address not configured")
	ErrTCPReadinessTimeout     = errors.New("tcp readiness check timed out")
)

// TCPReachable dials hostPort to verify a listener accepts connections.
func TCPReachable(hostPort string) Checker {
	return func(ctx context.Context) error {
		addr := strings.TrimSpace(hostPort)
		if addr == "" {
			return ErrTCPAddressNotConfigured
		}
		dialer := net.Dialer{Timeout: 500 * time.Millisecond}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			if ctx.Err() != nil {
				return ErrTCPReadinessTimeout
			}
			return fmt.Errorf("tcp unreachable: %w", err)
		}
		_ = conn.Close()
		return nil
	}
}
