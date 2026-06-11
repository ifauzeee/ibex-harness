package grpctest

import (
	"net"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"google.golang.org/grpc"
)

// StartUnimplementedAuthServer listens on 127.0.0.1:0 with a stub AuthService.
// Local loopback-only test fixture; production uses TLS/mTLS.
func StartUnimplementedAuthServer(t *testing.T) net.Listener {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	// nosemgrep: go.grpc.security.grpc-server-insecure-connection.grpc-server-insecure-connection
	srv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(srv, authv1.UnimplementedAuthServiceServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop(); _ = lis.Close() })
	return lis
}
