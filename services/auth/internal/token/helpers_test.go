package token_test

import (
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
)

func TestFutureExpiry(t *testing.T) {
	t.Parallel()

	exp := token.FutureExpiry()
	if exp == nil {
		t.Fatal("expected non-nil expiry")
	}
	if !exp.After(time.Now().UTC()) {
		t.Fatalf("expiry not in future: %v", exp)
	}
}

func TestMustParsePATForTest(t *testing.T) {
	t.Parallel()

	id := "00000000-0000-0000-0000-000000000001"
	bearer := "ibex_pat_" + id + "_secret"
	parsed := token.MustParsePATForTest(bearer)
	if parsed.Prefix != "ibex_pat_"+id {
		t.Fatalf("prefix: %q", parsed.Prefix)
	}
}
