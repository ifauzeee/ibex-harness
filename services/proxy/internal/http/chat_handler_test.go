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

func TestChatCompletions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		validator  auth.TokenValidator
		cfg        config.Config
		req        chatRequestOpts
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid JSON returns 501", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req: chatRequestOpts{
				body: `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
				auth: true, agentID: testChatAgentID,
			},
			wantStatus: http.StatusNotImplemented,
			wantBody:   string(apierror.CodeProviderNotConfigured),
		},
		{
			name: "invalid JSON returns 400", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req:        chatRequestOpts{body: `{bad`, auth: true, agentID: testChatAgentID},
			wantStatus: http.StatusBadRequest,
			wantBody:   string(apierror.CodeInvalidJSON),
		},
		{
			name: "no auth returns 401", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req:        chatRequestOpts{body: `{}`},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing agent ID returns 400", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req: chatRequestOpts{
				body: `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
				auth: true,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   string(apierror.CodeMissingAgentID),
		},
		{
			name: "empty messages returns 400", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req:        chatRequestOpts{body: `{"model":"m","messages":[]}`, auth: true, agentID: testChatAgentID},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unsupported media type returns 415", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req: chatRequestOpts{
				body: `{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}`,
				auth: true, contentType: "text/plain", agentID: testChatAgentID,
			},
			wantStatus: http.StatusUnsupportedMediaType,
			wantBody:   string(apierror.CodeUnsupportedMediaType),
		},
		{
			name: "method not allowed returns 405", validator: defaultChatValidator(), cfg: chatTestConfig(),
			req:        chatRequestOpts{method: http.MethodGet, auth: true, agentID: testChatAgentID},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "body too large returns 413", validator: defaultChatValidator(),
			cfg: func() config.Config {
				c := chatTestConfig()
				c.MaxRequestBodyBytes = 8
				return c
			}(),
			req: chatRequestOpts{
				body:    `{"model":"gpt-4","messages":[{"role":"user","content":"this body is definitely too large"}]}`,
				auth:    true,
				agentID: testChatAgentID,
			},
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   string(apierror.CodePayloadTooLarge),
		},
		{
			name: "validator error returns 503", validator: &chatMockValidator{err: auth.ErrAuthUnavailable},
			cfg: chatTestConfig(),
			req: chatRequestOpts{
				body: `{"model":"m","messages":[]}`, auth: true, agentID: testChatAgentID,
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := chatTestHandler(tc.validator, tc.cfg)
			rec := postChat(t, handler, tc.req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
			}
			if tc.wantBody != "" && !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("body: %s", rec.Body.String())
			}
		})
	}
}
