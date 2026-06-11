package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func assertAgentRecord(t *testing.T, got, want *AgentRecord) {
	t.Helper()
	if got.ID != want.ID || got.OrgID != want.OrgID || got.Status != want.Status {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func runAgentVerifierCase(t *testing.T, tc agentVerifierCase) {
	t.Helper()
	const (
		bearer  = "ibex_pat_test"
		agentID = "550e8400-e29b-41d4-a716-446655440000"
		orgID   = "550e8400-e29b-41d4-a716-446655440001"
	)
	got, err := NewGRPCAgentVerifier(tc.client, time.Second).Verify(context.Background(), bearer, agentID, orgID)
	if tc.wantErr != nil {
		assertValidatorError(t, err, tc.wantErr)
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertAgentRecord(t, got, tc.want)
}

func TestGRPCAgentVerifier_Verify_ok(t *testing.T) {
	t.Parallel()
	for _, tc := range agentVerifierOKCases(t) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runAgentVerifierCase(t, tc)
		})
	}
}

func TestGRPCAgentVerifier_Verify_errors(t *testing.T) {
	t.Parallel()
	for _, tc := range agentVerifierErrorCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runAgentVerifierCase(t, tc)
		})
	}
}

func TestGRPCAgentVerifier_contextDeadline(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, err := NewGRPCAgentVerifier(&mockAuthServiceClient{
		validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
			return nil, errors.New("transport error")
		},
	}, time.Second).Verify(ctx, "ibex_pat_test", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001")
	if !errors.Is(err, ErrAgentVerifyUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestGRPCAgentVerifier_unknownGRPCCode(t *testing.T) {
	t.Parallel()
	_, err := NewGRPCAgentVerifier(&mockAuthServiceClient{
		validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
			return nil, status.Error(codes.Unknown, "unexpected")
		},
	}, time.Second).Verify(context.Background(), "ibex_pat_test", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001")
	if !errors.Is(err, ErrAgentVerifyUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestNewGRPCAgentVerifier_defaultTimeout(t *testing.T) {
	t.Parallel()
	v := NewGRPCAgentVerifier(&mockAuthServiceClient{}, 0)
	if v.timeout != 50*time.Millisecond {
		t.Fatalf("timeout: %s", v.timeout)
	}
}
