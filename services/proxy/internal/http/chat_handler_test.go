package http

import (
	"net/http"
	"strings"
	"testing"

	apierror "github.com/Rick1330/ibex-harness/packages/apierror"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/auth"
	"github.com/Rick1330/ibex-harness/services/proxy/internal/config"
)

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
			rec := postChat(t, chatTestHandler(tc.validator, tc.cfg), tc.req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
			}
			if tc.wantBody != "" && !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("body: %s", rec.Body.String())
			}
		})
	}
}
