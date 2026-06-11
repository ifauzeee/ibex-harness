package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierror "github.com/Rick1330/ibex-harness/packages/apierror"
	"github.com/Rick1330/ibex-harness/packages/permissions"
	"github.com/Rick1330/ibex-harness/packages/ratelimit"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/auth"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/config"
)

const testChatOrgID = "550e8400-e29b-41d4-a716-446655440001"
const testChatAgentID = "550e8400-e29b-41d4-a716-446655440000"

type chatMockValidator struct {
	res *auth.ValidateResult
	err error
}

func (m *chatMockValidator) Validate(_ context.Context, _ string) (*auth.ValidateResult, error) {
	return m.res, m.err
}

func chatTestConfig() config.Config {
	return config.Config{
		Environment: "test", ServiceName: "proxy", Port: "8080",
		MaxRequestBodyBytes: 1 << 20, RequestIDHeader: "X-Request-ID", TraceIDHeader: "X-Trace-ID",
	}
}

func chatTestHandler(validator auth.TokenValidator, cfg config.Config) http.Handler {
	return newTestRouter(cfg, validator, ratelimit.Noop())
}

type chatRequestOpts struct {
	method      string
	body        string
	contentType string
	auth        bool
	agentID     string
}

func postChat(t *testing.T, handler http.Handler, opts chatRequestOpts) *httptest.ResponseRecorder {
	t.Helper()

	if opts.method == "" {
		opts.method = http.MethodPost
	}
	var bodyReader *strings.Reader
	if opts.body != "" {
		bodyReader = strings.NewReader(opts.body)
	}
	body := io.Reader(http.NoBody)
	if bodyReader != nil {
		body = bodyReader
	}
	req := httptest.NewRequest(opts.method, "/v1/chat/completions", body)
	if opts.auth {
		req.Header.Set("Authorization", "Bearer ibex_pat_test")
	}
	if opts.contentType != "" {
		req.Header.Set("Content-Type", opts.contentType)
	} else if opts.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if opts.agentID != "" {
		req.Header.Set("X-IBEX-Agent-ID", opts.agentID)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func defaultChatValidator() *chatMockValidator {
	return &chatMockValidator{res: &auth.ValidateResult{
		OrgID: testChatOrgID, Permissions: permissions.ProxyChatCompletion,
	}}
}

func TestChatCompletions_validJSON_returns501(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body:    `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
		auth:    true,
		agentID: testChatAgentID,
	})

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), string(apierror.CodeProviderNotConfigured)) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestChatCompletions_invalidJSON_returns400(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body: `{bad`, auth: true, agentID: testChatAgentID,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), string(apierror.CodeInvalidJSON)) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestChatCompletions_noAuth_returns401(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{body: `{}`})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestChatCompletions_missingAgentID_returns400(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body: `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
		auth: true,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), string(apierror.CodeMissingAgentID)) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestChatCompletions_emptyMessages_returns400(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body: `{"model":"m","messages":[]}`, auth: true, agentID: testChatAgentID,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestChatCompletions_unsupportedMediaType_returns415(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body:        `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
		auth:        true,
		contentType: "text/plain",
		agentID:     testChatAgentID,
	})

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), string(apierror.CodeUnsupportedMediaType)) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestChatCompletions_methodNotAllowed_returns405(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(defaultChatValidator(), chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		method: http.MethodGet, auth: true, agentID: testChatAgentID,
	})

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestChatCompletions_bodyTooLarge_returns413(t *testing.T) {
	t.Parallel()

	cfg := chatTestConfig()
	cfg.MaxRequestBodyBytes = 8
	handler := chatTestHandler(defaultChatValidator(), cfg)
	rec := postChat(t, handler, chatRequestOpts{
		body:    `{"model":"gpt-4","messages":[{"role":"user","content":"this body is definitely too large"}]}`,
		auth:    true,
		agentID: testChatAgentID,
	})

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), string(apierror.CodePayloadTooLarge)) {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestChatCompletions_validatorError_returns503(t *testing.T) {
	t.Parallel()

	handler := chatTestHandler(&chatMockValidator{err: auth.ErrAuthUnavailable}, chatTestConfig())
	rec := postChat(t, handler, chatRequestOpts{
		body: `{"model":"m","messages":[]}`, auth: true, agentID: testChatAgentID,
	})

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: %d", rec.Code)
	}
}
