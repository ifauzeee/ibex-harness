package llm

import (
	"context"
	"testing"
)

func TestChatRequestContext_roundTrip(t *testing.T) {
	t.Parallel()

	req := &ChatCompletionRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "hi"}}}
	ctx := WithChatRequest(context.Background(), req)

	got, ok := ChatRequestFromContext(ctx)
	if !ok {
		t.Fatal("expected chat request in context")
	}
	if got.Model != req.Model || len(got.Messages) != 1 || got.Messages[0].Content != "hi" {
		t.Fatalf("got %+v", got)
	}
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
