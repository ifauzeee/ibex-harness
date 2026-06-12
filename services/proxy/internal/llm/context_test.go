package llm

import (
	"context"
	"testing"
)

func assertChatRequestFields(t *testing.T, got, want *ChatCompletionRequest) {
	t.Helper()
	if got.Model != want.Model {
		t.Fatalf("model: %s", got.Model)
	}
	if len(got.Messages) != 1 || got.Messages[0].Content != want.Messages[0].Content {
		t.Fatalf("messages: %+v", got.Messages)
	}
}

func TestChatRequestContext_roundTrip(t *testing.T) {
	t.Parallel()
	req := &ChatCompletionRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "hi"}}}
	got, ok := ChatRequestFromContext(WithChatRequest(context.Background(), req))
	if !ok {
		t.Fatal("expected chat request in context")
	}
	assertChatRequestFields(t, got, req)
}

func TestChatRequestContext_missing(t *testing.T) {
	t.Parallel()
	_, ok := ChatRequestFromContext(context.Background())
	if ok {
		t.Fatal("expected false without chat request")
	}
}

func TestChatRequestContext_nilValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), chatRequestKey{}, (*ChatCompletionRequest)(nil))
	_, ok := ChatRequestFromContext(ctx)
	if ok {
		t.Fatal("expected false for nil chat request")
	}
}
