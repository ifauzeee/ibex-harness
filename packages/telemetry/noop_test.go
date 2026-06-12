package telemetry_test

import (
	"context"
	"testing"

	"github.com/Rick1330/ibex-harness/packages/telemetry"
)

func TestNoopTracer_createsSpan(t *testing.T) {
	t.Parallel()

	tracer := telemetry.NoopTracer("test")
	ctx, span := tracer.Start(context.Background(), "noop-span")
	span.End()
	if ctx == nil {
		t.Fatal("expected context")
	}
}
