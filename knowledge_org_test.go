package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Pin URL shapes, HTTP verbs, and body payloads so the Go SDK stays in
// sync with the Phase 5 handlers in services/platform/api. Response
// parsing is smoke-tested by decoding one field.

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient("test-api-key", WithBaseURL(server.URL))
}

func TestCreateOrgNode_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &seen.body)
		_ = json.NewEncoder(w).Encode(KBNode{NodeID: "n1", Label: "Refund"})
	})
	client := newTestClient(t, h)

	node, err := client.Knowledge.CreateOrgNode(context.Background(), "tenant-abc", CreateOrgNodeOptions{
		NodeType: "policy",
		Label:    "Refund",
		Properties: map[string]any{
			"description": "default refund",
		},
	})
	if err != nil {
		t.Fatalf("CreateOrgNode: %v", err)
	}
	if node.NodeID != "n1" || node.Label != "Refund" {
		t.Errorf("unexpected response: %+v", node)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/tenants/tenant-abc/knowledge/org-nodes" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["node_type"] != "policy" || seen.body["label"] != "Refund" {
		t.Errorf("body missing required fields: %+v", seen.body)
	}
}

func TestPromoteNodeToOrg_URLAndBody(t *testing.T) {
	var seen struct {
		path, method string
		body         map[string]any
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &seen.body)
		resp := KBNodeWithScope{
			KBNode:    &KBNode{NodeID: "org-n1", Label: "Privacy"},
			ScopeType: "organization",
			Relevance: 1.0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	client := newTestClient(t, h)

	got, err := client.Knowledge.PromoteNodeToOrg(context.Background(), "proj-a", "p1", "tenant-abc")
	if err != nil {
		t.Fatalf("PromoteNodeToOrg: %v", err)
	}
	if got.ScopeType != "organization" || got.NodeID != "org-n1" {
		t.Errorf("unexpected response: %+v", got)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	wantPath := "/api/v1/projects/proj-a/knowledge/nodes/p1/promote-to-org"
	if seen.path != wantPath {
		t.Errorf("path: got %q, want %q", seen.path, wantPath)
	}
	if seen.body["tenant_id"] != "tenant-abc" {
		t.Errorf("body.tenant_id: got %v, want tenant-abc", seen.body["tenant_id"])
	}
}

func TestKBScopeMode_Constants(t *testing.T) {
	// Lock the wire values — these round-trip through the platform API and
	// sonzai-python / sonzai-typescript SDKs. A drift breaks cross-SDK
	// compatibility.
	cases := map[KBScopeMode]string{
		KBScopeProjectOnly: "project_only",
		KBScopeOrgOnly:     "org_only",
		KBScopeCascade:     "cascade",
		KBScopeUnion:       "union",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("KBScopeMode %q: got %q", want, string(got))
		}
	}
	// Also ensure the string "union" starts with 'u' — sanity guard that
	// the constants line up rather than being accidentally swapped.
	if !strings.HasPrefix(string(KBScopeUnion), "u") {
		t.Error("KBScopeUnion constant drift")
	}
}
