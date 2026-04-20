package sonzai

import (
	"context"
	"fmt"
)

// WakeupResource provides wakeup scheduling operations for an agent.
type WakeupResource struct {
	http *httpClient
}

// ScheduleWakeupOptions configures a wakeup scheduling request.
type ScheduleWakeupOptions struct {
	UserID           string `json:"user_id"`
	ScheduledAt      string `json:"scheduled_at"` // RFC3339 timestamp
	CheckType        string `json:"check_type"`   // "birthday", "occasion", "recurring_event", "interest_check", "general"
	Intent           string `json:"intent,omitempty"`
	Occasion         string `json:"occasion,omitempty"`
	InterestTopic    string `json:"interest_topic,omitempty"`
	EventDescription string `json:"event_description,omitempty"`
}

// ScheduledWakeup represents a scheduled wakeup.
type ScheduledWakeup struct {
	WakeupID         string `json:"wakeup_id"`
	AgentID          string `json:"agent_id"`
	UserID           string `json:"user_id"`
	ScheduledAt      string `json:"scheduled_at"`
	CheckType        string `json:"check_type"`
	Status           string `json:"status"`
	Intent           string `json:"intent,omitempty"`
	LastTopic        string `json:"last_topic,omitempty"`
	EventDescription string `json:"event_description,omitempty"`
	Occasion         string `json:"occasion,omitempty"`
	InterestTopic    string `json:"interest_topic,omitempty"`
	ResearchSummary  string `json:"research_summary,omitempty"`
	ExecutedAt       string `json:"executed_at,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

// List returns scheduled wakeups for the agent.
func (w *WakeupResource) List(ctx context.Context, agentID string, opts *WakeupListOptions) (*WakeupsResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.UserID != "" {
			params["user_id"] = opts.UserID
		}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
	}
	var result WakeupsResponse
	err := w.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/wakeups", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Schedule creates a new scheduled wakeup for the agent.
func (w *WakeupResource) Schedule(ctx context.Context, agentID string, opts ScheduleWakeupOptions) (*ScheduledWakeup, error) {
	var result ScheduledWakeup
	err := w.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/wakeups", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

