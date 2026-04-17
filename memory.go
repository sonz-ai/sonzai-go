package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// MemoryResource provides memory operations for an agent.
type MemoryResource struct {
	http *httpClient
}

// List returns the memory tree for an agent.
func (m *MemoryResource) List(ctx context.Context, agentID string, opts *MemoryListOptions) (*MemoryResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.InstanceID != "" {
			params["instance_id"] = opts.InstanceID
		}
		if opts.ParentID != "" {
			params["parent_id"] = opts.ParentID
		}
		if opts.IncludeContents {
			params["include_contents"] = "true"
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}

	var result MemoryResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory", url.PathEscape(agentID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches agent memories.
func (m *MemoryResource) Search(ctx context.Context, agentID string, opts MemorySearchOptions) (*MemorySearchResponse, error) {
	params := map[string]string{
		"q": opts.Query,
	}
	if opts.InstanceID != "" {
		params["instance_id"] = opts.InstanceID
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}

	var result MemorySearchResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/search", url.PathEscape(agentID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Timeline returns the memory timeline for an agent.
func (m *MemoryResource) Timeline(ctx context.Context, agentID string, opts *MemoryTimelineOptions) (*MemoryTimelineResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.InstanceID != "" {
			params["instance_id"] = opts.InstanceID
		}
		if opts.Start != "" {
			params["start"] = opts.Start
		}
		if opts.End != "" {
			params["end"] = opts.End
		}
	}

	var result MemoryTimelineResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/timeline", url.PathEscape(agentID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListFacts returns atomic facts for an agent, optionally filtered by category.
func (m *MemoryResource) ListFacts(ctx context.Context, agentID string, opts *FactListOptions) (*FactListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.Category != "" {
			params["category"] = opts.Category
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = strconv.Itoa(opts.Offset)
		}
	}

	var result FactListResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts", url.PathEscape(agentID)), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Reset deletes all memory for an agent, optionally scoped to a single user.
func (m *MemoryResource) Reset(ctx context.Context, agentID string, opts *MemoryResetOptions) (*MemoryResetResponse, error) {
	path := fmt.Sprintf("/api/v1/agents/%s/memory", url.PathEscape(agentID))
	if opts != nil && opts.UserID != "" {
		params := url.Values{}
		params.Set("user_id", opts.UserID)
		path += "?" + params.Encode()
	}

	var result MemoryResetResponse
	err := m.http.Delete(ctx, path, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateFactOptions configures a fact creation request.
type CreateFactOptions struct {
	UserID     string                 `json:"user_id,omitempty"`
	Content    string                 `json:"content"`
	FactType   string                 `json:"fact_type,omitempty"`
	Importance *float64               `json:"importance,omitempty"`
	Confidence *float64               `json:"confidence,omitempty"`
	Entities   []string               `json:"entities,omitempty"`
	NodeID     string                 `json:"node_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateFactOptions configures a fact update request.
type UpdateFactOptions struct {
	Content    string                 `json:"content,omitempty"`
	FactType   string                 `json:"fact_type,omitempty"`
	Importance *float64               `json:"importance,omitempty"`
	Confidence *float64               `json:"confidence,omitempty"`
	Entities   []string               `json:"entities,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CreateFact creates a new fact for an agent. Facts created via this method
// are tagged with source_type="manual".
func (m *MemoryResource) CreateFact(ctx context.Context, agentID string, opts CreateFactOptions) (*AtomicFact, error) {
	var result AtomicFact
	err := m.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts", url.PathEscape(agentID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFact updates an existing fact by ID.
func (m *MemoryResource) UpdateFact(ctx context.Context, agentID, factID string, opts UpdateFactOptions) (*AtomicFact, error) {
	var result AtomicFact
	err := m.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts/%s", url.PathEscape(agentID), url.PathEscape(factID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFact removes a fact by ID.
func (m *MemoryResource) DeleteFact(ctx context.Context, agentID, factID string) error {
	return m.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts/%s", url.PathEscape(agentID), url.PathEscape(factID)), nil)
}

// Seed bulk imports initial memories for an agent during setup.
// Unlike GenerationResource.GenerateSeedMemories, this method stores the
// memories you provide directly without any AI generation step.
func (m *MemoryResource) Seed(ctx context.Context, agentID string, opts SeedMemoriesOptions) (*SeedMemoriesResponse, error) {
	var result SeedMemoriesResponse
	err := m.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/seed", url.PathEscape(agentID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWisdomFact deletes a wisdom fact by ID.
func (m *MemoryResource) DeleteWisdomFact(ctx context.Context, agentID, factID string) (*DeleteWisdomResponse, error) {
	var result DeleteWisdomResponse
	err := m.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/wisdom/%s", url.PathEscape(agentID), url.PathEscape(factID)), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWisdomAudit returns the audit trail for a wisdom fact.
func (m *MemoryResource) GetWisdomAudit(ctx context.Context, agentID, factID string) (*WisdomAuditResponse, error) {
	var result WisdomAuditResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/wisdom/audit/%s", url.PathEscape(agentID), url.PathEscape(factID)), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFactHistory returns the version history for a specific fact.
func (m *MemoryResource) GetFactHistory(ctx context.Context, agentID, factID string) (*FactHistoryResponse, error) {
	var result FactHistoryResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/fact/%s/history", url.PathEscape(agentID), url.PathEscape(factID)), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
