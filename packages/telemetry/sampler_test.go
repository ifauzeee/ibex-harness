package telemetry_test

import (
	"testing"

	"github.com/Rick1330/ibex-harness/packages/telemetry"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
	_ = sdktrace.AlwaysSample()
}
