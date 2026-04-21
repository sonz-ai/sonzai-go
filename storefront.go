package sonzai

import (
	"context"
	"fmt"
)

// StorefrontResource provides storefront (agent marketplace) management operations.
type StorefrontResource struct {
	http *httpClient
}

// StorefrontUpdateOptions configures a storefront update request.
type StorefrontUpdateOptions struct {
	DisplayName      string `json:"display_name,omitempty"`
	Description      string `json:"description,omitempty"`
	Slug             string `json:"slug,omitempty"`
	Theme            string `json:"theme,omitempty"`
	HeroImageURL     string `json:"hero_image_url,omitempty"`
	AccessType       string `json:"access_type,omitempty"`
	InviteCode       string `json:"invite_code,omitempty"`
	ContactEmail     string `json:"contact_email,omitempty"`
	MaxVisitsPerUser int    `json:"max_visits_per_user,omitempty"`
}

// StorefrontAgentOptions configures an agent listing on the storefront.
type StorefrontAgentOptions struct {
	DisplayName      string `json:"display_name,omitempty"`
	Description      string `json:"description,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	Slug             string `json:"slug,omitempty"`
	Featured         bool   `json:"featured,omitempty"`
	MaxTurnsPerVisit int    `json:"max_turns_per_visit,omitempty"`
}

// Get returns the tenant's storefront configuration.
func (s *StorefrontResource) Get(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := s.http.Get(ctx, "/api/v1/storefront", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Update updates the tenant's storefront configuration.
func (s *StorefrontResource) Update(ctx context.Context, opts StorefrontUpdateOptions) error {
	return s.http.Put(ctx, "/api/v1/storefront", opts, nil)
}

// ListAgents returns agents listed on the storefront.
func (s *StorefrontResource) ListAgents(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := s.http.Get(ctx, "/api/v1/storefront/agents", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpsertAgent adds or updates an agent listing on the storefront.
func (s *StorefrontResource) UpsertAgent(ctx context.Context, agentID string, opts StorefrontAgentOptions) error {
	return s.http.Put(ctx, fmt.Sprintf("/api/v1/storefront/agents/%s", agentID), opts, nil)
}

// RemoveAgent removes an agent from the storefront.
func (s *StorefrontResource) RemoveAgent(ctx context.Context, agentID string) error {
	return s.http.Delete(ctx, fmt.Sprintf("/api/v1/storefront/agents/%s", agentID), nil)
}

// Publish makes the storefront publicly visible.
func (s *StorefrontResource) Publish(ctx context.Context) error {
	return s.http.Post(ctx, "/api/v1/storefront/publish", nil, nil)
}

// Unpublish hides the storefront from public access.
func (s *StorefrontResource) Unpublish(ctx context.Context) error {
	return s.http.Post(ctx, "/api/v1/storefront/unpublish", nil, nil)
}
