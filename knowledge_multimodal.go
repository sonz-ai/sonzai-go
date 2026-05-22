// knowledge_multimodal.go — Plan 2/3 multimodal KB ingestion + retrieval
// surface for the Go SDK. These methods extend KnowledgeResource with the
// new endpoints introduced by the multimodal pipeline:
//
//   PATCH  /knowledge/documents/{id}/classification — resolve needs_classification
//   GET    /knowledge/facts                          — list active facts
//   GET    /knowledge/facts/active                   — fetch active fact for a tuple
//   GET    /knowledge/facts/history                  — version chain for a tuple
//   GET    /knowledge/entities/{type}/{key}          — kb_get_entity
//   GET    /knowledge/traverse                       — kb_traverse
//   POST   /knowledge/compare                        — kb_compare
//
// Spec: docs/superpowers/specs/2026-05-22-multimodal-kb-ingestion-design.md §6.4
package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// KBFact is a single relationship instance with provenance. Returned by the
// /facts endpoints and the 4 retrieval tools.
type KBFact struct {
	FactID               string         `json:"fact_id"`
	FromNodeID           string         `json:"from_node_id"`
	ToNodeID             string         `json:"to_node_id"`
	RelationType         string         `json:"relation_type"`
	Properties           map[string]any `json:"properties,omitempty"`
	SourceDocumentID     string         `json:"source_document_id"`
	SourcePage           int            `json:"source_page"`
	SourceSnippet        string         `json:"source_snippet"`
	ExtractionConfidence float64        `json:"extraction_confidence"`
	EffectiveDate        string         `json:"effective_date"` // ISO 8601
	Version              int            `json:"version"`
	IsActive             bool           `json:"is_active"`
	CreatedAt            string         `json:"created_at,omitempty"`
}

// KBFactListResponse is the paginated /facts response.
type KBFactListResponse struct {
	Facts         []*KBFact `json:"facts"`
	NextPageToken string    `json:"next_page_token,omitempty"`
}

// KBFactHistoryResponse is the version-chain response.
type KBFactHistoryResponse struct {
	Versions []*KBFact `json:"versions"`
}

// KBGetEntityResponse is the kb_get_entity payload.
type KBGetEntityResponse struct {
	EntityType    string         `json:"entity_type"`
	EntityKey     map[string]any `json:"entity_key"`
	EntityNodeID  string         `json:"entity_node_id"`
	OutgoingFacts []*KBFact      `json:"outgoing_facts"`
	IncomingFacts []*KBFact      `json:"incoming_facts"`
}

// KBTraverseTuple is one fact returned by kb_traverse with its depth.
type KBTraverseTuple struct {
	Depth int    `json:"depth"`
	Fact  KBFact `json:"fact"`
}

// KBTraverseResponse is the kb_traverse payload.
type KBTraverseResponse struct {
	Facts []KBTraverseTuple `json:"facts"`
}

// KBEntityRef identifies an entity by its (type, key) tuple.
type KBEntityRef struct {
	Type string         `json:"type"`
	Key  map[string]any `json:"key"`
}

// KBCompareRow is one row in a kb_compare response.
type KBCompareRow struct {
	Entity        KBEntityRef `json:"entity"`
	Value         any         `json:"value,omitempty"`
	Fact          *KBFact     `json:"fact,omitempty"`
	Missing       bool        `json:"missing"`
	MissingReason string      `json:"missing_reason,omitempty"`
}

// KBCompareResponse is the kb_compare payload.
type KBCompareResponse struct {
	Rows []KBCompareRow `json:"rows"`
}

// PatchClassificationRequest resolves a needs_classification document with
// a human-confirmed root entity.
type PatchClassificationRequest struct {
	RootEntity KBEntityRef `json:"root_entity"`
}

// PatchClassificationResponse is the result of resolving classification.
type PatchClassificationResponse struct {
	Status     string `json:"status"`
	DocumentID string `json:"document_id"`
}

// CompareRequest is the body of POST /knowledge/compare.
type CompareRequest struct {
	Entities     []KBEntityRef `json:"entities"`
	ViaRelation  string        `json:"via_relation"`
	TargetEntity KBEntityRef   `json:"target_entity"`
	PropertyPath string        `json:"property_path"`
}

// ---------------------------------------------------------------------------
// Classification resolution
// ---------------------------------------------------------------------------

