package sonzai

import (
	"context"
	"fmt"
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
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory", agentID), params, &result)
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
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/search", agentID), params, &result)
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
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/timeline", agentID), params, &result)
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
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/facts", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Reset deletes all memory for an agent, optionally scoped to a single user.
func (m *MemoryResource) Reset(ctx context.Context, agentID string, opts *MemoryResetOptions) (*MemoryResetResponse, error) {
	path := fmt.Sprintf("/api/v1/agents/%s/memory", agentID)
	if opts != nil && opts.UserID != "" {
		path += "?user_id=" + opts.UserID
	}

	var result MemoryResetResponse
	err := m.http.Delete(ctx, path, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFactHistory returns the version history for a specific fact.
func (m *MemoryResource) GetFactHistory(ctx context.Context, agentID, factID string) (*FactHistoryResponse, error) {
	var result FactHistoryResponse
	err := m.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/fact/%s/history", agentID, factID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
