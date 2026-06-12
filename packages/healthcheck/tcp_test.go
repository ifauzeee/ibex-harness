package healthcheck

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestTCPReachable_missingAddress(t *testing.T) {
	t.Parallel()

	err := TCPReachable("")(context.Background())
	if !errors.Is(err, ErrTCPAddressNotConfigured) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTCPReachable_unreachable(t *testing.T) {
	t.Parallel()

	err := TCPReachable("127.0.0.1:1")(context.Background())
	if err == nil {
		t.Fatal("expected error for closed port")
	}
}

func TestTCPReachable_ok(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()

	err = TCPReachable(ln.Addr().String())(context.Background())
	if err != nil {
		t.Fatalf("expected reachable: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("accept did not complete")
	}
}

func TestTCPReachable_contextTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := TCPReachable("127.0.0.1:9")(ctx)
	if !errors.Is(err, ErrTCPReadinessTimeout) {
		t.Fatalf("unexpected error: %v", err)
	}
}
