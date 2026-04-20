package sonzai

import (
	"context"
	"fmt"
)

// ProjectConfigResource provides project-scoped configuration operations.
type ProjectConfigResource struct {
	http *httpClient
}

// ProjectConfigEntry represents a single config key-value pair.
type ProjectConfigEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt string      `json:"updated_at,omitempty"`
}

// ProjectConfigListResponse is the response from listing project configs.
type ProjectConfigListResponse struct {
	Configs []ProjectConfigEntry `json:"configs"`
}

// List returns all config entries for a project.
func (c *ProjectConfigResource) List(ctx context.Context, projectID string) (*ProjectConfigListResponse, error) {
	var result ProjectConfigListResponse
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/config", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a config value by key.
func (c *ProjectConfigResource) Get(ctx context.Context, projectID, key string) (*ProjectConfigEntry, error) {
	var result ProjectConfigEntry
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/config/%s", projectID, key), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Set creates or updates a config value. The value must be valid JSON.
func (c *ProjectConfigResource) Set(ctx context.Context, projectID, key string, value interface{}) error {
	var result struct {
		Success bool `json:"success"`
	}
	return c.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/config/%s", projectID, key), value, &result)
}

// Delete removes a config entry.
func (c *ProjectConfigResource) Delete(ctx context.Context, projectID, key string) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/config/%s", projectID, key), nil)
}
