package telemetry_test

import (
	"testing"

	"github.com/Rick1330/ibex-harness/packages/telemetry"
)

func TestParentBasedSampler_zeroRatioUsesDefault(t *testing.T) {
	t.Parallel()

	providers, err := telemetry.Init(t.Context(), telemetry.Config{
		ServiceName: "proxy",
		SampleRatio: 0,
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if providers.TracerProvider == nil {
		t.Fatal("expected tracer provider")
	}
	t.Cleanup(func() { _ = providers.Shutdown(t.Context()) }) //nolint:errcheck // test teardown
}

func TestParentBasedSampler_defaultRatioFromEnv(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "proxy")
	t.Setenv("OTEL_SAMPLE_RATIO", "")

	cfg, err := telemetry.ConfigFromEnv("proxy", "development")
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if cfg.SampleRatio != 0.01 {
		t.Fatalf("default sample ratio: got %f want 0.01", cfg.SampleRatio)
	}
}
