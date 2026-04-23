package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// ComposioResource provides per-agent Composio connected-account
// operations (Gmail, Calendar, Slack, GitHub, Linear, Notion, Drive).
// Gated on the Composio agent capability server-side. Deployments
// without COMPOSIO_API_KEY return 503 from every endpoint; the SDK
// surface stays stable so callers can detect "not configured" cleanly.
type ComposioResource struct {
	http *httpClient
}

// ComposioConnection is one connected Composio account for an agent.
type ComposioConnection struct {
	AgentID            string    `json:"agent_id"`
	App                string    `json:"app"`
	Scope              string    `json:"scope,omitempty"`
	ConnectedAccountID string    `json:"connected_account_id"`
	AccountLabel       string    `json:"account_label,omitempty"`
	ConnectedByUserID  string    `json:"connected_by_user_id,omitempty"`
	ConnectedAt        time.Time `json:"connected_at"`
	LastVerifiedAt     time.Time `json:"last_verified_at,omitempty"`
}

// ListComposioConnectionsResponse is returned by ListConnections.
type ListComposioConnectionsResponse struct {
	Connections []ComposioConnection `json:"connections"`
}

// InitiateComposioConnectInput is the POST body for starting OAuth.
type InitiateComposioConnectInput struct {
	App string `json:"app"`
}

// InitiateComposioConnectResponse carries the redirect URL the end user
// must visit plus the pending connected_account_id that InitiateConnect
// then finalizes via ConnectCallback.
type InitiateComposioConnectResponse struct {
	RedirectURL        string `json:"redirect_url"`
	ConnectedAccountID string `json:"connected_account_id"`
}

// ComposioConnectCallbackInput finalizes an OAuth flow.
type ComposioConnectCallbackInput struct {
	App                string `json:"app"`
	ConnectedAccountID string `json:"connected_account_id"`
	AccountLabel       string `json:"account_label,omitempty"`
	ConnectedByUserID  string `json:"connected_by_user_id,omitempty"`
}

// ComposioConnectCallbackResponse echoes the persisted row.
type ComposioConnectCallbackResponse struct {
	OK         bool                `json:"ok"`
	Connection *ComposioConnection `json:"connection,omitempty"`
}

// ComposioActionLogEntry is one row of the action audit log.
type ComposioActionLogEntry struct {
	ID              string    `json:"id"`
	AgentID         string    `json:"agent_id"`
	UserID          string    `json:"user_id,omitempty"`
	TurnID          string    `json:"turn_id,omitempty"`
	App             string    `json:"app"`
	Action          string    `json:"action"`
	Status          string    `json:"status"` // ok | error | rate_limited
	LatencyMs       int64     `json:"latency_ms"`
	RequestSummary  string    `json:"request_summary,omitempty"`
	ResponseSummary string    `json:"response_summary,omitempty"`
	ErrorCode       string    `json:"error_code,omitempty"`
	RecordedAt      time.Time `json:"recorded_at"`
}

// ListComposioAuditResponse is returned by ListAudit.
type ListComposioAuditResponse struct {
	Entries []ComposioActionLogEntry `json:"entries"`
}

// ComposioAvailableAction is the per-action shape the LLM sees.
type ComposioAvailableAction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ComposioAvailableActionsApp groups actions for one connected app.
type ComposioAvailableActionsApp struct {
	App                   string                    `json:"app"`
	ConnectedAccountLabel string                    `json:"connected_account_label,omitempty"`
	Actions               []ComposioAvailableAction `json:"actions"`
}

// ListComposioAvailableActionsResponse is returned by ListAvailableActions.
type ListComposioAvailableActionsResponse struct {
	Apps []ComposioAvailableActionsApp `json:"apps"`
}

// ListComposioAuditOptions configures ListAudit.
type ListComposioAuditOptions struct {
	From   time.Time
	To     time.Time
	Status string // "ok" | "error" | "rate_limited"
	Limit  int
}

// ListConnections returns every Composio connected account for an agent.
func (r *ComposioResource) ListConnections(ctx context.Context, agentID string) (*ListComposioConnectionsResponse, error) {
	var out ListComposioConnectionsResponse
	path := fmt.Sprintf("/api/v1/agents/%s/composio/connections", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// InitiateConnect starts a Composio OAuth flow. Returns the redirect URL
// the end user must visit and the pending connected_account_id that
// ConnectCallback will finalize.
func (r *ComposioResource) InitiateConnect(ctx context.Context, agentID string, input InitiateComposioConnectInput) (*InitiateComposioConnectResponse, error) {
	var out InitiateComposioConnectResponse
	path := fmt.Sprintf("/api/v1/agents/%s/composio/connections", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConnectCallback persists a connection after the end user finishes OAuth.
func (r *ComposioResource) ConnectCallback(ctx context.Context, agentID string, input ComposioConnectCallbackInput) (*ComposioConnectCallbackResponse, error) {
	var out ComposioConnectCallbackResponse
	path := fmt.Sprintf("/api/v1/agents/%s/composio/connections/callback", url.PathEscape(agentID))
	if err := r.http.Post(ctx, path, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteConnection revokes a connection for (agent, app).
func (r *ComposioResource) DeleteConnection(ctx context.Context, agentID, app string) error {
	path := fmt.Sprintf("/api/v1/agents/%s/composio/connections/%s", url.PathEscape(agentID), url.PathEscape(app))
	return r.http.Delete(ctx, path, nil)
}

// ListAudit returns redacted Composio action log entries for an agent.
func (r *ComposioResource) ListAudit(ctx context.Context, agentID string, opts *ListComposioAuditOptions) (*ListComposioAuditResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if !opts.From.IsZero() {
			params["from"] = opts.From.UTC().Format(time.RFC3339)
		}
		if !opts.To.IsZero() {
			params["to"] = opts.To.UTC().Format(time.RFC3339)
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}
	var out ListComposioAuditResponse
	path := fmt.Sprintf("/api/v1/agents/%s/composio/audit", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAvailableActions returns the curated Composio action set available
// to the agent, grouped by connected app. The dashboard uses this to
// render "Gmail — 4 enabled actions"; the LLM sees a subset as declared
// tools during chat turns.
func (r *ComposioResource) ListAvailableActions(ctx context.Context, agentID string) (*ListComposioAvailableActionsResponse, error) {
	var out ListComposioAvailableActionsResponse
	path := fmt.Sprintf("/api/v1/agents/%s/composio/available_actions", url.PathEscape(agentID))
	if err := r.http.Get(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
