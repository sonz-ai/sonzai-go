package sonzai

import (
	"context"
	"fmt"
)

// InstancesResource provides agent instance operations.
type InstancesResource struct {
	http *httpClient
}

// List returns all instances for an agent.
func (i *InstancesResource) List(ctx context.Context, agentID string) (*InstanceListResponse, error) {
	var result InstanceListResponse
	err := i.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/instances", agentID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new agent instance.
func (i *InstancesResource) Create(ctx context.Context, agentID, name, description string) (*AgentInstance, error) {
	body := map[string]string{"name": name}
	if description != "" {
		body["description"] = description
	}

	var result AgentInstance
	err := i.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/instances", agentID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a specific instance.
func (i *InstancesResource) Get(ctx context.Context, agentID, instanceID string) (*AgentInstance, error) {
	var result AgentInstance
	err := i.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/instances/%s", agentID, instanceID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an instance.
func (i *InstancesResource) Delete(ctx context.Context, agentID, instanceID string) error {
	return i.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/instances/%s", agentID, instanceID), nil)
}

// Reset resets an instance (clears all context data).
func (i *InstancesResource) Reset(ctx context.Context, agentID, instanceID string) (*AgentInstance, error) {
	var result AgentInstance
	err := i.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/instances/%s/reset", agentID, instanceID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing agent instance.
func (i *InstancesResource) Update(ctx context.Context, agentID, instanceID string, opts UpdateInstanceOptions) (*AgentInstance, error) {
	var result AgentInstance
	err := i.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/instances/%s", agentID, instanceID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
