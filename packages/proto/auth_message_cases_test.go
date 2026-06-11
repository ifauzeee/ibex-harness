package proto_test

import (
	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authMessageCase struct {
	name string
	msg  proto.Message
}

func authValidateMessageCases(now *timestamppb.Timestamp) []authMessageCase {
	return []authMessageCase{
		{name: "ValidateTokenRequest", msg: &authv1.ValidateTokenRequest{AccessToken: "ibex_pat_test"}},
		{
			name: "ValidateTokenResponse",
			msg: &authv1.ValidateTokenResponse{
				OrgId: "00000000-0000-0000-0000-000000000001", Permissions: 42,
				AgentId: strPtr("00000000-0000-0000-0000-000000000002"),
				UserId:  strPtr("00000000-0000-0000-0000-000000000003"),
				TokenId: strPtr("00000000-0000-0000-0000-000000000004"), ExpiresAt: now,
			},
		},
		{
			name: "ValidateAgentRequest",
			msg:  &authv1.ValidateAgentRequest{AgentId: "00000000-0000-0000-0000-000000000002", OrgId: "00000000-0000-0000-0000-000000000001"},
		},
		{
			name: "ValidateAgentResponse",
			msg:  &authv1.ValidateAgentResponse{AgentId: "00000000-0000-0000-0000-000000000002", OrgId: "00000000-0000-0000-0000-000000000001", Status: "active"},
		},
	}
}

func authTokenMessageCases(now *timestamppb.Timestamp) []authMessageCase {
	return []authMessageCase{
		{
			name: "CreateTokenRequest",
			msg: &authv1.CreateTokenRequest{
				OrgId: "00000000-0000-0000-0000-000000000001", Name: "test-token", Description: "integration test",
				Type: authv1.TokenType_TOKEN_TYPE_PAT, Permissions: 7, ExpiresAt: now,
				UserId: strPtr("00000000-0000-0000-0000-000000000003"), AgentId: strPtr("00000000-0000-0000-0000-000000000002"),
			},
		},
		{
			name: "CreateTokenResponse",
			msg: &authv1.CreateTokenResponse{
				TokenId: "00000000-0000-0000-0000-000000000004", Plaintext: "ibex_pat_secret",
				Prefix: "ibex_pat_00000000", CreatedAt: now,
			},
		},
		{
			name: "RevokeTokenRequest",
			msg: &authv1.RevokeTokenRequest{
				OrgId: "00000000-0000-0000-0000-000000000001", TokenId: "00000000-0000-0000-0000-000000000004",
				RevokeReason: strPtr("test revoke"),
			},
		},
		{name: "RevokeTokenResponse", msg: &authv1.RevokeTokenResponse{}},
	}
}

func authListMessageCases(now *timestamppb.Timestamp) []authMessageCase {
	meta := &authv1.TokenMetadata{
		TokenId: "00000000-0000-0000-0000-000000000004", Name: "test-token", Prefix: "ibex_pat_00000000",
		Permissions: 7, ExpiresAt: now, CreatedAt: now, RevokedAt: now, IsRevoked: true,
	}
	return []authMessageCase{
		{name: "ListTokensRequest", msg: &authv1.ListTokensRequest{OrgId: "00000000-0000-0000-0000-000000000001", Cursor: "cursor-1", Limit: 25}},
		{name: "TokenMetadata", msg: meta},
		{name: "ListTokensResponse", msg: &authv1.ListTokensResponse{Tokens: []*authv1.TokenMetadata{meta}, NextCursor: "next-cursor"}},
	}
}
