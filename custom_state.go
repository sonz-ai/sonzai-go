package sonzai

import (
	"context"
	"fmt"
)

// CustomStateResource provides custom state operations for an agent.
type CustomStateResource struct {
	http *httpClient
}

// CustomStateListOptions configures a custom state list request.
type CustomStateListOptions struct {
	Scope      string // "global", "user", or "instance"
	UserID     string
	InstanceID string
}

// CustomStateCreateOptions configures a custom state creation request.
type CustomStateCreateOptions struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Scope       string      `json:"scope,omitempty"`        // "global" (default), "user", or "instance"
	ContentType string      `json:"content_type,omitempty"` // "text" (default), "json", or "binary"
	UserID      string      `json:"user_id,omitempty"`      // required if scope is "user"
	InstanceID  string      `json:"instance_id,omitempty"`
}

// CustomStateUpdateOptions configures a custom state update request.
type CustomStateUpdateOptions struct {
	Value       interface{} `json:"value"`
	ContentType string      `json:"content_type,omitempty"`
}

// CustomState represents a custom state entry.
type CustomState struct {
	StateID     string      `json:"state_id"`
	AgentID     string      `json:"agent_id"`
	Scope       string      `json:"scope"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ContentType string      `json:"content_type"`
	UserID      string      `json:"user_id,omitempty"`
	InstanceID  string      `json:"instance_id,omitempty"`
	CreatedAt   string      `json:"created_at,omitempty"`
	UpdatedAt   string      `json:"updated_at,omitempty"`
}

// CustomStateListResponse is the response from listing custom states.
type CustomStateListResponse struct {
	States []CustomState `json:"states"`
}

// List returns custom states for an agent.
func (c *CustomStateResource) List(ctx context.Context, agentID string, opts *CustomStateListOptions) (*CustomStateListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Scope != "" {
			params["scope"] = opts.Scope
		}
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.InstanceID != "" {
			params["instance_id"] = opts.InstanceID
		}
	}

	var result CustomStateListResponse
	err := c.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new custom state.
func (c *CustomStateResource) Create(ctx context.Context, agentID string, opts CustomStateCreateOptions) (*CustomState, error) {
	var result CustomState
	err := c.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a custom state.
func (c *CustomStateResource) Update(ctx context.Context, agentID, stateID string, opts CustomStateUpdateOptions) (*CustomState, error) {
	var result CustomState
	err := c.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states/%s", agentID, stateID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a custom state.
func (c *CustomStateResource) Delete(ctx context.Context, agentID, stateID string) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/custom-states/%s", agentID, stateID), nil)
}
