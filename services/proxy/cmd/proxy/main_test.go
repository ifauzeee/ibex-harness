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
	cfg := config.Config{
		RateLimit: config.RateLimitConfig{
			DefaultRPM: 120,
			OrgOverrides: map[uuid.UUID]int{
				orgID: 30,
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

func TestRunWithShutdown_serverFailureReturns1(t *testing.T) {
	log := logger.Discard("proxy")
	providers, err := telemetry.Init(context.Background(), telemetry.Config{ServiceName: "proxy"})
	if err != nil {
		t.Fatal(err)
	}

	badPort, _ := strconv.Atoi("99999")
	server := &http.Server{
		Addr:              net.JoinHostPort("127.0.0.1", strconv.Itoa(badPort)),
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	done := make(chan int, 1)
	go func() {
		done <- runWithShutdown(shutdownOpts{
			cfg: config.Config{
				Environment:     "development",
				ShutdownTimeout: 2 * time.Second,
			},
			logger:    log,
			providers: providers,
			server:    server,
		})
	}()

	select {
	case code := <-done:
		if code != 1 {
			t.Fatalf("runWithShutdown() = %d, want 1", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server failure")
	}
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

func TestRunWithShutdown_closesOptionalClients(t *testing.T) {
	log := logger.Discard("proxy")
	providers, err := telemetry.Init(context.Background(), telemetry.Config{ServiceName: "proxy"})
	if err != nil {
		t.Fatal(err)
	}

	mr := miniredis.RunT(t)
	redisClient, err := ratelimit.ParseRedisURL("redis://" + mr.Addr() + "/0")
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	done := make(chan int, 1)
	go func() {
		done <- runWithShutdown(shutdownOpts{
			cfg: config.Config{
				Environment:     "development",
				ShutdownTimeout: 2 * time.Second,
			},
			logger:      log,
			providers:   providers,
			server:      server,
			redisClient: redisClient,
			signalCh:    sigCh,
		})
	}()

	sigCh <- syscall.SIGTERM

	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runWithShutdown() = %d, want 0", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown")
	}
}

func TestRun_StopsOnSignal(t *testing.T) {
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

	sigCh := make(chan os.Signal, 1)
	done := make(chan int, 1)
	go func() { done <- runBootstrap(nil, sigCh) }()

	time.Sleep(200 * time.Millisecond)
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

func TestRun_AuthGRPCDialFailureReturns1(t *testing.T) {
	t.Setenv("IBEX_ENV", "development")
	t.Setenv("IBEX_AUTH_GRPC_ADDR", "127.0.0.1:1")
	t.Setenv("REDIS_URL", "")

	if got := run(nil); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func TestRunWithShutdown_StopsOnSignal(t *testing.T) {
	log := logger.Discard("proxy")
	providers, err := telemetry.Init(context.Background(), telemetry.Config{ServiceName: "proxy"})
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{
		Addr:              ":0",
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	done := make(chan int, 1)
	go func() {
		done <- runWithShutdown(shutdownOpts{
			cfg: config.Config{
				Environment:     "development",
				ShutdownTimeout: 2 * time.Second,
			},
			logger:    log,
			providers: providers,
			server:    server,
			signalCh:  sigCh,
		})
	}()

	sigCh <- syscall.SIGTERM

	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runWithShutdown() = %d, want 0", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown")
	}
}
