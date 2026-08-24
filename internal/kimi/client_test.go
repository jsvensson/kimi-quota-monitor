package kimi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClientUsages verifies that Usages parses a well-formed API response
// into the expected quota values. The API is protobuf-based and serializes
// int64 fields as strings.
func TestClientUsages(t *testing.T) {
	t.Parallel()

	body := `{
		"subType": "TYPE_PURCHASE",
		"usage": {"limit": "5000", "remaining": "3770", "resetTime": "2026-08-31T00:00:00Z"},
		"limits": [{
			"window": {"duration": "300", "timeUnit": "TIME_UNIT_MINUTE"},
			"detail": {"limit": "200", "remaining": "155", "resetTime": "2026-08-24T20:00:00Z"}
		}]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer test-key")
		}
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	u, err := NewClient(srv.URL, "test-key").Usages(context.Background())
	if err != nil {
		t.Fatalf("Usages() error = %v", err)
	}

	if u.Usage.Limit != 5000 || *u.Usage.Remaining != 3770 {
		t.Errorf("weekly quota = %+v, want limit 5000, remaining 3770", u.Usage)
	}
	if len(u.Limits) != 1 {
		t.Fatalf("len(Limits) = %d, want 1", len(u.Limits))
	}
	if u.Limits[0].Window.Duration != 300 {
		t.Errorf("window duration = %v, want 300", u.Limits[0].Window.Duration)
	}
	if u.Limits[0].Detail.Limit != 200 || *u.Limits[0].Detail.Remaining != 155 {
		t.Errorf("5h quota = %+v, want limit 200, remaining 155", u.Limits[0].Detail)
	}
}

// TestNumberUnmarshal verifies that Number accepts both JSON numbers
// and JSON strings, and rejects invalid values.
func TestNumberUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Number
		wantErr bool
	}{
		{"number", `42`, 42, false},
		{"string", `"42"`, 42, false},
		{"empty string", `""`, 0, false},
		{"null", `null`, 0, false},
		{"invalid string", `"abc"`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n Number
			err := json.Unmarshal([]byte(tt.input), &n)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && n != tt.want {
				t.Errorf("Unmarshal(%s) = %d, want %d", tt.input, n, tt.want)
			}
		})
	}
}

// TestWindowMinutes verifies that Minutes normalizes all time units
// to minutes and rejects unknown units and non-positive durations.
func TestWindowMinutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		window Window
		want   int64
		wantOk bool
	}{
		{"minutes", Window{Duration: 300, TimeUnit: "TIME_UNIT_MINUTE"}, 300, true},
		{"hours", Window{Duration: 5, TimeUnit: "TIME_UNIT_HOUR"}, 300, true},
		{"days", Window{Duration: 7, TimeUnit: "TIME_UNIT_DAY"}, 10080, true},
		{"unknown unit", Window{Duration: 5, TimeUnit: "TIME_UNIT_UNKNOWN"}, 0, false},
		{"zero duration", Window{Duration: 0, TimeUnit: "TIME_UNIT_HOUR"}, 0, false},
		{"negative duration", Window{Duration: -1, TimeUnit: "TIME_UNIT_DAY"}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.window.Minutes()
			if ok != tt.wantOk {
				t.Fatalf("Minutes() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("Minutes() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestClientUsagesHTTPError verifies that Usages returns an error
// for non-200 responses.
func TestClientUsagesHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, "bad-key").Usages(context.Background()); err == nil {
		t.Error("Usages() error = nil, want error for 401 response")
	}
}

// TestClientUsagesInvalidJSON verifies that Usages returns an error
// for a malformed response body.
func TestClientUsagesInvalidJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, "key").Usages(context.Background()); err == nil {
		t.Error("Usages() error = nil, want error for invalid JSON")
	}
}
