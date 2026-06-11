package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLoadFromEnvHappyPath(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("IBEX_LOG_LEVEL", "WARN")
	t.Setenv("IBEX_RATE_LIMIT_DEFAULT_RPM", "500")
	t.Setenv("IBEX_RATE_LIMIT_ORG_OVERRIDES", "550e8400-e29b-41d4-a716-446655440000=1000")
	t.Setenv("REDIS_URL", "redis://127.0.0.1:6379/0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != slog.LevelWarn {
		t.Fatalf("log level: %v", cfg.LogLevel)
	}
	if cfg.RateLimit.DefaultRPM != 500 {
		t.Fatalf("rpm: %d", cfg.RateLimit.DefaultRPM)
	}
	if len(cfg.RateLimit.OrgOverrides) != 1 {
		t.Fatalf("overrides: %v", cfg.RateLimit.OrgOverrides)
	}
}

func TestLoadRejectsInvalidLogLevel(t *testing.T) {
	t.Setenv("IBEX_LOG_LEVEL", "VERBOSE")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid log level error")
	}
}

func TestLoadRejectsInvalidOrgRPMOverrides(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("IBEX_RATE_LIMIT_ORG_OVERRIDES", "not-a-uuid=60")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid org override error")
	}
}

func validProxyConfig() Config {
	cfg := Config{
		Environment:         "development",
		ServiceName:         "proxy",
		Port:                "8080",
		AuthGRPCAddr:        "127.0.0.1:9091",
		AuthValidateTimeout: defaultAuthValidateTimeout,
		MaxRequestBodyBytes: defaultMaxRequestBodyBytes,
		RequestIDHeader:     defaultRequestIDHeader,
		TraceIDHeader:       defaultTraceIDHeader,
	}
	cfg.ApplyDefaults()
	return cfg
}

func TestValidate_rejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	orgID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name string
		mutate func(*Config)
	}{
		{
			name: "invalid environment",
			mutate: func(c *Config) {
				c.Environment = "prod"
				c.ServiceName = "proxy"
				c.Port = "8080"
			},
		},
		{
			name: "invalid port",
			mutate: func(c *Config) { c.Port = "not-a-port" },
		},
		{
			name: "zero port",
			mutate: func(c *Config) { c.Port = "0" },
		},
		{
			name: "port too large",
			mutate: func(c *Config) { c.Port = "70000" },
		},
		{
			name: "empty service name",
			mutate: func(c *Config) { c.ServiceName = "  " },
		},
		{
			name: "invalid auth grpc addr",
			mutate: func(c *Config) { c.AuthGRPCAddr = "not-host-port" },
		},
		{
			name: "zero rate limit rpm",
			mutate: func(c *Config) { c.RateLimit.DefaultRPM = 0 },
		},
		{
			name: "auth grpc required outside development",
			mutate: func(c *Config) {
				c.Environment = "staging"
				c.AuthGRPCAddr = ""
			},
		},
		{
			name: "empty trace id header",
			mutate: func(c *Config) { c.TraceIDHeader = "" },
		},
		{
			name: "zero max body bytes",
			mutate: func(c *Config) { c.MaxRequestBodyBytes = 0 },
		},
		{
			name: "empty request id header",
			mutate: func(c *Config) { c.RequestIDHeader = "" },
		},
		{
			name: "org override zero rpm",
			mutate: func(c *Config) {
				c.RateLimit.OrgOverrides = map[uuid.UUID]int{orgID: 0}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := validProxyConfig()
			tc.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestValidateAcceptsDefaultShape(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Environment:         "development",
		ServiceName:         "proxy",
		Port:                "8080",
		AuthGRPCAddr:        "127.0.0.1:9091",
		AuthValidateTimeout: defaultAuthValidateTimeout,
		MaxRequestBodyBytes: defaultMaxRequestBodyBytes,
		RequestIDHeader:     defaultRequestIDHeader,
		TraceIDHeader:       defaultTraceIDHeader,
	}
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to validate: %v", err)
	}
}

func TestApplyDefaultsZeroConfigValidates(t *testing.T) {
	t.Parallel()

	var cfg Config
	cfg.ApplyDefaults()
	cfg.Environment = "development"
	cfg.ServiceName = "proxy"
	cfg.Port = "8080"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected zero config with defaults to validate: %v", err)
	}
	if cfg.RequestIDHeader != defaultRequestIDHeader {
		t.Fatalf("RequestIDHeader: %s", cfg.RequestIDHeader)
	}
	if cfg.MaxRequestBodyBytes != defaultMaxRequestBodyBytes {
		t.Fatalf("MaxRequestBodyBytes: %d", cfg.MaxRequestBodyBytes)
	}
	if cfg.RateLimit.DefaultRPM != defaultRateLimitRPM {
		t.Fatalf("RateLimit.DefaultRPM: %d", cfg.RateLimit.DefaultRPM)
	}
	if cfg.ShutdownTimeout != defaultShutdownTimeout {
		t.Fatalf("ShutdownTimeout: %s", cfg.ShutdownTimeout)
	}
}

func TestApplyDefaultsShutdownTimeout(t *testing.T) {
	var cfg Config
	cfg.ApplyDefaults()
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("ShutdownTimeout: %s", cfg.ShutdownTimeout)
	}
}

func TestLoadRejectsNonPositiveShutdownTimeout(t *testing.T) {
	t.Setenv("IBEX_SHUTDOWN_TIMEOUT", "0s")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for zero shutdown timeout")
	}
}

func TestParseOrgRPMOverrides(t *testing.T) {
	orgID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	got, err := parseOrgRPMOverrides(orgID.String() + "=1000")
	if err != nil {
		t.Fatal(err)
	}
	if got[orgID] != 1000 {
		t.Fatalf("rpm: %d", got[orgID])
	}
}

func TestParseOrgRPMOverrides_invalid(t *testing.T) {
	if _, err := parseOrgRPMOverrides("not-a-uuid=60"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := parseOrgRPMOverrides("550e8400-e29b-41d4-a716-446655440000=0"); err == nil {
		t.Fatal("expected error for zero rpm")
	}
}

