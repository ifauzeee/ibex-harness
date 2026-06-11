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

func runAgentVerifierCase(t *testing.T, tc agentVerifierCase) {
	t.Helper()
	const (
		bearer  = "ibex_pat_test"
		agentID = "550e8400-e29b-41d4-a716-446655440000"
		orgID   = "550e8400-e29b-41d4-a716-446655440001"
	)
	v := NewGRPCAgentVerifier(tc.client, time.Second)
	got, err := v.Verify(context.Background(), bearer, agentID, orgID)
	if tc.wantErr != nil {
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("err = %v, want %v", err, tc.wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != tc.want.ID || got.OrgID != tc.want.OrgID || got.Status != tc.want.Status {
		t.Fatalf("got %+v, want %+v", got, tc.want)
	}
}

func TestGRPCAgentVerifier_Verify(t *testing.T) {
	t.Parallel()
	for _, tc := range agentVerifierCases(t) {
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
	v := NewGRPCAgentVerifier(&mockAuthServiceClient{
		validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
			return nil, errors.New("transport error")
		},
	}, time.Second)
	_, err := v.Verify(ctx, "ibex_pat_test", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001")
	if !errors.Is(err, ErrAgentVerifyUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestGRPCAgentVerifier_unknownGRPCCode(t *testing.T) {
	t.Parallel()
	v := NewGRPCAgentVerifier(&mockAuthServiceClient{
		validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
			return nil, status.Error(codes.Unknown, "unexpected")
		},
	}, time.Second)
	_, err := v.Verify(context.Background(), "ibex_pat_test", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001")
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
