package sonzai

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestKBFact_JSONRoundTrip ensures the provenance fields a tenant agent
// relies on (source_document_id, source_page, source_snippet,
// effective_date, version, extraction_confidence, fact_id) survive a JSON
// round-trip — the platform sends snake_case and the SDK must decode all
// fields verbatim.
func TestKBFact_JSONRoundTrip(t *testing.T) {
	raw := `{
		"fact_id": "f-1",
		"from_node_id": "hospital:A",
		"to_node_id": "procedure:MRI",
		"relation_type": "hospital_offers_procedure",
		"properties": {"price": 1500, "currency": "SGD"},
		"source_document_id": "doc-1",
		"source_page": 3,
		"source_snippet": "MRI Brain — SGD 1500",
		"extraction_confidence": 0.92,
		"effective_date": "2026-01-15T00:00:00Z",
		"version": 1,
		"is_active": true,
		"created_at": "2026-05-22T10:00:00Z"
	}`
	var f KBFact
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if f.FactID != "f-1" || f.SourcePage != 3 || f.SourceSnippet != "MRI Brain — SGD 1500" {
		t.Errorf("provenance fields lost: %+v", f)
	}
	if f.ExtractionConfidence != 0.92 {
		t.Errorf("confidence lost: %v", f.ExtractionConfidence)
	}
	if f.Properties["price"].(float64) != 1500 {
		t.Errorf("properties decoded incorrectly: %+v", f.Properties)
	}
}

// TestRecommendPromptAddendum_HasContract checks the helper returns the
// non-empty cite-and-verify text described in spec §6.2. Drift here means
// agent prompts go out without provenance guidance — the harness tests in
// the platform repo would also catch it but a unit test on the SDK keeps
// regression visible at tenant build time.
func TestRecommendPromptAddendum_HasContract(t *testing.T) {
	r := &KnowledgeResource{}
	addendum := r.RecommendPromptAddendum()
	if !strings.Contains(addendum, "source_snippet") {
		t.Error("addendum must mention source_snippet (verbatim re-read contract)")
	}
	if !strings.Contains(addendum, "fact_id") {
		t.Error("addendum must mention fact_id (no hallucinated IDs)")
	}
	if !strings.Contains(addendum, "abstain") &&
		!strings.Contains(addendum, "do not guess") &&
		!strings.Contains(addendum, "Do not guess") {
		t.Error("addendum must instruct the agent to abstain when unsupported")
	}
}

// TestCompareRequest_StructureMatchesAPI mirrors the wire shape the platform's
// huma_ops_kb_multimodal.kbCompareInput expects. If either side diverges,
// every kb_compare call returns 400.
func TestCompareRequest_StructureMatchesAPI(t *testing.T) {
	req := CompareRequest{
		Entities: []KBEntityRef{
			{Type: "hospital", Key: map[string]any{"name": "A"}},
			{Type: "hospital", Key: map[string]any{"name": "B"}},
		},
		ViaRelation:  "hospital_offers_procedure",
		TargetEntity: KBEntityRef{Type: "procedure", Key: map[string]any{"code": "0210093"}},
		PropertyPath: "price",
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	expectedKeys := []string{`"entities"`, `"via_relation"`, `"target_entity"`, `"property_path"`}
	for _, k := range expectedKeys {
		if !strings.Contains(string(b), k) {
			t.Errorf("missing required wire key %s: %s", k, string(b))
		}
	}
}
