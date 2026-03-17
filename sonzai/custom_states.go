package sonzai

import (
	"context"
	"fmt"
)

// CustomStatesResource provides custom state operations for an agent.
type CustomStatesResource struct {
	http *httpClient
}

// List returns all custom states for an agent.
func (c *CustomStatesResource) List(ctx context.Context, agentID string) (*CustomStateListResponse, error) {
	var result CustomStateListResponse
	err := c.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states", agentID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new custom state.
func (c *CustomStatesResource) Create(ctx context.Context, agentID string, params CustomStateCreateParams) (*CustomState, error) {
	var result CustomState
	err := c.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing custom state.
func (c *CustomStatesResource) Update(ctx context.Context, agentID, stateID string, params CustomStateUpdateParams) (*CustomState, error) {
	var result CustomState
	err := c.http.put(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states/%s", agentID, stateID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a custom state.
func (c *CustomStatesResource) Delete(ctx context.Context, agentID, stateID string) error {
	return c.http.del(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states/%s", agentID, stateID), nil)
}
