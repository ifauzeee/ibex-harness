package httptestx

import (
	"io"
	"net/http"
	"strings"
)

// ChatHeaders configures optional headers on a chat completion request.
type ChatHeaders struct {
	Auth        bool
	ContentType string
	AgentID     string
	HasBody     bool
}

// RequestBody returns an io.Reader for the given body string.
func RequestBody(body string) io.Reader {
	if body == "" {
		return http.NoBody
	}
	return strings.NewReader(body)
}

// ApplyChatHeaders sets chat completion headers on req.
func ApplyChatHeaders(req *http.Request, h ChatHeaders) {
	if h.Auth {
		req.Header.Set("Authorization", "Bearer ibex_pat_test")
	}
	switch {
	case h.ContentType != "":
		req.Header.Set("Content-Type", h.ContentType)
	case h.HasBody:
		req.Header.Set("Content-Type", "application/json")
	}
	if h.AgentID != "" {
		req.Header.Set("X-IBEX-Agent-ID", h.AgentID)
	}
}
