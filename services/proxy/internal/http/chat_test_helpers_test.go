package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
