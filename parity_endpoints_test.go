package sonzai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

// mustParseQuery parses a raw URL query string or fails the test.
func mustParseQuery(t *testing.T, rawQuery string) url.Values {
	t.Helper()
	v, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("parse query %q: %v", rawQuery, err)
	}
	return v
}

// Pin URL shapes, HTTP verbs, and body payloads for the Go-parity endpoints
// added to reach feature parity with the Python / TypeScript SDKs:
//
//   Knowledge.ListMultimodalSchemas / CreateMultimodalSchema /
//     ActivateMultimodalSchema / ReingestDocument / GetDocumentCost
//   Projects.Delete
//   Tenants.ListOrgKnowledgeNodes
//   Org.Me
//
// Each test pins the path + method (and body where there is one) against the
// committed openapi.json and the sibling SDK implementations, then smoke-tests
// response decode on the load-bearing fields. Follows the ml_test.go pattern.

// ---------------------------------------------------------------------------
// Knowledge: multimodal schema CRUD + reingest + cost
// ---------------------------------------------------------------------------

func TestKnowledge_ListMultimodalSchemas_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"schemas": []map[string]any{
				{
					"project_id":     "proj-1",
					"schema_version": 2,
					"status":         "active",
					"config":         map[string]any{"classify_model": "gemini", "classify_auto_threshold": 0.8},
					"doc_types": []map[string]any{
						{"type": "price_list", "root_entity_type": "hospital", "expected_relationships": []string{"hospital_offers_procedure"}},
					},
					"entity_types": []map[string]any{
						{"type": "hospital", "is_root_candidate": true, "key_fields": []string{"name"}},
					},
					"relationship_types": []map[string]any{
						{"type": "hospital_offers_procedure", "from": "hospital", "to": "procedure", "supersession_identity": []string{"procedure"}},
					},
					"created_at": "2026-05-22T10:00:00Z",
				},
			},
		})
	})
	client := newTestClient(t, h)

	res, err := client.Knowledge.ListMultimodalSchemas(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("ListMultimodalSchemas: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1/knowledge/multimodal-schemas" {
		t.Errorf("path: got %q", seen.path)
	}
	if len(res.Schemas) != 1 {
		t.Fatalf("schemas: got %d, want 1", len(res.Schemas))
	}
	s := res.Schemas[0]
	if s.SchemaVersion != 2 || s.Status != "active" || s.Config.ClassifyModel != "gemini" {
		t.Errorf("decoded schema mismatch: %+v", s)
	}
	if len(s.DocTypes) != 1 || s.DocTypes[0].RootEntityType != "hospital" {
		t.Errorf("decoded doc_types mismatch: %+v", s.DocTypes)
	}
	if len(s.EntityTypes) != 1 || !s.EntityTypes[0].IsRootCandidate || s.EntityTypes[0].KeyFields[0] != "name" {
		t.Errorf("decoded entity_types mismatch: %+v", s.EntityTypes)
	}
	if len(s.RelationshipTypes) != 1 || s.RelationshipTypes[0].From != "hospital" {
		t.Errorf("decoded relationship_types mismatch: %+v", s.RelationshipTypes)
	}
}

func TestKnowledge_CreateMultimodalSchema_URLBodyAndDecode(t *testing.T) {
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
			"schema": map[string]any{
				"project_id":     "proj-1",
				"schema_version": 3,
				"status":         "draft",
				"config":         map[string]any{"extract_model": "gpt"},
				"created_at":     "2026-05-22T11:00:00Z",
			},
		})
	})
	client := newTestClient(t, h)

	res, err := client.Knowledge.CreateMultimodalSchema(context.Background(), "proj-1", KBSchema{
		ProjectID:     "proj-1",
		SchemaVersion: 3,
		Status:        "draft",
		Config:        KBSchemaConfig{ExtractModel: "gpt", ClassifyAutoThreshold: 0.7},
		DocTypes:      []KBDocType{{Type: "price_list", RootEntityType: "hospital"}},
		EntityTypes:   []KBEntityType{{Type: "hospital", IsRootCandidate: true, KeyFields: []string{"name"}}},
		RelationshipTypes: []KBRelationType{
			{Type: "hospital_offers_procedure", From: "hospital", To: "procedure", SupersessionIdentity: []string{"procedure"}},
		},
		CreatedAt: "2026-05-22T11:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateMultimodalSchema: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1/knowledge/multimodal-schemas" {
		t.Errorf("path: got %q", seen.path)
	}
	if seen.body["schema_version"] != float64(3) || seen.body["status"] != "draft" {
		t.Errorf("body mismatch: %+v", seen.body)
	}
	// Nested type catalog must serialize with the platform's wire keys.
	dts, ok := seen.body["doc_types"].([]any)
	if !ok || len(dts) != 1 {
		t.Fatalf("body doc_types: got %v", seen.body["doc_types"])
	}
	if dt := dts[0].(map[string]any); dt["root_entity_type"] != "hospital" {
		t.Errorf("body doc_types[0] mismatch: %+v", dt)
	}
	cfg, ok := seen.body["config"].(map[string]any)
	if !ok || cfg["extract_model"] != "gpt" {
		t.Errorf("body config mismatch: %+v", seen.body["config"])
	}
	if res.Schema == nil || res.Schema.SchemaVersion != 3 || res.Schema.Status != "draft" {
		t.Errorf("decoded result mismatch: %+v", res.Schema)
	}
}

func TestKnowledge_ActivateMultimodalSchema_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"active_version": 3,
			"status":         "active",
		})
	})
	client := newTestClient(t, h)

	res, err := client.Knowledge.ActivateMultimodalSchema(context.Background(), "proj-1", 3)
	if err != nil {
		t.Fatalf("ActivateMultimodalSchema: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1/knowledge/multimodal-schemas/3/activate" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.ActiveVersion != 3 || res.Status != "active" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

func TestKnowledge_ReingestDocument_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document_id": "doc-9",
			"mode":        "full",
			"status":      "queued",
		})
	})
	client := newTestClient(t, h)

	res, err := client.Knowledge.ReingestDocument(context.Background(), "proj-1", "doc-9")
	if err != nil {
		t.Fatalf("ReingestDocument: %v", err)
	}
	if seen.method != http.MethodPost {
		t.Errorf("method: got %s, want POST", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1/knowledge/documents/doc-9/reingest" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.DocumentID != "doc-9" || res.Mode != "full" || res.Status != "queued" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

func TestKnowledge_GetDocumentCost_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document_id":    "doc-9",
			"total_cost_usd": 0.4231,
			"document_ai_rows": []map[string]any{
				{"operation": "ocr", "model": "documentai", "cost_usd": 0.12, "pages": 8},
			},
			"llm_rows": []map[string]any{
				{"operation": "extract", "model": "gemini", "cost_usd": 0.3031},
			},
		})
	})
	client := newTestClient(t, h)

	res, err := client.Knowledge.GetDocumentCost(context.Background(), "proj-1", "doc-9")
	if err != nil {
		t.Fatalf("GetDocumentCost: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1/knowledge/documents/doc-9/cost" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.DocumentID != "doc-9" || res.TotalCostUSD != 0.4231 {
		t.Errorf("decoded result mismatch: %+v", res)
	}
	if len(res.DocumentAIRows) != 1 || res.DocumentAIRows[0].Pages != 8 || res.DocumentAIRows[0].Operation != "ocr" {
		t.Errorf("decoded document_ai_rows mismatch: %+v", res.DocumentAIRows)
	}
	if len(res.LLMRows) != 1 || res.LLMRows[0].Model != "gemini" || res.LLMRows[0].CostUSD != 0.3031 {
		t.Errorf("decoded llm_rows mismatch: %+v", res.LLMRows)
	}
}

// ---------------------------------------------------------------------------
// Projects: Delete
// ---------------------------------------------------------------------------

func TestProjects_Delete_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "deleted"})
	})
	client := newTestClient(t, h)

	res, err := client.Projects.Delete(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if seen.method != http.MethodDelete {
		t.Errorf("method: got %s, want DELETE", seen.method)
	}
	if seen.path != "/api/v1/projects/proj-1" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.Status != "deleted" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

// ---------------------------------------------------------------------------
// Tenants: ListOrgKnowledgeNodes
// ---------------------------------------------------------------------------

func TestTenants_ListOrgKnowledgeNodes_URLParamsAndDecode(t *testing.T) {
	var seen struct {
		path, method, rawQuery string
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		seen.rawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"nodes": []map[string]any{
				{"node_id": "org-n1", "node_type": "policy", "label": "Refund"},
			},
			"total": 1,
		})
	})
	client := newTestClient(t, h)

	res, err := client.Tenants.ListOrgKnowledgeNodes(context.Background(), "tenant-abc", &ListOrgKnowledgeNodesOptions{
		NodeType: "policy",
		Limit:    50,
	})
	if err != nil {
		t.Fatalf("ListOrgKnowledgeNodes: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/tenants/tenant-abc/knowledge/org-nodes" {
		t.Errorf("path: got %q", seen.path)
	}
	q := mustParseQuery(t, seen.rawQuery)
	if q.Get("node_type") != "policy" {
		t.Errorf("node_type param: got %q, want policy (query=%q)", q.Get("node_type"), seen.rawQuery)
	}
	if q.Get("limit") != "50" {
		t.Errorf("limit param: got %q, want 50 (query=%q)", q.Get("limit"), seen.rawQuery)
	}
	if res.Total != 1 || len(res.Nodes) != 1 || res.Nodes[0].NodeID != "org-n1" || res.Nodes[0].NodeType != "policy" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
}

func TestTenants_ListOrgKnowledgeNodes_OmitsEmptyParams(t *testing.T) {
	var seen struct{ rawQuery string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.rawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"nodes": []map[string]any{}, "total": 0})
	})
	client := newTestClient(t, h)

	if _, err := client.Tenants.ListOrgKnowledgeNodes(context.Background(), "tenant-abc", nil); err != nil {
		t.Fatalf("ListOrgKnowledgeNodes: %v", err)
	}
	q := mustParseQuery(t, seen.rawQuery)
	if q.Has("node_type") {
		t.Errorf("node_type should be omitted when unset, query=%q", seen.rawQuery)
	}
	if q.Has("limit") {
		t.Errorf("limit should be omitted when unset, query=%q", seen.rawQuery)
	}
}

// ---------------------------------------------------------------------------
// Org: Me
// ---------------------------------------------------------------------------

func TestOrg_Me_URLAndDecode(t *testing.T) {
	var seen struct{ path, method string }
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.path = r.URL.Path
		seen.method = r.Method
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id": "user-1",
			"email":   "user@example.com",
			"orgs": []map[string]any{
				{"org_id": "org-1", "role": "admin", "name": "Acme"},
			},
		})
	})
	client := newTestClient(t, h)

	res, err := client.Org.Me(context.Background())
	if err != nil {
		t.Fatalf("Org.Me: %v", err)
	}
	if seen.method != http.MethodGet {
		t.Errorf("method: got %s, want GET", seen.method)
	}
	if seen.path != "/api/v1/me" {
		t.Errorf("path: got %q", seen.path)
	}
	if res.UserID != "user-1" || res.Email != "user@example.com" {
		t.Errorf("decoded result mismatch: %+v", res)
	}
	if len(res.Orgs) != 1 || res.Orgs[0].OrgID != "org-1" || res.Orgs[0].Role != "admin" {
		t.Errorf("decoded orgs mismatch: %+v", res.Orgs)
	}
}
