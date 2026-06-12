package llm

import (
	"errors"
	"strings"
	"testing"
)

func parseChatRequest(t *testing.T, body string) *ChatCompletionRequest {
	t.Helper()
	req, err := ParseChatCompletionRequest(strings.NewReader(body))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return req
}

func TestParseChatCompletionRequest_valid(t *testing.T) {
	body := `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}],"stream":false,"temperature":0.7}`
	req := parseChatRequest(t, body)
	if req.Model != "gpt-4" {
		t.Fatalf("model: %q", req.Model)
	}
	if len(req.Messages) != 1 {
		t.Fatalf("messages: %d", len(req.Messages))
	}
	if req.Messages[0].Role != "user" {
		t.Fatalf("role: %s", req.Messages[0].Role)
	}
	if req.Messages[0].Content != "hi" {
		t.Fatalf("content: %s", req.Messages[0].Content)
	}
	if req.Stream {
		t.Fatal("expected stream false")
	}
	if req.Temperature == nil {
		t.Fatal("temperature nil")
	}
	if *req.Temperature != 0.7 {
		t.Fatalf("temperature: %v", *req.Temperature)
	}
}

func TestParseChatCompletionRequest_unknownFieldsIgnored(t *testing.T) {
	req := parseChatRequest(t, `{"model":"m","messages":[],"extra_field":true}`)
	if req.Model != "m" {
		t.Fatalf("model: %q", req.Model)
	}
}

func assertParseInvalidJSON(t *testing.T, body string) {
	t.Helper()
	_, err := ParseChatCompletionRequest(strings.NewReader(body))
	if !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("err: %v", err)
	}
}

func TestParseChatCompletionRequest_invalidJSON(t *testing.T) {
	assertParseInvalidJSON(t, `{invalid`)
}

func TestParseChatCompletionRequest_rootArray(t *testing.T) {
	assertParseInvalidJSON(t, `[]`)
}

func TestParseChatCompletionRequest_messagesNotArray(t *testing.T) {
	assertParseInvalidJSON(t, `{"model":"m","messages":"x"}`)
}

func TestParseChatCompletionRequest_messageNotObject(t *testing.T) {
	assertParseInvalidJSON(t, `{"model":"m","messages":[1]}`)
}

func TestParseChatCompletionRequest_missingMessagesOK(t *testing.T) {
	req, err := ParseChatCompletionRequest(strings.NewReader(`{"model":"m"}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(req.Messages) != 0 {
		t.Fatalf("messages: %v", req.Messages)
	}
}

func TestParseChatCompletionRequest_trailingDataRejected(t *testing.T) {
	assertParseInvalidJSON(t, `{"model":"m","messages":[]}{}`)
}

func TestParseChatCompletionRequest_streamTrue(t *testing.T) {
	body := `{"model":"m","messages":[],"stream":true}`
	req, err := ParseChatCompletionRequest(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if !req.Stream {
		t.Fatal("expected stream true")
	}
}

func TestParseChatCompletionRequest_messageInvalidJSON(t *testing.T) {
	assertParseInvalidJSON(t, `{"model":"m","messages":[{"role":]}`)
}

func TestParseChatCompletionRequest_maxTokensPreserved(t *testing.T) {
	body := `{"model":"m","messages":[],"max_tokens":128}`
	req, err := ParseChatCompletionRequest(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if req.MaxTokens == nil || *req.MaxTokens != 128 {
		t.Fatalf("max_tokens: %v", req.MaxTokens)
	}
}

func TestParseChatCompletionRequest_emptyModelOK(t *testing.T) {
	req, err := ParseChatCompletionRequest(strings.NewReader(`{"messages":[]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if req.Model != "" {
		t.Fatalf("model: %q", req.Model)
	}
}
