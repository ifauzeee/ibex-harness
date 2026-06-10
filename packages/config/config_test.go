package config_test

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/Rick1330/ibex-harness/packages/config"
)

type sampleConfig struct {
	Required string        `env:"TEST_REQUIRED_VAR,required"`
	Optional string        `env:"TEST_OPTIONAL_VAR" envDefault:"fallback"`
	Secret   config.Secret `env:"TEST_SECRET_VAR" secret:"true"`
}

func TestLoad_missingRequiredAggregates(t *testing.T) {
	_ = os.Unsetenv("TEST_REQUIRED_VAR")
	t.Setenv("TEST_OPTIONAL_VAR", "")
	t.Setenv("TEST_SECRET_VAR", "")

	_, err := config.Load[sampleConfig]()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "TEST_REQUIRED_VAR") {
		t.Fatalf("error should mention missing var: %v", err)
	}
}

func TestLoad_defaultsAndSecret(t *testing.T) {
	t.Setenv("TEST_REQUIRED_VAR", "ok")
	t.Setenv("TEST_OPTIONAL_VAR", "")
	t.Setenv("TEST_SECRET_VAR", "s3cr3t")

	cfg, err := config.Load[sampleConfig]()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Optional != "fallback" {
		t.Fatalf("optional: %q", cfg.Optional)
	}
	if cfg.Secret.String() != "s3cr3t" {
		t.Fatalf("secret: %q", cfg.Secret)
	}
}

type nestedConfig struct {
	Host   string        `env:"NESTED_HOST" envDefault:"localhost"`
	Secret config.Secret `env:"NESTED_SECRET" secret:"true"`
}

func TestLogDebug_redactsSecretsAndNestedStructs(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(old) })

	type debugCfg struct {
		Visible string        `env:"DBG_VISIBLE"`
		Secret  config.Secret `env:"DBG_SECRET" secret:"true"`
		Nested  nestedConfig
		Ptr     *nestedConfig
	}
	cfg := debugCfg{
		Visible: "shown",
		Secret:  config.Secret("top-secret"),
		Nested:  nestedConfig{Host: "db.internal", Secret: config.Secret("nested-secret")},
		Ptr:     &nestedConfig{Host: "cache.internal", Secret: config.Secret("ptr-secret")},
	}

	config.LogDebug(cfg)
	out := buf.String()
	if !strings.Contains(out, "shown") {
		t.Fatalf("expected visible value in log: %s", out)
	}
	for _, secret := range []string{"top-secret", "nested-secret", "ptr-secret"} {
		if strings.Contains(out, secret) {
			t.Fatalf("secret leaked in log: %s", out)
		}
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("expected redaction marker: %s", out)
	}
}

func TestLogDebug_redactsNonStructValue(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(old) })

	config.LogDebug("plain-string")
	if !strings.Contains(buf.String(), "plain-string") {
		t.Fatalf("log: %s", buf.String())
	}
}

func TestLogDebug_usesFieldNameWhenNoEnvTag(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(old) })

	type noTag struct {
		PlainField string
	}
	config.LogDebug(noTag{PlainField: "visible"})
	if !strings.Contains(buf.String(), "PlainField") {
		t.Fatalf("log: %s", buf.String())
	}
}

func TestMustLoad_exitsOnError(t *testing.T) {
	if os.Getenv("TEST_MUST_LOAD_EXIT") == "1" {
		_ = os.Unsetenv("TEST_REQUIRED_VAR")
		_ = config.MustLoad[sampleConfig]()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMustLoad_exitsOnError$", "-test.v")
	cmd.Env = append(os.Environ(), "TEST_MUST_LOAD_EXIT=1", "TEST_REQUIRED_VAR=")
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() == 0 {
		t.Fatalf("MustLoad should exit non-zero: %v", err)
	}
}
