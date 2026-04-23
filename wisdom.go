package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// WisdomResource provides attributed-wisdom operations — the cross-user,
// agent-scoped knowledge tier gated by the WisdomPublicSharing capability.
// Writes go through a privacy-floor blocklist + an SP1-routed semantic
// validator so sensitive categories (compensation, health, politics, etc.)
// never persist regardless of provenance.
type WisdomResource struct {
	http *httpClient
}

// AttributedFact is a person/entity-attributed wisdom entry.
type AttributedFact struct {
	EntityType        string     `json:"entity_type"`
	EntityID          string     `json:"entity_id"`
	EntityDisplayName string     `json:"entity_display_name,omitempty"`
	Category          string     `json:"category"`
	Value             string     `json:"value"`
	Confidence        float64    `json:"confidence,omitempty"`
	Source            string     `json:"source"`
	SourceRef         string     `json:"source_ref,omitempty"`
	ObservedAt        time.Time  `json:"observed_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// AttributedRelation is a directed edge between two attributed entities.
type AttributedRelation struct {
	ID         string            `json:"id,omitempty"`
	FromType   string            `json:"from_type"`
	FromID     string            `json:"from_id"`
	EdgeType   string            `json:"edge_type"`
	ToType     string            `json:"to_type"`
	ToID       string            `json:"to_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Source     string            `json:"source"`
	SourceRef  string            `json:"source_ref,omitempty"`
	ObservedAt time.Time         `json:"observed_at"`
}

// CreateWisdomAttributedBody is the create-fact request.
type CreateWisdomAttributedBody struct {
	EntityType        string     `json:"entity_type"`
	EntityID          string     `json:"entity_id"`
	EntityDisplayName string     `json:"entity_display_name,omitempty"`
	Category          string     `json:"category"`
	Value             string     `json:"value"`
	Confidence        float64    `json:"confidence,omitempty"`
	SourceRef         string     `json:"source_ref,omitempty"`
	ObservedAt        *time.Time `json:"observed_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// CreateWisdomRelationBody is the create-relation request.
type CreateWisdomRelationBody struct {
	FromType   string            `json:"from_type"`
	FromID     string            `json:"from_id"`
	EdgeType   string            `json:"edge_type"`
	ToType     string            `json:"to_type"`
	ToID       string            `json:"to_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	SourceRef  string            `json:"source_ref,omitempty"`
	ObservedAt *time.Time        `json:"observed_at,omitempty"`
}

// ReplaceAttributedBody is the PUT payload for overwriting an existing
// (entity, category) fact. EntityType/EntityID/Category come from the URL.
type ReplaceAttributedBody struct {
	EntityDisplayName string     `json:"entity_display_name,omitempty"`
	Value             string     `json:"value"`
	Confidence        float64    `json:"confidence,omitempty"`
	SourceRef         string     `json:"source_ref,omitempty"`
	ObservedAt        *time.Time `json:"observed_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// WisdomImportBody is the CSV/JSON bulk-import payload.
type WisdomImportBody struct {
	Format string `json:"format"` // "csv" | "json"
	Data   string `json:"data"`
	DryRun bool   `json:"dry_run,omitempty"`
}

// WisdomImportReject is one rejected row from a bulk import.
type WisdomImportReject struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// WisdomImportResponse reports bulk-import outcome.
type WisdomImportResponse struct {
	Total    int64                `json:"total"`
	Accepted int64                `json:"accepted"`
	Rejected []WisdomImportReject `json:"rejected"`
}

// DisclosureEntry is one row of the audit log.
type DisclosureEntry struct {
	AgentID     string          `json:"agent_id"`
	RecordedAt  time.Time       `json:"recorded_at"`
	TurnID      string          `json:"turn_id"`
	UserID      string          `json:"user_id,omitempty"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Category    string          `json:"category"`
	Fact        *AttributedFact `json:"fact,omitempty"`
	Decision    string          `json:"decision"` // "disclosed" | "redacted"
	DecisionWhy string          `json:"decision_why,omitempty"`
}

// ListAttributedResponse is returned by list calls.
type ListAttributedResponse struct {
	Facts []AttributedFact `json:"facts"`
}

// ListRelationsResponse is returned by ListRelations.
type ListRelationsResponse struct {
	Relations []AttributedRelation `json:"relations"`
}

// ListAuditResponse is returned by ListAudit.
type ListAuditResponse struct {
	Entries []DisclosureEntry `json:"entries"`
}

// ListAttributed returns every attributed fact the agent has recorded.
func (r *WisdomResource) ListAttributed(ctx context.Context, agentID string) (*ListAttributedResponse, error) {
	var out ListAttributedResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAttributed records a new attributed fact.
func (r *WisdomResource) CreateAttributed(ctx context.Context, agentID string, body CreateWisdomAttributedBody) (*AttributedFact, error) {
	var out AttributedFact
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAttributedByEntity returns attributed facts for a specific (type, id).
func (r *WisdomResource) ListAttributedByEntity(ctx context.Context, agentID, entityType, entityID string) (*ListAttributedResponse, error) {
	var out ListAttributedResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed/%s/%s", url.PathEscape(agentID), url.PathEscape(entityType), url.PathEscape(entityID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplaceAttributed overwrites the value of a (entity, category) fact.
func (r *WisdomResource) ReplaceAttributed(ctx context.Context, agentID, entityType, entityID, category string, body ReplaceAttributedBody) (*AttributedFact, error) {
	var out AttributedFact
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed/%s/%s/%s",
		url.PathEscape(agentID), url.PathEscape(entityType), url.PathEscape(entityID), url.PathEscape(category))
	if err := r.http.Put(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAttributed removes a fact. Idempotent.
func (r *WisdomResource) DeleteAttributed(ctx context.Context, agentID, entityType, entityID, category string) (*DeleteWisdomResponse, error) {
	var out DeleteWisdomResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed/%s/%s/%s",
		url.PathEscape(agentID), url.PathEscape(entityType), url.PathEscape(entityID), url.PathEscape(category))
	if err := r.http.Delete(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ImportAttributed bulk-imports CSV or JSON facts.
func (r *WisdomResource) ImportAttributed(ctx context.Context, agentID string, body WisdomImportBody) (*WisdomImportResponse, error) {
	var out WisdomImportResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/attributed/import", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRelations returns attributed directed edges for the agent.
func (r *WisdomResource) ListRelations(ctx context.Context, agentID string) (*ListRelationsResponse, error) {
	var out ListRelationsResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/relations", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateRelation records a directed relation between two entities.
func (r *WisdomResource) CreateRelation(ctx context.Context, agentID string, body CreateWisdomRelationBody) (*AttributedRelation, error) {
	var out AttributedRelation
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/relations", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRelation removes a relation by ID. Idempotent.
func (r *WisdomResource) DeleteRelation(ctx context.Context, agentID, relationID string) error {
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/relations/%s", url.PathEscape(agentID), url.PathEscape(relationID))
	return r.http.Delete(ctx, path, nil)
}

// ListAudit reads the disclosure audit trail.
type ListAuditOptions struct {
	From  time.Time
	To    time.Time
	Limit int
}

// ListAudit returns the disclosure audit trail for an agent.
func (r *WisdomResource) ListAudit(ctx context.Context, agentID string, opts *ListAuditOptions) (*ListAuditResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if !opts.From.IsZero() {
			params["from"] = opts.From.UTC().Format(time.RFC3339)
		}
		if !opts.To.IsZero() {
			params["to"] = opts.To.UTC().Format(time.RFC3339)
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}
	var out ListAuditResponse
	path := fmt.Sprintf("/api/v1/agents/%s/wisdom/audit", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
