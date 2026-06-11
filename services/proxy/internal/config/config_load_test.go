package config

import (
	"log/slog"
	"testing"
)

func withProxyEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		check   func(t *testing.T, cfg Config)
	}{
		{
			name: "happy path",
			env: map[string]string{
				"IBEX_ENV": "development", "IBEX_LOG_LEVEL": "WARN",
				"IBEX_RATE_LIMIT_DEFAULT_RPM": "500",
				"IBEX_RATE_LIMIT_ORG_OVERRIDES": "550e8400-e29b-41d4-a716-446655440000=1000",
				"REDIS_URL": "redis://127.0.0.1:6379/0",
			},
			check: func(t *testing.T, cfg Config) {
				if cfg.LogLevel != slog.LevelWarn {
					t.Fatalf("log level: %v", cfg.LogLevel)
				}
				if cfg.RateLimit.DefaultRPM != 500 || len(cfg.RateLimit.OrgOverrides) != 1 {
					t.Fatalf("rate limit: %+v", cfg.RateLimit)
				}
			},
		},
		{
			name:    "invalid log level",
			env:     map[string]string{"IBEX_LOG_LEVEL": "VERBOSE"},
			wantErr: true,
		},
		{
			name: "invalid org rpm overrides",
			env: map[string]string{
				"IBEX_ENV": "development", "IBEX_RATE_LIMIT_ORG_OVERRIDES": "not-a-uuid=60",
			},
			wantErr: true,
		},
		{
			name:    "zero shutdown timeout",
			env:     map[string]string{"IBEX_SHUTDOWN_TIMEOUT": "0s"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			withProxyEnv(t, tc.env)
			cfg, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if tc.check != nil {
				tc.check(t, cfg)
			}
		})
	}
}
