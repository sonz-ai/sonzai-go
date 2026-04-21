package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// SchedulesResource provides recurring per-user schedule operations.
//
// The platform computes NextFireAt from the supplied cadence
// ({simple:{...}} or {cron:"..."}) and honors any ActiveWindow filter
// for quiet hours / days-of-week.
type SchedulesResource struct {
	http *httpClient
}

// Schedule represents a recurring per-user schedule.
//
// Cadence, ActiveWindow, and Metadata are stored as JSON strings on the
// server; callers building request bodies should pass map/struct values
// and the SDK will marshal them. Responses return these fields as raw
// JSON strings which callers can unmarshal as needed.
type Schedule struct {
	ScheduleID      string `json:"schedule_id"`
	CadenceType     string `json:"cadence_type"`
	Cadence         string `json:"cadence"`
	ActiveWindow    string `json:"active_window,omitempty"`
	Timezone        string `json:"timezone"`
	NextFireAt      string `json:"next_fire_at"`
	InventoryItemID string `json:"inventory_item_id,omitempty"`
	Intent          string `json:"intent"`
	CheckType       string `json:"check_type"`
	Metadata        string `json:"metadata,omitempty"`
	Enabled         bool   `json:"enabled"`
	StartsAt        string `json:"starts_at,omitempty"`
	EndsAt          string `json:"ends_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ScheduleListResponse is the response from listing schedules for a
// (agent, user) pair.
type ScheduleListResponse struct {
	Schedules []Schedule `json:"schedules"`
}

// CreateScheduleOptions configures a schedule creation request.
//
// Cadence accepts {"simple": {...}} or {"cron": "..."} with a required
// timezone field. ActiveWindow is an optional quiet-hours/days filter.
// Metadata is free-form JSON kept as-is.
type CreateScheduleOptions struct {
	Cadence         json.RawMessage `json:"cadence"`
	ActiveWindow    json.RawMessage `json:"active_window,omitempty"`
	Intent          string          `json:"intent"`
	CheckType       string          `json:"check_type"`
	InventoryItemID string          `json:"inventory_item_id,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
	StartsAt        string          `json:"starts_at,omitempty"` // RFC3339
	EndsAt          string          `json:"ends_at,omitempty"`   // RFC3339
}

// CreateScheduleResponse is the response from creating a schedule.
type CreateScheduleResponse struct {
	ScheduleID      string `json:"schedule_id"`
	NextFireAt      string `json:"next_fire_at"`
	NextFireAtLocal string `json:"next_fire_at_local"`
	Enabled         bool   `json:"enabled"`
}

// UpdateScheduleOptions configures a partial schedule update.
//
// All fields are optional. NextFireAt is recomputed only when
// Cadence, ActiveWindow, or StartsAt change.
type UpdateScheduleOptions struct {
	Cadence      json.RawMessage `json:"cadence,omitempty"`
	ActiveWindow json.RawMessage `json:"active_window,omitempty"`
	Intent       *string         `json:"intent,omitempty"`
	CheckType    *string         `json:"check_type,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	Enabled      *bool           `json:"enabled,omitempty"`
	StartsAt     *string         `json:"starts_at,omitempty"`
	EndsAt       *string         `json:"ends_at,omitempty"`
}

// UpcomingScheduleResponse is the response from previewing upcoming fire times.
type UpcomingScheduleResponse struct {
	Upcoming []string `json:"upcoming"` // RFC3339 UTC timestamps
}

// List returns all schedules for a (agent, user) pair.
func (s *SchedulesResource) List(ctx context.Context, agentID, userID string) (*ScheduleListResponse, error) {
	var result ScheduleListResponse
	err := s.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules", agentID, userID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a recurring schedule for a user.
func (s *SchedulesResource) Create(ctx context.Context, agentID, userID string, opts CreateScheduleOptions) (*CreateScheduleResponse, error) {
	var result CreateScheduleResponse
	err := s.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules", agentID, userID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get fetches a single schedule by ID.
func (s *SchedulesResource) Get(ctx context.Context, agentID, userID, scheduleID string) (*Schedule, error) {
	var result Schedule
	err := s.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules/%s", agentID, userID, scheduleID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update partially updates a schedule. NextFireAt is recomputed only when
// Cadence, ActiveWindow, or StartsAt change.
func (s *SchedulesResource) Update(ctx context.Context, agentID, userID, scheduleID string, opts UpdateScheduleOptions) (*Schedule, error) {
	var result Schedule
	err := s.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules/%s", agentID, userID, scheduleID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a schedule. Idempotent — missing IDs return 204.
func (s *SchedulesResource) Delete(ctx context.Context, agentID, userID, scheduleID string) error {
	return s.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules/%s", agentID, userID, scheduleID), nil)
}

// Upcoming previews the next N allowed fire times. Does not mutate state.
// When limit is 0, the server default (10) is used; the server caps limit at 100.
func (s *SchedulesResource) Upcoming(ctx context.Context, agentID, userID, scheduleID string, limit int) (*UpcomingScheduleResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result UpcomingScheduleResponse
	err := s.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/schedules/%s/upcoming", agentID, userID, scheduleID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
