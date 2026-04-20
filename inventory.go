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
	Resolved     bool           `json:"resolved"`
	KBNodeID     string         `json:"kb_node_id,omitempty"`
	KBLabel      string         `json:"kb_label,omitempty"`
	KBProperties map[string]any `json:"kb_properties,omitempty"`
}

// KBCandidate represents a candidate KB node for disambiguation.
type KBCandidate struct {
	KBNodeID   string         `json:"kb_node_id"`
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties,omitempty"`
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
	FactID           string         `json:"fact_id"`
	ItemLabel        string         `json:"item_label"`
	KBNodeID         string         `json:"kb_node_id,omitempty"`
	UserProperties   map[string]any `json:"user_properties"`
	MarketProperties map[string]any `json:"market_properties,omitempty"`
	GainLoss         *float64       `json:"gain_loss,omitempty"`
}

// InventoryGroupResult represents a group in an aggregate query.
type InventoryGroupResult struct {
	Group  string         `json:"group"`
	Values map[string]any `json:"values"`
}

// InventoryQueryResponse is the response from an inventory query.
type InventoryQueryResponse struct {
	Items      []InventoryItem        `json:"items"`
	TotalItems int                    `json:"total_items"`
	NextCursor string                 `json:"next_cursor,omitempty"`
	Totals     map[string]any         `json:"totals,omitempty"`
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
	FactID       string         `json:"fact_id"`
	Content      string         `json:"content"`
	FactType     string         `json:"fact_type"`
	Importance   float64        `json:"importance"`
	Confidence   float64        `json:"confidence"`
	Entity       string         `json:"entity,omitempty"`
	SourceType   string         `json:"source_type,omitempty"`
	MentionCount int            `json:"mention_count"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

// ListAllFactsResponse is the response from listing all facts for an agent+user.
type ListAllFactsResponse struct {
	Facts []StoredFact `json:"facts"`
	Total int          `json:"total"`
}

// ---------------------------------------------------------------------------
// Request Options
// ---------------------------------------------------------------------------

// InventoryUpdateOptions configures an inventory write operation.
type InventoryUpdateOptions struct {
	Action      string         `json:"action"`                // "add", "update", "remove"
	ItemType    string         `json:"item_type"`             // e.g. "pokemon_card", "property"
	Description string         `json:"description,omitempty"` // Natural language (for KB search)
	KBNodeID    string         `json:"kb_node_id,omitempty"`  // If already resolved
	Properties  map[string]any `json:"properties,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	InstanceID  string         `json:"-"` // Query param for multi-instance scoping
}

// InventoryQueryOptions configures an inventory query.
// Filters format: "field:op:value,field:op:value" where op is one of
// eq, neq, gt, gte, lt, lte, in (pipe-separated values), contains.
// Example: "grade:eq:mint,market_price:gte:100"
type InventoryQueryOptions struct {
	Mode         string // "list", "value", "aggregate"
	ItemType     string
	Query        string
	ProjectID    string
	Filters      string // Structured metadata filtering, e.g. "grade:eq:mint,market_price:gte:100"
	SortBy       string // Metadata field to sort by
	SortOrder    string // "asc" or "desc"
	Aggregations string // e.g. "market_price:sum,*:count"
	GroupBy      string
	Limit        int
	Offset       int
	Cursor       string // Base64-encoded pagination cursor (takes precedence over Offset)
	InstanceID   string
}

// InventoryBatchItem represents a single item in a batch import.
type InventoryBatchItem struct {
	ItemType    string         `json:"item_type"`
	Description string         `json:"description,omitempty"`
	KBNodeID    string         `json:"kb_node_id,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
}

// InventoryBatchImportOptions configures a batch inventory import.
type InventoryBatchImportOptions struct {
	Items      []InventoryBatchItem `json:"items"`
	ProjectID  string               `json:"project_id,omitempty"`
	InstanceID string               `json:"-"` // Query param for multi-instance scoping
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
	path := fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory", agentID, url.PathEscape(userID))
	if opts.InstanceID != "" {
		path += "?instance_id=" + url.QueryEscape(opts.InstanceID)
	}
	var result InventoryUpdateResponse
	err := inv.http.Post(ctx, path, opts, &result)
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
	if opts.Filters != "" {
		params["filters"] = opts.Filters
	}
	if opts.SortBy != "" {
		params["sort_by"] = opts.SortBy
	}
	if opts.SortOrder != "" {
		params["sort_order"] = opts.SortOrder
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
	if opts.Offset > 0 {
		params["offset"] = strconv.Itoa(opts.Offset)
	}
	if opts.Cursor != "" {
		params["cursor"] = opts.Cursor
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
	path := fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/batch", agentID, url.PathEscape(userID))
	if opts.InstanceID != "" {
		path += "?instance_id=" + url.QueryEscape(opts.InstanceID)
	}
	var result InventoryBatchImportResponse
	err := inv.http.Post(ctx, path, opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DirectUpdate updates an inventory fact's properties by fact ID.
// Pass instanceID to scope to a specific backend instance (empty string for default).
func (inv *InventoryResource) DirectUpdate(ctx context.Context, agentID, userID, factID string, properties map[string]any, instanceID string) (*InventoryDirectUpdateResponse, error) {
	path := fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/%s", agentID, url.PathEscape(userID), url.PathEscape(factID))
	if instanceID != "" {
		path += "?instance_id=" + url.QueryEscape(instanceID)
	}
	body := map[string]any{"properties": properties}
	var result InventoryDirectUpdateResponse
	err := inv.http.Put(ctx, path, body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DirectDelete removes an inventory item by fact ID.
// Pass instanceID to scope to a specific backend instance (empty string for default).
func (inv *InventoryResource) DirectDelete(ctx context.Context, agentID, userID, factID string, instanceID string) (*InventoryDirectUpdateResponse, error) {
	path := fmt.Sprintf("/api/v1/agents/%s/users/%s/inventory/%s", agentID, url.PathEscape(userID), url.PathEscape(factID))
	if instanceID != "" {
		path += "?instance_id=" + url.QueryEscape(instanceID)
	}
	var result InventoryDirectUpdateResponse
	err := inv.http.Delete(ctx, path, &result)
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
