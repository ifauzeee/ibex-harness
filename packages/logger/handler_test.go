package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/reqid"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestIBEXHandler_injectsServiceAndTraceFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	h := newIBEXHandler("proxy", slog.LevelDebug, &buf, false)

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()
	ctx = reqid.WithRequestID(ctx, "550e8400-e29b-41d4-a716-446655440000")

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	if err := h.Handle(ctx, record); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("json: %v raw=%q", err, buf.String())
	}
	if out["service"] != "proxy" {
		t.Fatalf("service: %v", out["service"])
	}
	if out["request_id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("request_id: %v", out["request_id"])
	}
	if out["trace_id"] == "" {
		t.Fatalf("missing trace_id: %v", out)
	}
}

func TestIBEXHandler_WithAttrsAndGroup(t *testing.T) {
	t.Parallel()

	h := newIBEXHandler("auth", slog.LevelInfo, ioDiscard{}, false)
	withAttrs := h.WithAttrs([]slog.Attr{slog.String("component", "store")})
	if withAttrs == nil {
		t.Fatal("WithAttrs returned nil")
	}
	withGroup := h.WithGroup("db")
	if withGroup == nil {
		t.Fatal("WithGroup returned nil")
	}
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected info level enabled")
	}
}

func TestTraceIDFrom_emptyWithoutSpan(t *testing.T) {
	t.Parallel()

	if got := traceIDFrom(context.Background()); got != "" {
		t.Fatalf("trace_id: %q", got)
	}
}

func TestTraceIDFrom_validSpan(t *testing.T) {
	t.Parallel()

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	ctx, span := tp.Tracer("test").Start(context.Background(), "op")
	defer span.End()

	got := traceIDFrom(ctx)
	sc := trace.SpanFromContext(ctx).SpanContext()
	if got != sc.TraceID().String() {
		t.Fatalf("trace_id: %q want %q", got, sc.TraceID().String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
