package auth

import (
	"context"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type agentVerifierCase struct {
	name    string
	client  *mockAuthServiceClient
	want    *AgentRecord
	wantErr error
}

func agentVerifierCases(t *testing.T) []agentVerifierCase {
	t.Helper()
	const (
		bearer  = "ibex_pat_test"
		agentID = "550e8400-e29b-41d4-a716-446655440000"
		orgID   = "550e8400-e29b-41d4-a716-446655440001"
	)
	return []agentVerifierCase{
		{
			name: "ok",
			client: &mockAuthServiceClient{
				validateAgentFn: func(ctx context.Context, req *authv1.ValidateAgentRequest, _ ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					if req.GetAgentId() != agentID || req.GetOrgId() != orgID {
						t.Fatalf("request: %+v", req)
					}
					md, ok := metadata.FromOutgoingContext(ctx)
					if !ok || md.Get("authorization")[0] != "Bearer "+bearer {
						t.Fatalf("metadata: %v", md)
					}
					return &authv1.ValidateAgentResponse{AgentId: agentID, OrgId: orgID, Status: "active"}, nil
				},
			},
			want: &AgentRecord{ID: uuid.MustParse(agentID), OrgID: uuid.MustParse(orgID), Status: "active"},
		},
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
					return &authv1.ValidateAgentResponse{AgentId: "not-a-uuid", OrgId: orgID, Status: "active"}, nil
				},
			},
			wantErr: ErrAgentVerifyUnavailable,
		},
		{
			name: "bad org uuid in response",
			client: &mockAuthServiceClient{
				validateAgentFn: func(context.Context, *authv1.ValidateAgentRequest, ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
					return &authv1.ValidateAgentResponse{AgentId: agentID, OrgId: "bad", Status: "active"}, nil
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
