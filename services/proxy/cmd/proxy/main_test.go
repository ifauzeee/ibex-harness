package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/healthcheck"
	"github.com/Rick1330/ibex-harness/packages/logger"
	ibexmetrics "github.com/Rick1330/ibex-harness/packages/metrics"
	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/packages/ratelimit"
	"github.com/Rick1330/ibex-harness/packages/telemetry"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/config"
	proxyhttp "github.com/Rick1330/ibex-harness/services/proxy/internal/http"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func TestRun_InvalidConfigReturns1(t *testing.T) {
	t.Setenv("IBEX_ENV", "not-valid")

	if got := run(nil); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func TestRateLimitSliderConfig(t *testing.T) {
	t.Parallel()
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherOrg := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	cfg := config.Config{
		RateLimit: config.RateLimitConfig{
			DefaultRPM: 120,
			OrgOverrides: map[uuid.UUID]int{
				orgID:    30,
				otherOrg: 45,
			},
		},
	}

	got := rateLimitSliderConfig(cfg)
	if got.DefaultRPM != 120 {
		t.Fatalf("DefaultRPM = %d, want 120", got.DefaultRPM)
	}
	if got.OrgOverrides[orgID] != 30 {
		t.Fatalf("org override = %d, want 30", got.OrgOverrides[orgID])
	}
	if got.OrgOverrides[otherOrg] != 45 {
		t.Fatalf("other org override = %d, want 45", got.OrgOverrides[otherOrg])
	}
	if len(got.OrgOverrides) != 2 {
		t.Fatalf("overrides: %+v", got.OrgOverrides)
	}
}

func TestSetupRateLimiter_NoRedis(t *testing.T) {
	log := logger.Discard("proxy")
	client, limiter, err := setupRateLimiter(config.Config{}, log)
	if err != nil {
		t.Fatalf("setupRateLimiter: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil redis client")
	}
	if limiter == nil {
		t.Fatal("expected noop limiter")
	}
	result, err := limiter.Check(context.Background(), uuid.Nil, uuid.Nil)
	if err != nil || !result.Allowed {
		t.Fatalf("noop limiter check: result=%+v err=%v", result, err)
	}
}

func TestSetupRateLimiter_WithMiniredis(t *testing.T) {
	mr := miniredis.RunT(t)
	log := logger.Discard("proxy")
	cfg := config.Config{RedisURL: "redis://" + mr.Addr() + "/0"}

	client, limiter, err := setupRateLimiter(cfg, log)
	if err != nil {
		t.Fatalf("setupRateLimiter: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if client == nil {
		t.Fatal("expected redis client")
	}
	if limiter == nil {
		t.Fatal("expected limiter")
	}
}

func TestSetupRateLimiter_InvalidURL(t *testing.T) {
	log := logger.Discard("proxy")
	_, _, err := setupRateLimiter(config.Config{RedisURL: "not-a-redis-url"}, log)
	if err == nil {
		t.Fatal("expected error for invalid redis URL")
	}
}

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Port: "8080"}
	cfg.ApplyDefaults()
	srv := newHTTPServer(proxyhttp.RouterDeps{
		Config:  cfg,
		Logger:  logger.Discard("proxy"),
		Metrics: ibexmetrics.NewProxy("proxy"),
		Limiter: ratelimit.Noop(),
		Health:  &healthcheck.Server{},
	})
	if srv.Addr != ":8080" {
		t.Fatalf("addr: %s", srv.Addr)
	}
	if srv.Handler == nil || srv.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("server: %+v", srv)
	}
}

