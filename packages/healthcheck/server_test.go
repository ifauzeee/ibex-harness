package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func invokeHandler(t *testing.T, handler http.HandlerFunc, method, path string) (int, Response) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return rec.Code, body
}

func TestHealthHandler_AlwaysOK(t *testing.T) {
	t.Parallel()
	srv := &Server{}
	code, body := invokeHandler(t, srv.HealthHandler(), http.MethodGet, "/health")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if body.Status != statusOK || len(body.Checks) != 0 {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestReadyHandler_AllCriticalOK(t *testing.T) {
	t.Parallel()
	srv := &Server{
		CriticalCheckers: map[string]Checker{
			"a": func(ctx context.Context) error { return nil },
			"b": func(ctx context.Context) error { return nil },
		},
	}
	code, body := invokeHandler(t, srv.ReadyHandler(), http.MethodGet, "/ready")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if body.Status != statusOK {
		t.Fatalf("expected ok status, got %s", body.Status)
	}
	if body.Checks["a"].Status != checkOK || body.Checks["b"].Status != checkOK {
		t.Fatalf("unexpected checks: %+v", body.Checks)
	}
}

func TestReadyHandler_CriticalFailure(t *testing.T) {
	t.Parallel()
	srv := &Server{
		CriticalCheckers: map[string]Checker{
			"postgres": func(ctx context.Context) error { return errors.New("connection refused") },
			"redis":    func(ctx context.Context) error { return nil },
		},
	}
	code, body := invokeHandler(t, srv.ReadyHandler(), http.MethodGet, "/ready")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", code)
	}
	if body.Status != statusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", body.Status)
	}
	if body.Checks["postgres"].Status != checkFailed || body.Checks["postgres"].Message == "" {
		t.Fatalf("unexpected postgres check: %+v", body.Checks["postgres"])
	}
}

func TestHandler_rejectsNonGET(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
	}{
		{name: "ready POST", handler: (&Server{CriticalCheckers: map[string]Checker{
			"a": func(ctx context.Context) error { return nil },
		}}).ReadyHandler(), method: http.MethodPost, path: "/ready"},
		{name: "health PUT", handler: (&Server{}).HealthHandler(), method: http.MethodPut, path: "/health"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", rec.Code)
			}
			if tc.path == "/ready" && rec.Header().Get("Allow") != http.MethodGet {
				t.Fatalf("Allow header: %q", rec.Header().Get("Allow"))
			}
		})
	}
}

func TestReadyHandler_customTimeouts(t *testing.T) {
	t.Parallel()

	srv := &Server{
		OverallTimeout:  2,
		PerCheckTimeout: 1,
		CriticalCheckers: map[string]Checker{
			"a": func(ctx context.Context) error { return nil },
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if srv.overallTimeout() != 2 || srv.perCheckTimeout() != 1 {
		t.Fatalf("timeouts: overall=%s per=%s", srv.overallTimeout(), srv.perCheckTimeout())
	}
}

func TestReadyHandler_AdvisoryFailure(t *testing.T) {
	t.Parallel()
	srv := &Server{
		CriticalCheckers: map[string]Checker{
			"postgres": func(ctx context.Context) error { return nil },
		},
		AdvisoryCheckers: map[string]Checker{
			"llm": func(ctx context.Context) error { return errors.New("provider down") },
		},
	}
	code, body := invokeHandler(t, srv.ReadyHandler(), http.MethodGet, "/ready")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if body.Status != statusDegraded {
		t.Fatalf("expected degraded, got %s", body.Status)
	}
}
