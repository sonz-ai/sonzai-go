package sonzai

import (
	"context"
	"fmt"
)

// BYOKProvider is the typed enum of supported BYOK providers.
type BYOKProvider string

const (
	BYOKProviderOpenAI     BYOKProvider = "openai"
	BYOKProviderGemini     BYOKProvider = "gemini"
	BYOKProviderXAI        BYOKProvider = "xai"
	BYOKProviderOpenRouter BYOKProvider = "openrouter"
)

// BYOKKeyResponse is the metadata returned for a stored BYOK key. Key material
// is never returned by the API — only the prefix and health metadata.
type BYOKKeyResponse struct {
	Provider          string  `json:"provider"`
	APIKeyPrefix      string  `json:"api_key_prefix"`
	IsActive          bool    `json:"is_active"`
	HealthStatus      string  `json:"health_status"`
	LastHealthError   *string `json:"last_health_error,omitempty"`
	LastHealthCheckAt *string `json:"last_health_check_at,omitempty"`
	LastUsedAt        *string `json:"last_used_at,omitempty"`
	UpdatedAt         string  `json:"updated_at"`
}

// BYOKResource provides project-scoped bring-your-own-key (BYOK) configuration.
type BYOKResource struct {
	http *httpClient
}

// List returns all configured BYOK providers for a project.
func (c *BYOKResource) List(ctx context.Context, projectID string) ([]BYOKKeyResponse, error) {
	var result struct {
		Keys []BYOKKeyResponse `json:"keys"`
	}
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/byok-keys", projectID), nil, &result); err != nil {
		return nil, err
	}
	return result.Keys, nil
}

// Set creates or replaces the BYOK key for a provider. The API validates the
// key against the provider before storage; an invalid key returns a 400.
func (c *BYOKResource) Set(ctx context.Context, projectID string, provider BYOKProvider, apiKey string) (*BYOKKeyResponse, error) {
	body := map[string]string{"api_key": apiKey}
	var result BYOKKeyResponse
	if err := c.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/byok-keys/%s", projectID, provider), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes the BYOK key for a provider. Subsequent calls fall back to
// the platform's standard billing.
func (c *BYOKResource) Delete(ctx context.Context, projectID string, provider BYOKProvider) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/byok-keys/%s", projectID, provider), nil)
}

// SetActive enables or disables a stored BYOK key without rotating it.
func (c *BYOKResource) SetActive(ctx context.Context, projectID string, provider BYOKProvider, isActive bool) (*BYOKKeyResponse, error) {
	body := map[string]bool{"is_active": isActive}
	var result BYOKKeyResponse
	if err := c.http.Patch(ctx, fmt.Sprintf("/api/v1/projects/%s/byok-keys/%s", projectID, provider), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Test re-tests a stored BYOK key against the provider and updates
// health_status with the result.
func (c *BYOKResource) Test(ctx context.Context, projectID string, provider BYOKProvider) (*BYOKKeyResponse, error) {
	var result BYOKKeyResponse
	if err := c.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/byok-keys/%s/test", projectID, provider), struct{}{}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
