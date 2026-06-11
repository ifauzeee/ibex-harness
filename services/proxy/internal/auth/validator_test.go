package auth

import (
	"context"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
)

func strPtr(s string) *string { return &s }

func runGRPCValidatorCase(t *testing.T, tc grpcValidatorCase, accessToken string) {
	t.Helper()
	got, err := NewGRPCValidator(tc.client, time.Second).Validate(context.Background(), accessToken)
	if tc.wantErr != nil {
		assertValidatorError(t, err, tc.wantErr)
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidatorResult(t, got, tc.want)
}

func TestGRPCValidator_Validate_success(t *testing.T) {
	t.Parallel()
	for _, tc := range grpcValidatorSuccessCases(t) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runGRPCValidatorCase(t, tc, "ibex_pat_test")
		})
	}
}

func TestGRPCValidator_Validate_errors(t *testing.T) {
	t.Parallel()
	for _, tc := range grpcValidatorErrorCases() {
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
