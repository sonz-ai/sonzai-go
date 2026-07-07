package sonzai

import (
	"context"
	"net/http"
	"testing"
)

// Pin the Raw/RawStream passthrough contract (status/body/headers relayed
// verbatim, no error on non-2xx) plus the WithTenantHost/WithOperatorID
// context overrides every request-building path in http.go applies — the
// mechanism app-runtime (services/app-runtime) relies on to proxy
// platform-api through this SDK instead of hand-rolled REST.

func TestRaw_NonSuccessStatus_NotAnError(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"title":"not found","status":404}`))
	})
	client := newTestClient(t, h)

	raw, err := client.Raw(context.Background(), http.MethodGet, "/api/v1/whatever", nil, nil)
	if err != nil {
		t.Fatalf("Raw returned an error for a 404: %v", err)
	}
	if raw.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", raw.StatusCode)
	}
	if string(raw.Body) != `{"title":"not found","status":404}` {
		t.Errorf("Body = %q", raw.Body)
	}
	if raw.Header.Get("X-Test") != "yes" {
		t.Errorf("expected response headers to be preserved")
	}
}

func TestRaw_TransportFailure_ReturnsError(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	// Point at an unroutable address to force a transport-level failure.
	client.http.baseURL = "http://127.0.0.1:1"

	if _, err := client.Raw(context.Background(), http.MethodGet, "/x", nil, nil); err == nil {
		t.Fatal("expected a transport error, got nil")
	}
}

func TestRaw_AppliesTenantHostAndOperatorID(t *testing.T) {
	var gotTenantHost, gotOperatorID, gotHost, gotAuth string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenantHost = r.Header.Get("X-Sonzai-Tenant-Host")
		gotOperatorID = r.Header.Get("X-Sonzai-Operator-ID")
		gotHost = r.Host
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(t, h)

	ctx := WithTenantHost(context.Background(), "ayala.sonz.ai")
	ctx = WithOperatorID(ctx, "user-42")

	if _, err := client.Raw(ctx, http.MethodGet, "/api/v1/conversations", nil, nil); err != nil {
		t.Fatalf("Raw: %v", err)
	}
	if gotTenantHost != "ayala.sonz.ai" {
		t.Errorf("X-Sonzai-Tenant-Host = %q", gotTenantHost)
	}
	if gotOperatorID != "user-42" {
		t.Errorf("X-Sonzai-Operator-ID = %q", gotOperatorID)
	}
	if gotHost != "ayala.sonz.ai" {
		t.Errorf("Host = %q", gotHost)
	}
	if gotAuth != "Bearer test-api-key" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestRaw_WithoutContextOverrides_NoExtraHeaders(t *testing.T) {
	var gotTenantHost, gotOperatorID string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenantHost = r.Header.Get("X-Sonzai-Tenant-Host")
		gotOperatorID = r.Header.Get("X-Sonzai-Operator-ID")
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(t, h)

	if _, err := client.Raw(context.Background(), http.MethodGet, "/api/v1/agents", nil, nil); err != nil {
		t.Fatalf("Raw: %v", err)
	}
	if gotTenantHost != "" || gotOperatorID != "" {
		t.Errorf("expected no tenant-host/operator-id headers, got %q / %q", gotTenantHost, gotOperatorID)
	}
}

func TestRawStream_RelaysLiveResponse(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: hello\n\n"))
	})
	client := newTestClient(t, h)

	resp, err := client.RawStream(context.Background(), http.MethodGet, "/api/v1/conversations/stream", nil, map[string]string{"Accept": "text/event-stream"})
	if err != nil {
		t.Fatalf("RawStream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d", resp.StatusCode)
	}
}
