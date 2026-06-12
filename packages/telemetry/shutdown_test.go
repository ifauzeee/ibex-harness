package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/telemetry"
)

func TestInitTracer_success(t *testing.T) {
	t.Parallel()

	providers, tracer, err := telemetry.InitTracer(context.Background(), telemetry.Config{
		ServiceName: "proxy",
	}, "ibex-proxy")
	if err != nil {
		t.Fatalf("InitTracer: %v", err)
	}
	if providers == nil || tracer == nil {
		t.Fatal("expected providers and tracer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := providers.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestInit_invalidConfig(t *testing.T) {
	t.Parallel()

	_, err := telemetry.Init(context.Background(), telemetry.Config{})
	if err == nil {
		t.Fatal("expected error for empty ServiceName")
	}
}
