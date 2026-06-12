package http

import (
	"context"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/services/proxy/internal/auth"
	"github.com/google/uuid"
)

func TestTraceIDContext(t *testing.T) {
	t.Parallel()

	ctx := WithTraceID(context.Background(), "trace-1")
	if got := TraceIDFromContext(ctx); got != "trace-1" {
		t.Fatalf("trace id: %q", got)
	}
	if TraceIDFromContext(context.Background()) != "" {
		t.Fatal("expected empty trace id")
	}
}

func TestRequestStartAndErrorDocsContext(t *testing.T) {
	t.Parallel()

	start := time.Now().UTC()
	ctx := WithRequestStart(context.Background(), start)
	got, ok := RequestStartFromContext(ctx)
	if !ok || !got.Equal(start) {
		t.Fatalf("start: %v ok=%v", got, ok)
	}

	ctx = WithErrorDocsBase(ctx, "https://docs.example/errors")
	if got := ErrorDocsBaseFromContext(ctx); got != "https://docs.example/errors" {
		t.Fatalf("docs base: %q", got)
	}
}

func TestAgentContext(t *testing.T) {
	t.Parallel()

	rec := auth.AgentRecord{
		ID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		OrgID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		Status: "active",
	}
	ctx := WithAgent(context.Background(), rec)
	got, ok := AgentFromContext(ctx)
	if !ok || got.ID != rec.ID {
		t.Fatalf("agent: %+v ok=%v", got, ok)
	}
}
