package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// InventoryResource provides inventory/asset tracking operations for an agent.
type InventoryResource struct {
	http *httpClient
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// KBResolutionInfo contains KB node resolution details from an inventory write.
type KBResolutionInfo struct {
	Resolved     bool                   `json:"resolved"`
	KBNodeID     string                 `json:"kb_node_id,omitempty"`
	KBLabel      string                 `json:"kb_label,omitempty"`
	KBProperties map[string]interface{} `json:"kb_properties,omitempty"`
}

// KBCandidate represents a candidate KB node for disambiguation.
type KBCandidate struct {
	KBNodeID   string                 `json:"kb_node_id"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// InventoryUpdateResponse is the response from an inventory write operation.
type InventoryUpdateResponse struct {
	Status       string            `json:"status"`
	FactID       string            `json:"fact_id,omitempty"`
	KBResolution *KBResolutionInfo `json:"kb_resolution,omitempty"`
	Candidates   []KBCandidate     `json:"candidates,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// InventoryItem represents a single item in an inventory query response.
type InventoryItem struct {
	FactID           string                 `json:"fact_id"`
	ItemLabel        string                 `json:"item_label"`
	KBNodeID         string                 `json:"kb_node_id,omitempty"`
	UserProperties   map[string]interface{} `json:"user_properties"`
	MarketProperties map[string]interface{} `json:"market_properties,omitempty"`
	GainLoss         *float64               `json:"gain_loss,omitempty"`
}

// InventoryGroupResult represents a group in an aggregate query.
type InventoryGroupResult struct {
	Group  string                 `json:"group"`
	Values map[string]interface{} `json:"values"`
}

// InventoryQueryResponse is the response from an inventory query.
type InventoryQueryResponse struct {
	Items      []InventoryItem        `json:"items"`
	TotalItems int                    `json:"total_items"`
	Totals     map[string]interface{} `json:"totals,omitempty"`
	Groups     []InventoryGroupResult `json:"groups,omitempty"`
}

// InventoryBatchImportResponse is the response from a batch inventory import.
type InventoryBatchImportResponse struct {
	Status string `json:"status"`
	Added  int    `json:"added"`
	Failed int    `json:"failed"`
	Total  int    `json:"total"`
	Error  string `json:"error,omitempty"`
}

// InventoryDirectUpdateResponse is the response from a direct update/delete.
type InventoryDirectUpdateResponse struct {
	Status string `json:"status"`
	FactID string `json:"fact_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

// StoredFact represents a stored fact returned by the full fact recall endpoint.
type StoredFact struct {
	FactID       string                 `json:"fact_id"`
	Content      string                 `json:"content"`
	FactType     string                 `json:"fact_type"`
	Importance   float64                `json:"importance"`
	Confidence   float64                `json:"confidence"`
	Entity       string                 `json:"entity,omitempty"`
	SourceType   string                 `json:"source_type,omitempty"`
	MentionCount int                    `json:"mention_count"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// ListAllFactsResponse is the response from listing all facts for an agent+user.
type ListAllFactsResponse struct {
	Facts      []StoredFact `json:"facts"`
	TotalCount int          `json:"total_count"`
}

// ---------------------------------------------------------------------------
// Request Options
// ---------------------------------------------------------------------------

// InventoryUpdateOptions configures an inventory write operation.
type InventoryUpdateOptions struct {
	Action      string                 `json:"action"`                // "add", "update", "remove"
	ItemType    string                 `json:"item_type"`             // e.g. "pokemon_card", "property"
	Description string                 `json:"description,omitempty"` // Natural language (for KB search)
	KBNodeID    string                 `json:"kb_node_id,omitempty"`  // If already resolved
	Properties  map[string]interface{} `json:"properties,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
}

// InventoryQueryOptions configures an inventory query.
type InventoryQueryOptions struct {
	Mode         string // "list", "value", "aggregate"
	ItemType     string
	Query        string
	ProjectID    string
	Aggregations string // e.g. "market_price:sum,*:count"
	GroupBy      string
	Limit        int
	InstanceID   string
}

// InventoryBatchItem represents a single item in a batch import.
type InventoryBatchItem struct {
	ItemType    string                 `json:"item_type"`
	Description string                 `json:"description,omitempty"`
	KBNodeID    string                 `json:"kb_node_id,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// InventoryBatchImportOptions configures a batch inventory import.
type InventoryBatchImportOptions struct {
	Items     []InventoryBatchItem `json:"items"`
	ProjectID string               `json:"project_id,omitempty"`
}

// ListAllFactsOptions configures a full fact recall query.
type ListAllFactsOptions struct {
	HasMetadata bool
	ItemType    string
	Limit       int
	InstanceID  string
}

// ---------------------------------------------------------------------------
// Methods
// ---------------------------------------------------------------------------

// Update adds, updates, or removes an inventory item.
func (inv *InventoryResource) Update(ctx context.Context, agentID, userID string, opts InventoryUpdateOptions) (*InventoryUpdateResponse, error) {
	var result InventoryUpdateResponse
	err := inv.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory", agentID, url.PathEscape(userID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Query queries a user's inventory with optional KB-joined valuations and aggregation.
func (inv *InventoryResource) Query(ctx context.Context, agentID, userID string, opts InventoryQueryOptions) (*InventoryQueryResponse, error) {
	params := map[string]string{}
	if opts.Mode != "" {
		params["mode"] = opts.Mode
	}
	if opts.ItemType != "" {
		params["item_type"] = opts.ItemType
	}
	if opts.Query != "" {
		params["query"] = opts.Query
	}
	if opts.ProjectID != "" {
		params["project_id"] = opts.ProjectID
	}
	if opts.Aggregations != "" {
		params["aggregations"] = opts.Aggregations
	}
	if opts.GroupBy != "" {
		params["group_by"] = opts.GroupBy
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.InstanceID != "" {
		params["instance_id"] = opts.InstanceID
	}

	var result InventoryQueryResponse
	err := inv.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory", agentID, url.PathEscape(userID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchImport adds multiple inventory items in a single request (up to 1000).
func (inv *InventoryResource) BatchImport(ctx context.Context, agentID, userID string, opts InventoryBatchImportOptions) (*InventoryBatchImportResponse, error) {
	var result InventoryBatchImportResponse
	err := inv.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/batch", agentID, url.PathEscape(userID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DirectUpdate updates an inventory fact's properties by fact ID.
func (inv *InventoryResource) DirectUpdate(ctx context.Context, agentID, userID, factID string, properties map[string]interface{}) (*InventoryDirectUpdateResponse, error) {
	body := map[string]interface{}{"properties": properties}
	var result InventoryDirectUpdateResponse
	err := inv.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/%s", agentID, url.PathEscape(userID), factID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DirectDelete removes an inventory item by fact ID.
func (inv *InventoryResource) DirectDelete(ctx context.Context, agentID, userID, factID string) (*InventoryDirectUpdateResponse, error) {
	var result InventoryDirectUpdateResponse
	err := inv.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/%s", agentID, url.PathEscape(userID), factID), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAllFacts returns all active facts for an agent+user pair. Supports
// filtering by metadata presence and item_type. Bypasses token-budgeted recall.
func (inv *InventoryResource) ListAllFacts(ctx context.Context, agentID, userID string, opts ListAllFactsOptions) (*ListAllFactsResponse, error) {
	params := map[string]string{}
	if opts.HasMetadata {
		params["has_metadata"] = "true"
	}
	if opts.ItemType != "" {
		params["metadata.item_type"] = opts.ItemType
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.InstanceID != "" {
		params["instance_id"] = opts.InstanceID
	}

	var result ListAllFactsResponse
	err := inv.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/facts", agentID, url.PathEscape(userID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
