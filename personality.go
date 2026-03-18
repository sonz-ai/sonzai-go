package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// PersonalityResource provides personality operations for an agent.
type PersonalityResource struct {
	http *httpClient
}

// Get returns the personality profile and evolution history.
func (p *PersonalityResource) Get(ctx context.Context, agentID string, opts *PersonalityGetOptions) (*PersonalityResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.HistoryLimit > 0 {
			params["history_limit"] = strconv.Itoa(opts.HistoryLimit)
		}
		if opts.Since != "" {
			params["since"] = opts.Since
		}
	}

	var result PersonalityResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates the personality Big5 scores for an agent.
func (p *PersonalityResource) Update(ctx context.Context, agentID string, opts PersonalityUpdateOptions) (*PersonalityUpdateResponse, error) {
	var result PersonalityUpdateResponse
	err := p.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/personality", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchGet returns personality profiles for multiple agents.
func (p *PersonalityResource) BatchGet(ctx context.Context, agentIDs []string) (*BatchPersonalityResponse, error) {
	var result BatchPersonalityResponse
	err := p.http.Post(ctx, "/api/v1/agents/personalities/batch", map[string]interface{}{"agent_ids": agentIDs}, &result)
	return &result, err
}

// GetSignificantMoments returns significant moments for an agent.
func (p *PersonalityResource) GetSignificantMoments(ctx context.Context, agentID string, limit int) (*SignificantMomentsResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	var result SignificantMomentsResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality/significant-moments", agentID), params, &result)
	return &result, err
}

// GetRecentShifts returns recent personality shifts for an agent.
func (p *PersonalityResource) GetRecentShifts(ctx context.Context, agentID string) (*RecentShiftsResponse, error) {
	var result RecentShiftsResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality/recent-shifts", agentID), nil, &result)
	return &result, err
}

// ListUserOverlays returns all user-specific personality overlays for an agent.
func (p *PersonalityResource) ListUserOverlays(ctx context.Context, agentID string) (*UserOverlaysListResponse, error) {
	var result UserOverlaysListResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality/users", agentID), nil, &result)
	return &result, err
}

// GetUserOverlay returns the personality overlay for a specific user.
func (p *PersonalityResource) GetUserOverlay(ctx context.Context, agentID, userID string, opts UserOverlayOptions) (*UserOverlayDetailResponse, error) {
	params := map[string]string{}
	if opts.InstanceID != "" {
		params["instance_id"] = opts.InstanceID
	}
	if opts.Since != "" {
		params["since"] = opts.Since
	}
	var result UserOverlayDetailResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/personality/users/%s", agentID, userID), params, &result)
	return &result, err
}
