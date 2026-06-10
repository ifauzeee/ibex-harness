package telemetry_test

import (
	"context"
	"testing"

	"github.com/Rick1330/ibex-harness/packages/telemetry"
)

func TestInit_noOpExporterWhenEndpointEmpty(t *testing.T) {
	t.Parallel()

	providers, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName: "proxy",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if providers.Shutdown == nil {
		t.Fatal("expected shutdown function")
	}
	ctx := context.Background()
	if err := providers.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestInit_withOTLPEndpoint(t *testing.T) {
	t.Parallel()

	providers, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName:    "proxy",
		ServiceVersion: "test",
		Environment:    "test",
		OTLPEndpoint:   "127.0.0.1:4317",
		SampleRatio:    1,
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = providers.Shutdown(context.Background())
}

func TestInit_appliesConfigDefaults(t *testing.T) {
	t.Parallel()

	providers, err := telemetry.Init(context.Background(), telemetry.Config{
		ServiceName: "auth",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if providers.TracerProvider == nil || providers.MeterProvider == nil {
		t.Fatal("expected providers")
	}
}
