package sonzai

import (
	"context"
	"fmt"
)

// APIKeysResource provides project API key management operations.
type APIKeysResource struct {
	http *httpClient
}

// APIKey represents a project API key.
type APIKey struct {
	KeyID     string   `json:"key_id"`
	Key       string   `json:"key,omitempty"`
	KeyPrefix string   `json:"key_prefix,omitempty"`
	Name      string   `json:"name,omitempty"`
	ProjectID string   `json:"project_id"`
	TenantID  string   `json:"tenant_id,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
	IsActive  bool     `json:"is_active"`
	CreatedBy string   `json:"created_by,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

// APIKeyListResponse is the response from listing API keys.
type APIKeyListResponse struct {
	Keys []APIKey `json:"keys"`
}

// CreateAPIKeyOptions configures an API key creation request.
type CreateAPIKeyOptions struct {
	Name        string   `json:"name,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	ExpiresDays int      `json:"expires_days,omitempty"`
}

// List returns all API keys for a project.
func (a *APIKeysResource) List(ctx context.Context, projectID string) (*APIKeyListResponse, error) {
	var result APIKeyListResponse
	if err := a.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/keys", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new API key for a project.
// The full key value is only returned once in the response — store it securely.
func (a *APIKeysResource) Create(ctx context.Context, projectID string, opts CreateAPIKeyOptions) (*APIKey, error) {
	var result APIKey
	if err := a.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/keys", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Revoke permanently revokes an API key.
func (a *APIKeysResource) Revoke(ctx context.Context, projectID, keyID string) error {
	return a.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/keys/%s", projectID, keyID), nil)
}
