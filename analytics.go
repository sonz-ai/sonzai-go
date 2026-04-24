package sonzai

import (
	"context"
	"strconv"
)

// AnalyticsResource provides platform analytics endpoints.
type AnalyticsResource struct {
	http *httpClient
}

// AnalyticsOptions configures analytics query time range.
// Use Days for a lookback window (1-365). Use Month ("YYYY-MM") or
// Start+End ("YYYY-MM-DD") for a specific range.
type AnalyticsOptions struct {
	Days      int
	Month     string
	Start     string
	End       string
	ProjectID string
	// AgentID scopes the result to a single agent. Only honored by endpoints
	// that accept per-agent filtering (e.g. Composio usage).
	AgentID string
}

// ComposioAppUsage is per-app tool-call volume and cost within the queried window.
type ComposioAppUsage struct {
	App     string  `json:"app"`
	Calls   int64   `json:"calls"`
	CostUSD float64 `json:"cost_usd"`
}

// ComposioUsagePeriod is the actual window returned by the server (after
// defaulting to the last 30 days when Start/End are omitted).
type ComposioUsagePeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ComposioUsageSummary is aggregate totals across every app in the window.
type ComposioUsageSummary struct {
	TotalCalls   int64   `json:"total_calls"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// ComposioUsageResponse is the payload returned from /analytics/composio.
type ComposioUsageResponse struct {
	ByApp   []ComposioAppUsage   `json:"by_app"`
	Period  ComposioUsagePeriod  `json:"period"`
	Summary ComposioUsageSummary `json:"summary"`
}

func (o *AnalyticsOptions) toParams() map[string]string {
	if o == nil {
		return nil
	}
	p := map[string]string{}
	if o.Days > 0 {
		p["days"] = strconv.Itoa(o.Days)
	}
	if o.Month != "" {
		p["month"] = o.Month
	}
	if o.Start != "" {
		p["start"] = o.Start
	}
	if o.End != "" {
		p["end"] = o.End
	}
	if o.ProjectID != "" {
		p["project_id"] = o.ProjectID
	}
	if o.AgentID != "" {
		p["agent_id"] = o.AgentID
	}
	return p
}

// Overview returns high-level platform analytics.
func (a *AnalyticsResource) Overview(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/overview", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Usage returns token and request usage metrics.
func (a *AnalyticsResource) Usage(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/usage", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Cost returns cost summary metrics.
func (a *AnalyticsResource) Cost(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/cost", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CostBreakdown returns cost broken down by operation, model, and agent.
func (a *AnalyticsResource) CostBreakdown(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/cost/breakdown", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Realtime returns real-time platform metrics.
func (a *AnalyticsResource) Realtime(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/realtime", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ComposioUsage returns Composio tool-call counts and costs grouped by app
// (gmail, slack, github, …) over the requested window. Start/End accept
// "YYYY-MM-DD"; when both are empty the server defaults to the last 30 days.
// Setting AgentID scopes the result to a single agent.
func (a *AnalyticsResource) ComposioUsage(ctx context.Context, opts *AnalyticsOptions) (*ComposioUsageResponse, error) {
	var result ComposioUsageResponse
	if err := a.http.Get(ctx, "/api/v1/analytics/composio", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
