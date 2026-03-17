package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// Seed imports bulk initial memories for an agent.
func (m *MemoryResource) Seed(ctx context.Context, agentID string, params SeedMemoriesParams) (*SeedMemoriesResult, error) {
	var result SeedMemoriesResult
	err := m.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/seed", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListFacts returns stored facts for an agent.
func (m *MemoryResource) ListFacts(ctx context.Context, agentID string, opts *ListFactsOptions) (*ListFactsResult, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.InstanceID != "" {
			params["instance_id"] = opts.InstanceID
		}
		if opts.FactType != "" {
			params["fact_type"] = opts.FactType
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}

	var result ListFactsResult
	err := m.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Reset deletes all memories for an agent-user pair.
func (m *MemoryResource) Reset(ctx context.Context, agentID string, params ResetMemoryParams) (*ResetMemoryResult, error) {
	var result ResetMemoryResult
	err := m.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/reset", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

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
	err := m.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory", agentID), params, &result)
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
	err := m.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/search", agentID), params, &result)
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
	err := m.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/timeline", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
