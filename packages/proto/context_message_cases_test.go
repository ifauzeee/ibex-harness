package proto_test

import (
	contextv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/context/v1"
	"google.golang.org/protobuf/proto"
)

type contextMessageCase struct {
	name string
	msg  proto.Message
}

func contextAssembleMessageCases() []contextMessageCase {
	return []contextMessageCase{
		{
			name: "AssembleContextRequest",
			msg: &contextv1.AssembleContextRequest{
				AgentId: "00000000-0000-0000-0000-000000000002", OrgId: "00000000-0000-0000-0000-000000000001",
				SessionId: "session-1", Query: "hello", Model: "gpt-4",
				DirectiveVersionId: "00000000-0000-0000-0000-000000000005", AvailableTokens: 4096,
				RecentMessages: []*contextv1.Message{
					{Role: "user", Content: "hi"}, {Role: "assistant", Content: "hello"},
				},
				Options: &contextv1.AssemblyOptions{
					SkipColdMemories: true, SkipHotMemories: false,
					RecencyWeight: 0.1, RelevanceWeight: 0.2, UsefulnessWeight: 0.3, ConfidenceWeight: 0.4, MaxMemories: 10,
				},
			},
		},
		{
			name: "AssembleContextResponse",
			msg: &contextv1.AssembleContextResponse{
				AssembledContext: "context block", TokensUsed: 100, MemoriesIncluded: 2,
				MemoriesUsed: []*contextv1.MemoryUsed{{
					MemoryId: "mem-1", CompositeScore: 0.9, RelevanceScore: 0.8, RecencyScore: 0.7,
					UsefulnessScore: 0.6, Rank: 1, Category: "fact",
				}},
				DirectiveTokens: 10, HistoryTokens: 20, MemoryTokens: 70,
				Metrics: &contextv1.AssemblyMetrics{
					BudgetCalculationMs: 1, DirectiveLoadMs: 2, HotMemoryRetrievalMs: 3, ColdMemoryRetrievalMs: 4,
					RankingMs: 5, PackingMs: 6, FormattingMs: 7, TotalMs: 28, CandidatesEvaluated: 42,
				},
			},
		},
	}
}

func contextOptionsMessageCases() []contextMessageCase {
	return []contextMessageCase{
		{
			name: "AssemblyOptions",
			msg: &contextv1.AssemblyOptions{
				SkipColdMemories: true, SkipHotMemories: false,
				RecencyWeight: 0.1, RelevanceWeight: 0.2, UsefulnessWeight: 0.3, ConfidenceWeight: 0.4, MaxMemories: 10,
			},
		},
		{
			name: "AssemblyMetrics",
			msg: &contextv1.AssemblyMetrics{
				BudgetCalculationMs: 1, DirectiveLoadMs: 2, HotMemoryRetrievalMs: 3, ColdMemoryRetrievalMs: 4,
				RankingMs: 5, PackingMs: 6, FormattingMs: 7, TotalMs: 28, CandidatesEvaluated: 42,
			},
		},
		{
			name: "MemoryUsed",
			msg: &contextv1.MemoryUsed{
				MemoryId: "mem-1", CompositeScore: 0.9, RelevanceScore: 0.8, RecencyScore: 0.7,
				UsefulnessScore: 0.6, Rank: 1, Category: "fact",
			},
		},
		{name: "Message", msg: &contextv1.Message{Role: "user", Content: "hello"}},
	}
}

func contextSearchMessageCases() []contextMessageCase {
	return []contextMessageCase{
		{
			name: "SearchMemoriesRequest",
			msg: &contextv1.SearchMemoriesRequest{
				AgentId: "00000000-0000-0000-0000-000000000002", OrgId: "00000000-0000-0000-0000-000000000001",
				Query: "search query", Limit: 5, MinSimilarity: 0.75,
				Categories: []string{"fact", "preference"}, Tags: []string{"tag-a"}, SessionId: "session-1",
			},
		},
		{
			name: "SearchMemoriesResponse",
			msg: &contextv1.SearchMemoriesResponse{
				Memories: []*contextv1.Memory{{
					Id: "mem-1", Content: "memory text", Category: "fact", Confidence: 0.95,
					CompositeScore: 0.88, RetrievalCount: 3, CreatedAt: "2026-01-01T00:00:00Z",
				}},
				TotalCandidates: 10, SearchTimeMs: 12,
			},
		},
		{
			name: "Memory",
			msg: &contextv1.Memory{
				Id: "mem-1", Content: "memory text", Category: "fact", Confidence: 0.95,
				CompositeScore: 0.88, RetrievalCount: 3, CreatedAt: "2026-01-01T00:00:00Z",
			},
		},
		{
			name: "RecordMemoryFeedbackRequest",
			msg: &contextv1.RecordMemoryFeedbackRequest{
				MemoryIds: []string{"mem-1", "mem-2"}, SessionId: "session-1", TraceId: "trace-1",
				OrgId: "00000000-0000-0000-0000-000000000001", Feedback: "positive",
			},
		},
		{
			name: "RecordMemoryFeedbackResponse",
			msg: &contextv1.RecordMemoryFeedbackResponse{
				Success: true,
				Updates: []*contextv1.MemoryScoreUpdate{{MemoryId: "mem-1", PreviousScore: 0.5, NewScore: 0.6}},
			},
		},
		{name: "MemoryScoreUpdate", msg: &contextv1.MemoryScoreUpdate{MemoryId: "mem-1", PreviousScore: 0.5, NewScore: 0.6}},
	}
}
