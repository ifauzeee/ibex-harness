package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
)

func strPtr(s string) *string { return &s }

func runGRPCValidatorCase(t *testing.T, tc grpcValidatorCase, accessToken string) {
	t.Helper()
	v := NewGRPCValidator(tc.client, time.Second)
	got, err := v.Validate(context.Background(), accessToken)
	if tc.wantErr != nil {
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("err = %v, want %v", err, tc.wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OrgID != tc.want.OrgID || got.Permissions != tc.want.Permissions {
		t.Fatalf("result: %+v, want %+v", got, tc.want)
	}
	if got.AgentID != tc.want.AgentID || got.UserID != tc.want.UserID || got.TokenID != tc.want.TokenID {
		t.Fatalf("optional fields: %+v, want %+v", got, tc.want)
	}
}

func TestGRPCValidator_Validate(t *testing.T) {
	t.Parallel()
	for _, tc := range grpcValidatorCases(t) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGRPCValidatorCase(t, tc, "ibex_pat_test")
		})
	}
}

func TestNewGRPCValidator_defaultTimeout(t *testing.T) {
	t.Parallel()
	client := &mockAuthServiceClient{
		validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
			return &authv1.ValidateTokenResponse{OrgId: "org", Permissions: 1}, nil
		},
	}
	v := NewGRPCValidator(client, 0)
	if v.timeout != 50*time.Millisecond {
		t.Fatalf("timeout: %s", v.timeout)
	}
}
