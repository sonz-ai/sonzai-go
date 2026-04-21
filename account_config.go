package sonzai

import (
	"context"
	"fmt"
)

// AccountConfigResource provides tenant-scoped ("account-level") configuration
// operations. The tenant is resolved from the caller's API key or Clerk
// session on the server — never from a URL parameter — so callers can only
// read or write config for the tenant they are currently authenticated to.
//
// Account configuration lives in a JSONB KV store backed by
// platform.account_config in CockroachDB. Use it for settings that should
// apply to every project inside the tenant without per-project duplication:
// for example, the default post-processing model map (see
// PostProcessingModelMapKey).
type AccountConfigResource struct {
	http *httpClient
}

// AccountConfigEntry is one key/value pair from the tenant's config store.
// Shape mirrors ProjectConfigEntry deliberately so callers that already
// handle project config can reuse their serialisation logic.
type AccountConfigEntry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt string      `json:"updated_at,omitempty"`
}

// AccountConfigListResponse is the response from listing account configs.
type AccountConfigListResponse struct {
	Configs []AccountConfigEntry `json:"configs"`
}

// List returns all config entries for the authenticated tenant.
func (c *AccountConfigResource) List(ctx context.Context) (*AccountConfigListResponse, error) {
	var result AccountConfigListResponse
	if err := c.http.Get(ctx, "/api/v1/account/config", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single config value by key.
func (c *AccountConfigResource) Get(ctx context.Context, key string) (*AccountConfigEntry, error) {
	var result AccountConfigEntry
	if err := c.http.Get(ctx, fmt.Sprintf("/api/v1/account/config/%s", key), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Set creates or updates a config value. The value is serialised as JSON.
func (c *AccountConfigResource) Set(ctx context.Context, key string, value interface{}) error {
	var result struct {
		Success bool `json:"success"`
	}
	return c.http.Put(ctx, fmt.Sprintf("/api/v1/account/config/%s", key), value, &result)
}

// Delete removes a config entry.
func (c *AccountConfigResource) Delete(ctx context.Context, key string) error {
	return c.http.Delete(ctx, fmt.Sprintf("/api/v1/account/config/%s", key), nil)
}
