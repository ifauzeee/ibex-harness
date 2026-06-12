package healthcheck

import (
	"context"
	"strings"
	"testing"
)

func TestPostgresSelect1_NilDB(t *testing.T) {
	t.Parallel()
	err := PostgresSelect1(nil)(context.Background())
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("error: %v", err)
	}
}
