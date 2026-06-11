package auth

import (
	"context"
	"errors"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcValidatorCase struct {
	name    string
	client  *mockAuthServiceClient
	want    *ValidateResult
	wantErr error
}

func grpcValidatorCases(t *testing.T) []grpcValidatorCase {
	t.Helper()
	const (
		orgID       = "550e8400-e29b-41d4-a716-446655440001"
		agentID     = "550e8400-e29b-41d4-a716-446655440000"
		userID      = "550e8400-e29b-41d4-a716-446655440002"
		fixtureRef  = "test-token-id-1"
		accessToken = "ibex_pat_test"
	)
	return []grpcValidatorCase{
		{
			name: "ok minimal fields",
			client: &mockAuthServiceClient{
				validateTokenFn: func(_ context.Context, req *authv1.ValidateTokenRequest, _ ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					if req.GetAccessToken() != accessToken {
						t.Fatalf("token: %q", req.GetAccessToken())
					}
					return &authv1.ValidateTokenResponse{OrgId: orgID, Permissions: 42}, nil
				},
			},
			want: &ValidateResult{OrgID: orgID, Permissions: 42},
		},
		{
			name: "ok optional fields",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return &authv1.ValidateTokenResponse{
						OrgId: orgID, Permissions: 7,
						AgentId: strPtr(agentID), UserId: strPtr(userID), TokenId: strPtr(fixtureRef),
					}, nil
				},
			},
			want: &ValidateResult{OrgID: orgID, Permissions: 7, AgentID: agentID, UserID: userID, TokenID: fixtureRef},
		},
		{
			name: "unauthenticated",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, status.Error(codes.Unauthenticated, "invalid token")
				},
			},
			wantErr: ErrInvalidToken,
		},
		{
			name: "unavailable",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, status.Error(codes.Unavailable, "down")
				},
			},
			wantErr: ErrAuthUnavailable,
		},
		{
			name: "deadline exceeded",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, status.Error(codes.DeadlineExceeded, "deadline")
				},
			},
			wantErr: ErrAuthUnavailable,
		},
		{
			name: "canceled",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, status.Error(codes.Canceled, "canceled")
				},
			},
			wantErr: ErrAuthUnavailable,
		},
		{
			name: "internal maps to unavailable",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, status.Error(codes.Internal, "boom")
				},
			},
			wantErr: ErrAuthUnavailable,
		},
		{
			name: "non status error",
			client: &mockAuthServiceClient{
				validateTokenFn: func(context.Context, *authv1.ValidateTokenRequest, ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
					return nil, errors.New("transport reset")
				},
			},
			wantErr: ErrAuthUnavailable,
		},
	}
}
