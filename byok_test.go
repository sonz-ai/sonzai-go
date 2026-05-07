package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// Pin URL shapes, HTTP verbs, and body payloads so the Go SDK stays in
// sync with the BYOK handlers in services/platform/api and the sibling
// Python / TypeScript SDKs. Response parsing is smoke-tested by decoding
// one or two fields.

func TestBYOK_List_URLAndDecode(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{
					"provider":       "openai",
					"api_key_prefix": "sk-abc",
					"is_active":      true,
					"health_status":  "healthy",
					"updated_at":     "2026-05-07T00:00:00Z",
				},
			},
		})
	})
	client := newTestClient(t, h)

	keys, err := client.BYOK.List(context.Background(), "proj-a")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-a/byok-keys" {
		t.Errorf("path: got %q", seen.path)
	}
	if len(keys) != 1 || keys[0].Provider != "openai" || keys[0].APIKeyPrefix != "sk-abc" {
		t.Errorf("decoded list mismatch: %+v", keys)
	}
}

func TestBYOK_Set_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":       "gemini",
			"api_key_prefix": "AIza...",
			"is_active":      true,
			"health_status":  "healthy",
			"updated_at":     "2026-05-07T00:00:00Z",
		})
	})
	client := newTestClient(t, h)

	got, err := client.BYOK.Set(context.Background(), "proj-a", BYOKProviderGemini, "real-secret-key")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if seen.method != http.MethodPut {
		t.Errorf("method: got %s, want PUT", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-a/byok-keys/gemini" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["api_key"] != "real-secret-key" {
		t.Errorf("body.api_key: got %v, want real-secret-key", seen.body["api_key"])
	}
	// Snake-case body key must be exactly "api_key" (not "apiKey").
	if _, ok := seen.body["apiKey"]; ok {
		t.Error("body must use snake_case api_key, not apiKey")
	}
	if got == nil || got.Provider != "gemini" {
		t.Errorf("decoded mismatch: %+v", got)
	}
}

func TestBYOK_Delete_URLAndMethod(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	client := newTestClient(t, h)

	if err := client.BYOK.Delete(context.Background(), "proj-a", BYOKProviderXAI); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if seen.method != http.MethodDelete {
		t.Errorf("method: got %s, want DELETE", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-a/byok-keys/xai" {
		t.Errorf("path: got %q", seen.path)
	}
}

func TestBYOK_SetActive_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seen.body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":       "openrouter",
			"api_key_prefix": "sk-or-",
			"is_active":      false,
			"health_status":  "healthy",
			"updated_at":     "2026-05-07T00:00:00Z",
		})
	})
	client := newTestClient(t, h)

	got, err := client.BYOK.SetActive(context.Background(), "proj-a", BYOKProviderOpenRouter, false)
	if err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if seen.method != http.MethodPatch {
		t.Errorf("method: got %s, want PATCH", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-a/byok-keys/openrouter" {
		t.Errorf("path: got %q", seen.path)
	}
	if active, ok := seen.body["is_active"].(bool); !ok || active != false {
		t.Errorf("body.is_active: got %v, want false", seen.body["is_active"])
	}
	// Snake-case body key must be exactly "is_active" (not "isActive").
	if _, ok := seen.body["isActive"]; ok {
		t.Error("body must use snake_case is_active, not isActive")
	}
	if got == nil || got.IsActive != false {
		t.Errorf("decoded mismatch: %+v", got)
	}
}

func TestBYOK_Test_URLAndMethod(t *testing.T) {
	var seen struct {
		path, method string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":       "openai",
			"api_key_prefix": "sk-abc",
			"is_active":      true,
			"health_status":  "healthy",
			"updated_at":     "2026-05-07T00:00:00Z",
		})
	})
	client := newTestClient(t, h)

	got, err := client.BYOK.Test(context.Background(), "proj-a", BYOKProviderOpenAI)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	// Test endpoint MUST be /test slash, NOT :test colon.
	wantPath := "/api/v1/projects/proj-a/byok-keys/openai/test"
	if seen.path != wantPath {
		t.Errorf("path: got %q, want %q", seen.path, wantPath)
	}
	if got == nil || got.HealthStatus != "healthy" {
		t.Errorf("decoded mismatch: %+v", got)
	}
}

func TestBYOKProvider_Constants(t *testing.T) {
	// Lock the wire values — these round-trip through the platform API
	// and the sibling sonzai-python / sonzai-typescript SDKs.
	cases := map[BYOKProvider]string{
		BYOKProviderOpenAI:     "openai",
		BYOKProviderGemini:     "gemini",
		BYOKProviderXAI:        "xai",
		BYOKProviderOpenRouter: "openrouter",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("BYOKProvider %q: got %q", want, string(got))
		}
	}
}
