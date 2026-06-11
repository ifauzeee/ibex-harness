package main

import (
	"net"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/logger"
	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/auth"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/config"
	"google.golang.org/grpc"
)

func startAuthGRPCForTest(t *testing.T) net.Listener {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcSrv := grpc.NewServer() // nosemgrep: go.grpc.security.grpc-server-insecure-connection
	authv1.RegisterAuthServiceServer(grpcSrv, authv1.UnimplementedAuthServiceServer{})
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(func() { grpcSrv.Stop() })
	return lis
}

func assertAuthClientsPresent(t *testing.T, validator auth.TokenValidator, agentVerifier auth.AgentVerifier, client authv1.AuthServiceClient, conn *grpc.ClientConn) {
	t.Helper()
	if validator == nil {
		t.Fatal("validator nil")
	}
	if agentVerifier == nil {
		t.Fatal("agentVerifier nil")
	}
	if client == nil {
		t.Fatal("client nil")
	}
	if conn == nil {
		t.Fatal("conn nil")
	}
}

func assertAuthClientsAbsent(t *testing.T, validator auth.TokenValidator, agentVerifier auth.AgentVerifier, client authv1.AuthServiceClient, conn *grpc.ClientConn) {
	t.Helper()
	if validator != nil {
		t.Fatal("validator should be nil")
	}
	if agentVerifier != nil {
		t.Fatal("agentVerifier should be nil")
	}
	if client != nil {
		t.Fatal("client should be nil")
	}
	if conn != nil {
		t.Fatal("conn should be nil")
	}
}

func TestSetupAuthClients_WithGRPCServer(t *testing.T) {
	lis := startAuthGRPCForTest(t)
	log := logger.Discard("proxy")
	validator, agentVerifier, client, conn, err := setupAuthClients(config.Config{
		AuthGRPCAddr: lis.Addr().String(), AuthValidateTimeout: time.Second,
	}, log)
	if err != nil {
		t.Fatalf("setupAuthClients: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	assertAuthClientsPresent(t, validator, agentVerifier, client, conn)
}

func TestSetupAuthClients_EmptyAddr(t *testing.T) {
	log := logger.Discard("proxy")
	validator, agentVerifier, client, conn, err := setupAuthClients(config.Config{AuthGRPCAddr: ""}, log)
	if err != nil {
		t.Fatalf("setupAuthClients: %v", err)
	}
	assertAuthClientsAbsent(t, validator, agentVerifier, client, conn)
}