// PatchDocumentClassification resolves a document parked in
// needs_classification by recording the human-confirmed root entity and
// triggering Pass B extract.
func (k *KnowledgeResource) PatchDocumentClassification(
	ctx context.Context, projectID, documentID string,
	body PatchClassificationRequest,
) (*PatchClassificationResponse, error) {
	var resp PatchClassificationResponse
	path := fmt.Sprintf("/api/v1/projects/%s/knowledge/documents/%s/classification", projectID, documentID)
	err := k.http.Patch(ctx, path, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// /facts endpoints
// ---------------------------------------------------------------------------

// ListFactsOptions configures a fact list request.
type ListFactsOptions struct {
	Limit     int
	PageToken string
}

// ListFacts returns active KB facts for a project (paginated).
func (k *KnowledgeResource) ListFacts(ctx context.Context, projectID string, opts *ListFactsOptions) (*KBFactListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.PageToken != "" {
			params["page_token"] = opts.PageToken
		}
	}
	var result KBFactListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/facts", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetActiveFact returns the active fact for a (from, to, relation) tuple, or nil.
func (k *KnowledgeResource) GetActiveFact(ctx context.Context, projectID, fromNodeID, toNodeID, relationType string) (*KBFact, error) {
	params := map[string]string{
		"from_node_id":  fromNodeID,
		"to_node_id":    toNodeID,
		"relation_type": relationType,
	}
	var result struct {
		Fact *KBFact `json:"fact"`
	}
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/facts/active", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return result.Fact, nil
}

// GetFactHistory returns the version chain for a (from, to, relation) tuple.
func (k *KnowledgeResource) GetFactHistory(ctx context.Context, projectID, fromNodeID, toNodeID, relationType string) (*KBFactHistoryResponse, error) {
	params := map[string]string{
		"from_node_id":  fromNodeID,
		"to_node_id":    toNodeID,
		"relation_type": relationType,
	}
	var result KBFactHistoryResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/facts/history", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// 4-tool retrieval surface
// ---------------------------------------------------------------------------

// GetEntity (kb_get_entity) looks up an entity by (type, key) and returns
// all active facts attached to it.
func (k *KnowledgeResource) GetEntity(ctx context.Context, projectID, entityType string, entityKey map[string]any) (*KBGetEntityResponse, error) {
	keyJSON, err := json.Marshal(entityKey)
	if err != nil {
		return nil, fmt.Errorf("marshal entity key: %w", err)
	}
	path := fmt.Sprintf("/api/v1/projects/%s/knowledge/entities/%s/%s",
		projectID, url.PathEscape(entityType), url.PathEscape(string(keyJSON)))
	var result KBGetEntityResponse
	err = k.http.Get(ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TraverseOptions configures a kb_traverse call.
type TraverseOptions struct {
	Direction string // "outbound" | "inbound" | "both" (default outbound)
	MaxDepth  int    // 1..3 (default 1)
}

// Traverse (kb_traverse) walks the graph from a starting entity along
// relationType up to MaxDepth.
func (k *KnowledgeResource) Traverse(ctx context.Context, projectID string, from KBEntityRef, relationType string, opts *TraverseOptions) (*KBTraverseResponse, error) {
	keyJSON, err := json.Marshal(from.Key)
	if err != nil {
		return nil, fmt.Errorf("marshal from key: %w", err)
	}
	params := map[string]string{
		"from_type":     from.Type,
		"from_key":      string(keyJSON),
		"relation_type": relationType,
	}
	if opts != nil {
		if opts.Direction != "" {
			params["direction"] = opts.Direction
		}
		if opts.MaxDepth > 0 {
			params["max_depth"] = strconv.Itoa(opts.MaxDepth)
		}
	}
	var result KBTraverseResponse
	err = k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/traverse", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Compare (kb_compare) returns property values across multiple entities
// connected to a shared target via the same relation.
func (k *KnowledgeResource) Compare(ctx context.Context, projectID string, req CompareRequest) (*KBCompareResponse, error) {
	var result KBCompareResponse
	err := k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/compare", projectID), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RecommendPromptAddendum returns the cite-and-verify system-prompt addendum
// tenants should paste into agent prompts that use kb_search / kb_get_entity /
// kb_traverse / kb_compare. Returned inline (no network call) since the text
// is stable across the SDK release.
//
// Spec: docs/superpowers/specs/2026-05-22-multimodal-kb-ingestion-design.md §6.2
func (k *KnowledgeResource) RecommendPromptAddendum() string {
	return `
You have access to four knowledge-base tools (kb_search, kb_get_entity, kb_traverse, kb_compare). Every fact they return carries source_document_id, source_page, source_snippet, effective_date, version, extraction_confidence, fact_id.

When you assert a fact in your answer:
1. Re-read the source_snippet character-by-character. If it does not support your claim, drop the claim or call another tool.
2. Cite the source: include the source_document_id and source_page in your answer (e.g. "[doc: abc-123, page 4]").
3. Prefer facts with the most recent effective_date when multiple versions exist.

If no snippet clearly supports the answer the user asked for, say so explicitly and stop. Do not guess. Do not invent fact IDs between tool calls — use the exact fact_id strings the tools returned.
`
}
