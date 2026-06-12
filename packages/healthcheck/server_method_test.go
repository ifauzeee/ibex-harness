package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadyHandler_rejectsNonGET(t *testing.T) {
	t.Parallel()
	srv := &Server{CriticalCheckers: map[string]Checker{
		"a": func(ctx context.Context) error { return nil },
	}}
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, httptest.NewRequest(http.MethodPost, "/ready", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow header: %q", rec.Header().Get("Allow"))
	}
}

func TestHealthHandler_rejectsNonGET(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	(&Server{}).HealthHandler()(rec, httptest.NewRequest(http.MethodPut, "/health", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
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
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if srv.overallTimeout() != 2 {
		t.Fatalf("overall timeout: %s", srv.overallTimeout())
	}
	if srv.perCheckTimeout() != 1 {
		t.Fatalf("per-check timeout: %s", srv.perCheckTimeout())
	}
}
