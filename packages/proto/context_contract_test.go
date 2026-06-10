package proto_test

import (
	"context"
	"net"
	"testing"

	contextv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/context/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
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

func contextMessageTestCases() []struct {
	name string
	msg  proto.Message
} {
	return []struct {
		name string
		msg  proto.Message
	}{
		{
			name: "AssembleContextRequest",
			msg: &contextv1.AssembleContextRequest{
				AgentId:            "00000000-0000-0000-0000-000000000002",
				OrgId:              "00000000-0000-0000-0000-000000000001",
				SessionId:          "session-1",
				Query:              "hello",
				Model:              "gpt-4",
				DirectiveVersionId: "00000000-0000-0000-0000-000000000005",
				AvailableTokens:    4096,
				RecentMessages: []*contextv1.Message{
					{Role: "user", Content: "hi"},
					{Role: "assistant", Content: "hello"},
				},
				Options: &contextv1.AssemblyOptions{
					SkipColdMemories: true,
					SkipHotMemories:  false,
					RecencyWeight:    0.1,
					RelevanceWeight:  0.2,
					UsefulnessWeight: 0.3,
					ConfidenceWeight: 0.4,
					MaxMemories:      10,
				},
			},
		},
		{
			name: "AssemblyOptions",
			msg: &contextv1.AssemblyOptions{
				SkipColdMemories: true,
				SkipHotMemories:  false,
				RecencyWeight:    0.1,
				RelevanceWeight:  0.2,
				UsefulnessWeight: 0.3,
				ConfidenceWeight: 0.4,
				MaxMemories:      10,
			},
		},
		{
			name: "AssembleContextResponse",
			msg: &contextv1.AssembleContextResponse{
				AssembledContext: "context block",
				TokensUsed:       100,
				MemoriesIncluded: 2,
				MemoriesUsed: []*contextv1.MemoryUsed{
					{
						MemoryId:        "mem-1",
						CompositeScore:  0.9,
						RelevanceScore:  0.8,
						RecencyScore:    0.7,
						UsefulnessScore: 0.6,
						Rank:            1,
						Category:        "fact",
					},
				},
				DirectiveTokens: 10,
				HistoryTokens:   20,
				MemoryTokens:    70,
				Metrics: &contextv1.AssemblyMetrics{
					BudgetCalculationMs:   1,
					DirectiveLoadMs:       2,
					HotMemoryRetrievalMs:  3,
					ColdMemoryRetrievalMs: 4,
					RankingMs:             5,
					PackingMs:             6,
					FormattingMs:          7,
					TotalMs:               28,
					CandidatesEvaluated:   42,
				},
			},
		},
		{
			name: "AssemblyMetrics",
			msg: &contextv1.AssemblyMetrics{
				BudgetCalculationMs:   1,
				DirectiveLoadMs:       2,
				HotMemoryRetrievalMs:  3,
				ColdMemoryRetrievalMs: 4,
				RankingMs:             5,
				PackingMs:             6,
				FormattingMs:          7,
				TotalMs:               28,
				CandidatesEvaluated:   42,
			},
		},
		{
			name: "MemoryUsed",
			msg: &contextv1.MemoryUsed{
				MemoryId:        "mem-1",
				CompositeScore:  0.9,
				RelevanceScore:  0.8,
				RecencyScore:    0.7,
				UsefulnessScore: 0.6,
				Rank:            1,
				Category:        "fact",
			},
		},
		{
			name: "Message",
			msg:  &contextv1.Message{Role: "user", Content: "hello"},
		},
		{
			name: "SearchMemoriesRequest",
			msg: &contextv1.SearchMemoriesRequest{
				AgentId:       "00000000-0000-0000-0000-000000000002",
				OrgId:         "00000000-0000-0000-0000-000000000001",
				Query:         "search query",
				Limit:         5,
				MinSimilarity: 0.75,
				Categories:    []string{"fact", "preference"},
				Tags:          []string{"tag-a"},
				SessionId:     "session-1",
			},
		},
		{
			name: "SearchMemoriesResponse",
			msg: &contextv1.SearchMemoriesResponse{
				Memories: []*contextv1.Memory{
					{
						Id:             "mem-1",
						Content:        "memory text",
						Category:       "fact",
						Confidence:     0.95,
						CompositeScore: 0.88,
						RetrievalCount: 3,
						CreatedAt:      "2026-01-01T00:00:00Z",
					},
				},
				TotalCandidates: 10,
				SearchTimeMs:    12,
			},
		},
		{
			name: "Memory",
			msg: &contextv1.Memory{
				Id:             "mem-1",
				Content:        "memory text",
				Category:       "fact",
				Confidence:     0.95,
				CompositeScore: 0.88,
				RetrievalCount: 3,
				CreatedAt:      "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "RecordMemoryFeedbackRequest",
			msg: &contextv1.RecordMemoryFeedbackRequest{
				MemoryIds: []string{"mem-1", "mem-2"},
				SessionId: "session-1",
				TraceId:   "trace-1",
				OrgId:     "00000000-0000-0000-0000-000000000001",
				Feedback:  "positive",
			},
		},
		{
			name: "RecordMemoryFeedbackResponse",
			msg: &contextv1.RecordMemoryFeedbackResponse{
				Success: true,
				Updates: []*contextv1.MemoryScoreUpdate{
					{MemoryId: "mem-1", PreviousScore: 0.5, NewScore: 0.6},
				},
			},
		},
		{
			name: "MemoryScoreUpdate",
			msg:  &contextv1.MemoryScoreUpdate{MemoryId: "mem-1", PreviousScore: 0.5, NewScore: 0.6},
		},
	}
}

func TestContextMessagesProtoRoundTrip(t *testing.T) {
	t.Parallel()
	tests := contextMessageTestCases()

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			protoRoundTrip(t, tc.msg)
		})
	}
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
