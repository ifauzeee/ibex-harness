package auth

import (
	"context"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func agentVerifierErrorCases() []agentVerifierCase {
	return []agentVerifierCase{
		{
			name: "not found",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return nil, status.Error(codes.NotFound, "missing")
				},
			},
			wantErr: ErrAgentNotAuthorized,
		},
		{
			name: "permission denied suspended",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return nil, status.Error(codes.PermissionDenied, "agent is not active")
				},
			},
			wantErr: ErrAgentSuspended,
		},
		{
			name: "permission denied cross tenant",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return nil, status.Error(codes.PermissionDenied, "forbidden")
				},
			},
			wantErr: ErrAgentNotAuthorized,
		},
		{
			name: "unavailable",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return nil, status.Error(codes.Unavailable, "down")
				},
			},
			wantErr: ErrAgentVerifyUnavailable,
		},
		{
			name: "bad agent uuid in response",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return &authv1.ValidateAgentResponse{AgentId: "not-a-uuid", OrgId: "550e8400-e29b-41d4-a716-446655440001", Status: "active"}, nil
				},
			},
			wantErr: ErrAgentVerifyUnavailable,
		},
		{
			name: "bad org uuid in response",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return &authv1.ValidateAgentResponse{AgentId: "550e8400-e29b-41d4-a716-446655440000", OrgId: "bad", Status: "active"}, nil
				},
			},
			wantErr: ErrAgentVerifyUnavailable,
		},
		{
			name: "unauthenticated",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return nil, status.Error(codes.Unauthenticated, "bad token")
				},
			},
			wantErr: ErrAgentNotAuthorized,
		},
	}
}
