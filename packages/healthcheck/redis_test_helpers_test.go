package healthcheck

import (
	"bufio"
	"net"
	"testing"
)

func runMockRedisServer(t *testing.T, onConn func(conn net.Conn, reader *bufio.Reader)) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		onConn(conn, bufio.NewReader(conn))
	}()
	t.Cleanup(func() { <-done })

	return listener.Addr().String()
}
