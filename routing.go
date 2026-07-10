package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// RoutingResource manages project routing configuration and permanent routes.
type RoutingResource struct {
	http *httpClient
}

type RoutingTier struct {
	Name      string          `json:"name"`
	Label     string          `json:"label"`
	Criteria  json.RawMessage `json:"criteria"`
	AgentID   string          `json:"agent_id"`
	AgentName string          `json:"agent_name"`
	Prompt    string          `json:"prompt"`
}

type RoutingGuideAgent struct {
	AgentID   string   `json:"agent_id"`
	AgentName string   `json:"agent_name"`
	Criteria  []string `json:"criteria"`
	Questions []string `json:"questions"`
}

type RoutingHandoff struct {
	From  string   `json:"from"`
	To    string   `json:"to"`
	When  string   `json:"when"`
	Mode  string   `json:"mode"`
	Carry []string `json:"carry"`
}

type RoutingChannelBinding struct {
	ConnectionID string `json:"connection_id,omitempty"`
	ChannelType  string `json:"channel_type,omitempty"`
	AgentID      string `json:"agent_id"`
	UseGuide     bool   `json:"use_guide"`
}

type RoutingConfig struct {
	Tiers           []RoutingTier           `json:"tiers"`
	GuideAgent      RoutingGuideAgent       `json:"guide_agent"`
	Handoffs        []RoutingHandoff        `json:"handoffs"`
	ChannelBindings []RoutingChannelBinding `json:"channel_bindings"`
}

type PermanentRoute struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	UserID          string    `json:"user_id"`
	ContactName     string    `json:"contact_name"`
	Tier            string    `json:"tier"`
	AgentID         string    `json:"agent_id"`
	Overridden      bool      `json:"overridden"`
	OverrideAgentID string    `json:"override_agent_id"`
	ClassifiedAt    time.Time `json:"classified_at"`
}

type PermanentRouteList struct {
	Routes []PermanentRoute `json:"routes"`
	Total  int              `json:"total"`
}

type ClassifyContactOptions struct {
	UserID      string `json:"user_id"`
	ContactName string `json:"contact_name"`
	Tier        string `json:"tier"`
	AgentID     string `json:"agent_id"`
}

func (r *RoutingResource) GetConfig(ctx context.Context, projectID string) (*RoutingConfig, error) {
	var out RoutingConfig
	if err := r.http.Get(ctx, routingConfigPath(projectID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RoutingResource) PutConfig(ctx context.Context, projectID string, cfg RoutingConfig) (*RoutingConfig, error) {
	var out RoutingConfig
	if err := r.http.Put(ctx, routingConfigPath(projectID), cfg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RoutingResource) ListPermanentRoutes(ctx context.Context, projectID string) (*PermanentRouteList, error) {
	var out PermanentRouteList
	if err := r.http.Get(ctx, permanentRoutesPath(projectID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RoutingResource) ClassifyContact(ctx context.Context, projectID string, opts ClassifyContactOptions) (*PermanentRoute, error) {
	var out PermanentRoute
	if err := r.http.Post(ctx, permanentRoutesPath(projectID), opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *RoutingResource) OverridePermanentRoute(ctx context.Context, projectID, userID, agentID string) (*PermanentRoute, error) {
	path := fmt.Sprintf("%s/%s/override", permanentRoutesPath(projectID), url.PathEscape(userID))
	var out PermanentRoute
	if err := r.http.Post(ctx, path, map[string]string{"agent_id": agentID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func routingConfigPath(projectID string) string {
	return fmt.Sprintf("/api/v1/projects/%s/routing-config", url.PathEscape(projectID))
}

func permanentRoutesPath(projectID string) string {
	return fmt.Sprintf("/api/v1/projects/%s/permanent-routes", url.PathEscape(projectID))
}
