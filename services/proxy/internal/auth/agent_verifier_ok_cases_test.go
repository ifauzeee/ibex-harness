package auth

import (
	"context"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func agentVerifierOKCases(t *testing.T) []agentVerifierCase {
	t.Helper()
	const (
		bearer  = "ibex_pat_test"
		agentID = "550e8400-e29b-41d4-a716-446655440000"
		orgID   = "550e8400-e29b-41d4-a716-446655440001"
	)
	return []agentVerifierCase{{
		name: "ok",
		client: &mockAuthServiceClient{
			validateAgentFn: func(ctx context.Context, req *authv1.ValidateAgentRequest, _ ...grpc.CallOption) (*authv1.ValidateAgentResponse, error) {
				if req.GetAgentId() != agentID || req.GetOrgId() != orgID {
					t.Fatalf("request: %+v", req)
				}
				md, _ := metadata.FromOutgoingContext(ctx)
				if md.Get("authorization")[0] != "Bearer "+bearer {
					t.Fatalf("metadata: %v", md)
				}
				return &authv1.ValidateAgentResponse{AgentId: agentID, OrgId: orgID, Status: "active"}, nil
			},
		},
		want: &AgentRecord{ID: uuid.MustParse(agentID), OrgID: uuid.MustParse(orgID), Status: "active"},
	}}
}
