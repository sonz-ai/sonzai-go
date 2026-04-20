package sonzai

import (
	"context"
	"fmt"
)

// CustomLLMResource provides project-scoped custom LLM configuration operations.
type CustomLLMResource struct {
	http *httpClient
}

// CustomLLMConfigResponse is the response from getting the custom LLM config.
type CustomLLMConfigResponse struct {
	Endpoint     string `json:"endpoint"`
	APIKeyPrefix string `json:"api_key_prefix"`
	Model        string `json:"model"`
	DisplayName  string `json:"display_name"`
	IsActive     bool   `json:"is_active"`
	Configured   bool   `json:"configured"`
}

// SetCustomLLMOptions configures a custom LLM provider.
type SetCustomLLMOptions struct {
	Endpoint    string `json:"endpoint"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// Get returns the custom LLM config for a project.
func (c *CustomLLMResource) Get(ctx context.Context, projectID string) (*CustomLLMConfigResponse, error) {
	var result CustomLLMConfigResponse
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/custom-llm", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Set creates or updates the custom LLM config for a project.
func (c *CustomLLMResource) Set(ctx context.Context, projectID string, opts SetCustomLLMOptions) (*CustomLLMConfigResponse, error) {
	var result CustomLLMConfigResponse
	if err := c.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/custom-llm", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes the custom LLM config for a project.
func (c *CustomLLMResource) Delete(ctx context.Context, projectID string) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/custom-llm", projectID), nil)
}
