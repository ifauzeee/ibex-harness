package proto_test

import (
	"context"
	"net"
	"testing"

	contextv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/context/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestContextProtoContract(t *testing.T) {
	fd := compileProto(t, "ibex/context/v1/context.proto")

	if got := string(fd.Package()); got != "ibex.context.v1" {
		t.Errorf("package: got %q want ibex.context.v1", got)
	}

	svc := findService(fd, "ContextAssemblyService")
	if svc == nil {
		t.Fatal("ContextAssemblyService not found")
	}
	if svc.Methods().Len() != 3 {
		t.Fatalf("ContextAssemblyService RPC count: got %d want 3", svc.Methods().Len())
	}
}

func runContextMessageRoundTrips(t *testing.T, cases []contextMessageCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			protoRoundTrip(t, tc.msg)
		})
	}
}

func TestContextAssembleMessagesProtoRoundTrip(t *testing.T) {
	t.Parallel()
	runContextMessageRoundTrips(t, contextAssembleMessageCases())
}

func TestContextOptionsMessagesProtoRoundTrip(t *testing.T) {
	t.Parallel()
	runContextMessageRoundTrips(t, contextOptionsMessageCases())
}

func TestContextSearchMessagesProtoRoundTrip(t *testing.T) {
	t.Parallel()
	runContextMessageRoundTrips(t, contextSearchMessageCases())
}

type noopContextServer struct {
	contextv1.UnimplementedContextAssemblyServiceServer
}

func (noopContextServer) AssembleContext(context.Context, *contextv1.AssembleContextRequest) (*contextv1.AssembleContextResponse, error) {
	return &contextv1.AssembleContextResponse{AssembledContext: "ok"}, nil
}

func (noopContextServer) SearchMemories(context.Context, *contextv1.SearchMemoriesRequest) (*contextv1.SearchMemoriesResponse, error) {
	return &contextv1.SearchMemoriesResponse{}, nil
}

func (noopContextServer) RecordMemoryFeedback(context.Context, *contextv1.RecordMemoryFeedbackRequest) (*contextv1.RecordMemoryFeedbackResponse, error) {
	return &contextv1.RecordMemoryFeedbackResponse{Success: true}, nil
}

func TestContextAssemblyServiceGRPCRegistration(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer() // nosemgrep: go.grpc.security.grpc-server-insecure-connection
	contextv1.RegisterContextAssemblyServiceServer(srv, noopContextServer{})
	go func() { _ = srv.Serve(lis) }() //nolint:errcheck // bufconn test server; stopped via t.Cleanup
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

	client := contextv1.NewContextAssemblyServiceClient(conn)
	ctx := context.Background()

	if _, err := client.AssembleContext(ctx, &contextv1.AssembleContextRequest{}); err != nil {
		t.Fatalf("AssembleContext: %v", err)
	}
	if _, err := client.SearchMemories(ctx, &contextv1.SearchMemoriesRequest{}); err != nil {
		t.Fatalf("SearchMemories: %v", err)
	}
	if _, err := client.RecordMemoryFeedback(ctx, &contextv1.RecordMemoryFeedbackRequest{}); err != nil {
		t.Fatalf("RecordMemoryFeedback: %v", err)
	}
}
