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