func TestSetupAuthClients_WithGRPCServer(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcSrv := grpc.NewServer() // nosemgrep: go.grpc.security.grpc-server-insecure-connection
	authv1.RegisterAuthServiceServer(grpcSrv, authv1.UnimplementedAuthServiceServer{})
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(func() { grpcSrv.Stop() })

	log := logger.Discard("proxy")
	validator, agentVerifier, client, conn, err := setupAuthClients(config.Config{
		AuthGRPCAddr:        lis.Addr().String(),
		AuthValidateTimeout: time.Second,
	}, log)
	if err != nil {
		t.Fatalf("setupAuthClients: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if validator == nil || agentVerifier == nil || client == nil || conn == nil {
		t.Fatal("expected auth clients")
	}
}

func TestRun_InvalidLoggerLevelReturns1(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("IBEX_LOG_LEVEL", "not-a-level")

	if got := run(nil); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func TestRun_InvalidOTELSampleRatioReturns1(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("OTEL_SAMPLE_RATIO", "2")

	if got := run(nil); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func shutdownTestProviders(t *testing.T) *telemetry.Providers {
	t.Helper()
	providers, err := telemetry.Init(context.Background(), telemetry.Config{ServiceName: "proxy"})
	if err != nil {
		t.Fatal(err)
	}
	return providers
}

func runShutdownTest(t *testing.T, opts shutdownOpts, wantCode int, trigger func()) int {
	t.Helper()
	done := make(chan int, 1)
	go func() { done <- runWithShutdown(opts) }()
	if trigger != nil {
		trigger()
	}
	select {
	case code := <-done:
		if code != wantCode {
			t.Fatalf("runWithShutdown() = %d, want %d", code, wantCode)
		}
		return code
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown")
		return -1
	}
}

func TestRunWithShutdown_serverFailureReturns1(t *testing.T) {
	badPort, _ := strconv.Atoi("99999")
	server := &http.Server{
		Addr:              net.JoinHostPort("127.0.0.1", strconv.Itoa(badPort)),
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	runShutdownTest(t, shutdownOpts{
		cfg: config.Config{
			Environment:     "development",
			ShutdownTimeout: 2 * time.Second,
		},
		logger:    logger.Discard("proxy"),
		providers: shutdownTestProviders(t),
		server:    server,
	}, 1, nil)
}

func TestRun_InvalidRedisURLReturns1(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("REDIS_URL", "not-a-redis-url")

	if got := run(nil); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func TestSetupAuthClients_EmptyAddr(t *testing.T) {
	log := logger.Discard("proxy")
	validator, agentVerifier, client, conn, err := setupAuthClients(config.Config{AuthGRPCAddr: ""}, log)
	if err != nil {
		t.Fatalf("setupAuthClients: %v", err)
	}
	if validator != nil || agentVerifier != nil || client != nil || conn != nil {
		t.Fatal("expected nil auth clients when addr is empty")
	}
}

type shutdownSignalCase struct {
	name string
	opts func(t *testing.T) (shutdownOpts, func())
}

func shutdownSignalCases(t *testing.T) []shutdownSignalCase {
	t.Helper()
	baseServer := func() *http.Server {
		return &http.Server{
			Addr:              "127.0.0.1:0",
			Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
			ReadHeaderTimeout: 5 * time.Second,
		}
	}
	baseCfg := config.Config{Environment: "development", ShutdownTimeout: 2 * time.Second}
	return []shutdownSignalCase{
		{
			name: "stops on signal",
			opts: func(t *testing.T) (shutdownOpts, func()) {
				sigCh := make(chan os.Signal, 1)
				return shutdownOpts{
					cfg: baseCfg, logger: logger.Discard("proxy"),
					providers: shutdownTestProviders(t), server: baseServer(), signalCh: sigCh,
				}, func() { sigCh <- syscall.SIGTERM }
			},
		},
		{
			name: "closes optional clients",
			opts: func(t *testing.T) (shutdownOpts, func()) {
				mr := miniredis.RunT(t)
				redisClient, err := ratelimit.ParseRedisURL("redis://" + mr.Addr() + "/0")
				if err != nil {
					t.Fatal(err)
				}
				sigCh := make(chan os.Signal, 1)
				return shutdownOpts{
					cfg: baseCfg, logger: logger.Discard("proxy"),
					providers: shutdownTestProviders(t), server: baseServer(),
					redisClient: redisClient, signalCh: sigCh,
				}, func() { sigCh <- syscall.SIGTERM }
			},
		},
	}
}

func TestRunWithShutdown_onSignal(t *testing.T) {
	for _, tc := range shutdownSignalCases(t) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			opts, trigger := tc.opts(t)
			runShutdownTest(t, opts, 0, trigger)
		})
	}
}

func proxyBootstrapSmokeEnv(t *testing.T) (sigCh chan os.Signal, httpPort string) {
	t.Helper()
	mr := miniredis.RunT(t)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcSrv := grpc.NewServer() // nosemgrep: go.grpc.security.grpc-server-insecure-connection
	authv1.RegisterAuthServiceServer(grpcSrv, authv1.UnimplementedAuthServiceServer{})
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(func() { grpcSrv.Stop(); _ = lis.Close() })

	httpLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, portStr, err := net.SplitHostPort(httpLis.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_ = httpLis.Close()

	t.Setenv("IBEX_ENV", "development")
	t.Setenv("REDIS_URL", "redis://"+mr.Addr()+"/0")
	t.Setenv("IBEX_AUTH_GRPC_ADDR", lis.Addr().String())
	t.Setenv("IBEX_PORT", portStr)

	return make(chan os.Signal, 1), portStr
}

func TestRun_StopsOnSignal(t *testing.T) {
	sigCh, portStr := proxyBootstrapSmokeEnv(t)
	done := make(chan int, 1)
	go func() { done <- runBootstrap(nil, sigCh) }()

	waitForTCP(t, net.JoinHostPort("127.0.0.1", portStr))
	sigCh <- syscall.SIGTERM

	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runBootstrap() = %d, want 0", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown")
	}
}

func waitForTCP(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
	}
	t.Fatalf("timeout waiting for %s", addr)
}
