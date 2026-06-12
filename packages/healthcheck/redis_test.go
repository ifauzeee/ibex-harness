package healthcheck

import (
	"bufio"
	"context"
	"net"
	"testing"
)

func TestRedisPing_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantMsg string
	}{
		{name: "missing url", url: "", wantMsg: "missing REDIS_URL"},
		{name: "invalid url", url: "http://not-redis", wantMsg: "invalid REDIS_URL"},
		{name: "unreachable", url: "redis://127.0.0.1:1/0", wantMsg: "redis unreachable"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := RedisPing(tc.url)(context.Background())
			if err == nil || err.Error() != tc.wantMsg {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRedisPing_authFailed(t *testing.T) {
	t.Parallel()
	addr := runMockRedisServer(t, func(conn net.Conn, reader *bufio.Reader) {
		_, _ = reader.ReadString('\n')
		_, _ = conn.Write([]byte("-ERR auth failed\r\n"))
	})

	err := RedisPing("redis://:secret@" + addr + "/0")(context.Background())
	if err == nil || err.Error() != "redis auth failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedisPing_pingFailed(t *testing.T) {
	t.Parallel()
	addr := runMockRedisServer(t, func(conn net.Conn, reader *bufio.Reader) {
		_, _ = reader.ReadString('\n')
		_, _ = conn.Write([]byte("-ERR unknown\r\n"))
	})

	err := RedisPing("redis://" + addr + "/0")(context.Background())
	if err == nil || err.Error() != "redis ping failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedisEndpoint_defaultPort(t *testing.T) {
	t.Parallel()
	ep, err := redisEndpoint("redis://redis.internal")
	if err != nil {
		t.Fatal(err)
	}
	if ep.address != "redis.internal:6379" {
		t.Fatalf("address: %q", ep.address)
	}
}

func TestRedisPing_withPasswordAuthOK(t *testing.T) {
	t.Parallel()
	addr := runMockRedisServer(t, func(conn net.Conn, reader *bufio.Reader) {
		for i := 0; i < 5; i++ {
			_, _ = reader.ReadString('\n')
		}
		_, _ = conn.Write([]byte("+OK\r\n"))
		for i := 0; i < 3; i++ {
			_, _ = reader.ReadString('\n')
		}
		_, _ = conn.Write([]byte("+PONG\r\n"))
	})

	err := RedisPing("redis://:secret@" + addr + "/0")(context.Background())
	if err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
}

func TestRedisPing_OK(t *testing.T) {
	t.Parallel()
	addr := runMockRedisServer(t, func(conn net.Conn, reader *bufio.Reader) {
		for i := 0; i < 3; i++ {
			_, _ = reader.ReadString('\n')
		}
		_, _ = conn.Write([]byte("+PONG\r\n"))
	})

	err := RedisPing("redis://" + addr + "/0")(context.Background())
	if err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
}
