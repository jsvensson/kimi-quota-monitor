package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestQuotaNoPayload verifies that the endpoint returns 503 before
// any payload is set.
func TestQuotaNoPayload(t *testing.T) {
	srv := NewServer(":0")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/quota", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

// TestQuotaServesPayload verifies that the endpoint returns the payload
// set with SetPayload, with a JSON content type.
func TestQuotaServesPayload(t *testing.T) {
	srv := NewServer(":0")
	want := `{"5h":{"pct":42,"resets_at":1756340000},"7d":{"pct":7,"resets_at":1756340000}}`
	srv.SetPayload([]byte(want))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/quota", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

// TestQuotaPayloadUpdated verifies that the endpoint serves the most
// recent payload after SetPayload is called again.
func TestQuotaPayloadUpdated(t *testing.T) {
	srv := NewServer(":0")
	srv.SetPayload([]byte(`{"old":true}`))
	srv.SetPayload([]byte(`{"new":true}`))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/quota", nil)
	srv.Handler().ServeHTTP(rec, req)

	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != `{"new":true}` {
		t.Errorf("body = %q, want %q", body, `{"new":true}`)
	}
}

// TestQuotaMethodNotAllowed verifies that non-GET methods are rejected.
func TestQuotaMethodNotAllowed(t *testing.T) {
	srv := NewServer(":0")
	srv.SetPayload([]byte(`{}`))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/quota", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
