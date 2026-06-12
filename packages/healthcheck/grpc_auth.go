package healthcheck

import (
	"context"
	"errors"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// probeToken is rejected at PAT parse time without hitting Argon2 or Postgres.
const probeToken = "ibex_health_probe_invalid"

const defaultGRPCProbeTimeout = 500 * time.Millisecond

var (
	// ErrAuthGRPCClientNotConfigured is returned when the auth gRPC client was not wired.
	ErrAuthGRPCClientNotConfigured = errors.New("auth grpc client not configured")
	// ErrAuthGRPCReadinessTimeout is returned when ValidateToken does not respond in time.
	ErrAuthGRPCReadinessTimeout = errors.New("auth grpc readiness check timed out")
	// ErrAuthGRPCUnreachable is returned for transport or unknown gRPC failures during probe.
	ErrAuthGRPCUnreachable = errors.New("auth grpc unreachable")
)

// AuthGRPC returns a checker that calls ValidateToken with a sentinel invalid token.
// codes.Unauthenticated means the auth service is reachable.
func AuthGRPC(client authv1.AuthServiceClient, timeout time.Duration) Checker {
	if timeout <= 0 {
		timeout = defaultGRPCProbeTimeout
	}
	return func(ctx context.Context) error {
		if client == nil {
			return ErrAuthGRPCClientNotConfigured
		}
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		_, err := client.ValidateToken(callCtx, &authv1.ValidateTokenRequest{AccessToken: probeToken})
		if err == nil {
			return nil
		}
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			return nil
		}
		if errors.Is(callCtx.Err(), context.DeadlineExceeded) {
			return ErrAuthGRPCReadinessTimeout
		}
		return ErrAuthGRPCUnreachable
	}
}
