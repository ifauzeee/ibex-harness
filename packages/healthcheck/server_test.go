package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_AlwaysOK(t *testing.T) {
	t.Parallel()
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.HealthHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
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
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
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
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != statusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", body.Status)
	}
	if body.Checks["postgres"].Status != checkFailed || body.Checks["postgres"].Message == "" {
		t.Fatalf("unexpected postgres check: %+v", body.Checks["postgres"])
	}
}

func TestReadyHandler_rejectsNonGET(t *testing.T) {
	t.Parallel()

	srv := &Server{
		CriticalCheckers: map[string]Checker{
			"a": func(ctx context.Context) error { return nil },
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow header: %q", rec.Header().Get("Allow"))
	}
}

func TestHealthHandler_rejectsNonGET(t *testing.T) {
	t.Parallel()

	srv := &Server{}
	req := httptest.NewRequest(http.MethodPut, "/health", nil)
	rec := httptest.NewRecorder()
	srv.HealthHandler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
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
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	srv.ReadyHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != statusDegraded {
		t.Fatalf("expected degraded, got %s", body.Status)
	}
}
