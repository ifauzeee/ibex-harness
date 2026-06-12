package service

import (
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type createTokenCase struct {
	name    string
	req     *authv1.CreateTokenRequest
	wantErr error
}

func createTokenCases(orgID, userID, agentID string, expires time.Time) []createTokenCase {
	return []createTokenCase{
		{name: "invalid empty org", req: &authv1.CreateTokenRequest{Name: "x"}, wantErr: ErrInvalidArgument},
		{name: "invalid empty name", req: &authv1.CreateTokenRequest{OrgId: orgID}, wantErr: ErrInvalidArgument},
		{
			name:    "invalid token type",
			req:     &authv1.CreateTokenRequest{OrgId: orgID, Name: "bad-type", Type: authv1.TokenType(99)},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "happy path",
			req: &authv1.CreateTokenRequest{
				OrgId: orgID, Name: "ci-pat", Description: "desc", Permissions: 42,
				UserId: &userID, AgentId: &agentID, ExpiresAt: timestamppb.New(expires),
			},
		},
	}
}
