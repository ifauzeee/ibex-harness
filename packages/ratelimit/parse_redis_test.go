package ratelimit

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestParseRedisURL_empty(t *testing.T) {
	t.Parallel()

	_, err := ParseRedisURL("")
	if err == nil || !strings.Contains(err.Error(), "REDIS_URL is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRedisURL_invalid(t *testing.T) {
	t.Parallel()

	_, err := ParseRedisURL("not-a-url")
	if err == nil || !strings.Contains(err.Error(), "parse REDIS_URL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRedisURL_ok(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	client, err := ParseRedisURL("redis://" + mr.Addr() + "/0")
	if err != nil {
		t.Fatalf("ParseRedisURL: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}
