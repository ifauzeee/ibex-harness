package healthcheck

import (
	"bufio"
	"context"
	"net"
	"testing"
)

func TestRedisPing_MissingURL(t *testing.T) {
	t.Parallel()
	err := RedisPing("")(context.Background())
	if err == nil || err.Error() != "missing REDIS_URL" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedisPing_invalidURL(t *testing.T) {
	t.Parallel()
	err := RedisPing("http://not-redis")(context.Background())
	if err == nil || err.Error() != "invalid REDIS_URL" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedisPing_unreachable(t *testing.T) {
	t.Parallel()
	err := RedisPing("redis://127.0.0.1:1/0")(context.Background())
	if err == nil || err.Error() != "redis unreachable" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedisPing_authFailed(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		_, _ = reader.ReadString('\n')
		_, _ = conn.Write([]byte("-ERR auth failed\r\n"))
	}()

	err = RedisPing("redis://:secret@" + listener.Addr().String() + "/0")(context.Background())
	if err == nil || err.Error() != "redis auth failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	<-done
}

func TestRedisPing_pingFailed(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		_, _ = reader.ReadString('\n')
		_, _ = conn.Write([]byte("-ERR unknown\r\n"))
	}()

	err = RedisPing("redis://" + listener.Addr().String() + "/0")(context.Background())
	if err == nil || err.Error() != "redis ping failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	<-done
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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		// AUTH *2 $4 AUTH $6 secret
		for i := 0; i < 5; i++ {
			_, _ = reader.ReadString('\n')
		}
		_, _ = conn.Write([]byte("+OK\r\n"))
		// PING *1 $4 PING
		for i := 0; i < 3; i++ {
			_, _ = reader.ReadString('\n')
		}
		_, _ = conn.Write([]byte("+PONG\r\n"))
	}()

	err = RedisPing("redis://:secret@" + listener.Addr().String() + "/0")(context.Background())
	if err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
	<-done
}

func TestRedisPing_OK(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		_, _ = reader.ReadString('\n')
		_, _ = reader.ReadString('\n')
		_, _ = reader.ReadString('\n')
		_, _ = conn.Write([]byte("+PONG\r\n"))
	}()

	err = RedisPing("redis://" + listener.Addr().String() + "/0")(context.Background())
	if err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
	<-done
}
