package sonzai

import (
	"context"
	"net/http"
	"testing"
)

func TestGetHostManifest_200_ReturnsManifestAndETag(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/host/manifest" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("ETag", `"3"`)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"schema_version":1,"tenant_id":"t-1"}`))
	})
	client := newTestClient(t, h)

	result, err := client.Tenants.GetHostManifest(context.Background(), HostManifestOptions{})
	if err != nil {
		t.Fatalf("GetHostManifest: %v", err)
	}
	if result.ETag != `"3"` {
		t.Errorf("ETag = %q", result.ETag)
	}
	if result.NotModified {
		t.Error("expected NotModified=false on 200")
	}
	if string(result.Manifest) != `{"schema_version":1,"tenant_id":"t-1"}` {
		t.Errorf("Manifest = %s", result.Manifest)
	}
}

func TestGetHostManifest_304_NotModified(t *testing.T) {
	var gotIfNoneMatch string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIfNoneMatch = r.Header.Get("If-None-Match")
		w.Header().Set("ETag", `"3"`)
		w.WriteHeader(http.StatusNotModified)
	})
	client := newTestClient(t, h)

	result, err := client.Tenants.GetHostManifest(context.Background(), HostManifestOptions{IfNoneMatch: `"3"`})
	if err != nil {
		t.Fatalf("GetHostManifest: %v", err)
	}
	if gotIfNoneMatch != `"3"` {
		t.Errorf("If-None-Match sent = %q", gotIfNoneMatch)
	}
	if !result.NotModified {
		t.Error("expected NotModified=true on 304")
	}
	if result.Manifest != nil {
		t.Errorf("expected nil Manifest on 304, got %s", result.Manifest)
	}
}

func TestGetHostManifest_404_ReturnsNotFoundError(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"unknown host"}`, http.StatusNotFound)
	})
	client := newTestClient(t, h)

	_, err := client.Tenants.GetHostManifest(context.Background(), HostManifestOptions{})
	if err == nil {
		t.Fatal("expected an error")
	}
	var nf *NotFoundError
	if !isNotFoundError(err, &nf) {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func isNotFoundError(err error, target **NotFoundError) bool {
	nf, ok := err.(*NotFoundError)
	if ok {
		*target = nf
	}
	return ok
}
