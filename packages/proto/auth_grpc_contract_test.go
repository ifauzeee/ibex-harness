package proto_test

import (
	"context"
	"net"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type noopAuthServer struct {
	authv1.UnimplementedAuthServiceServer
}

func (noopAuthServer) ValidateToken(context.Context, *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	return &authv1.ValidateTokenResponse{OrgId: "00000000-0000-0000-0000-000000000001"}, nil
}

func (noopAuthServer) ValidateAgent(context.Context, *authv1.ValidateAgentRequest) (*authv1.ValidateAgentResponse, error) {
	return &authv1.ValidateAgentResponse{Status: "active"}, nil
}

func (noopAuthServer) CreateToken(context.Context, *authv1.CreateTokenRequest) (*authv1.CreateTokenResponse, error) {
	return &authv1.CreateTokenResponse{TokenId: "00000000-0000-0000-0000-000000000004"}, nil
}

func (noopAuthServer) RevokeToken(context.Context, *authv1.RevokeTokenRequest) (*authv1.RevokeTokenResponse, error) {
	return &authv1.RevokeTokenResponse{}, nil
}

func (noopAuthServer) ListTokens(context.Context, *authv1.ListTokensRequest) (*authv1.ListTokensResponse, error) {
	return &authv1.ListTokensResponse{}, nil
}

func TestAuthServiceGRPCRegistration(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(srv, noopAuthServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := authv1.NewAuthServiceClient(conn)
	ctx := context.Background()

	if _, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{AccessToken: "ibex_pat_test"}); err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if _, err := client.ValidateAgent(ctx, &authv1.ValidateAgentRequest{}); err != nil {
		t.Fatalf("ValidateAgent: %v", err)
	}
	if _, err := client.CreateToken(ctx, &authv1.CreateTokenRequest{}); err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if _, err := client.RevokeToken(ctx, &authv1.RevokeTokenRequest{}); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if _, err := client.ListTokens(ctx, &authv1.ListTokensRequest{}); err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
}
