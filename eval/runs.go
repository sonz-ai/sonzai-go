package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// RunsResource provides eval run operations.
type RunsResource struct {
	backend Backend
}

// List returns eval runs, optionally filtered by agent.
func (r *RunsResource) List(ctx context.Context, agentID string, limit, offset int) (*RunListResponse, error) {
	params := map[string]string{}
	if agentID != "" {
		params["agent_id"] = agentID
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if offset > 0 {
		params["offset"] = strconv.Itoa(offset)
	}

	var result RunListResponse
	err := r.backend.Get(ctx, "/api/v1/eval-runs", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a specific eval run.
func (r *RunsResource) Get(ctx context.Context, runID string) (*Run, error) {
	var result Run
	err := r.backend.Get(ctx, fmt.Sprintf("/api/v1/eval-runs/%s", runID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an eval run.
func (r *RunsResource) Delete(ctx context.Context, runID string) error {
	return r.backend.Delete(ctx, fmt.Sprintf("/api/v1/eval-runs/%s", runID), nil)
}

// StreamEvents connects to a running eval's SSE event stream and calls
// the callback for each event. Supports reconnection via fromIndex.
func (r *RunsResource) StreamEvents(ctx context.Context, runID string, fromIndex int, callback func(SimulationEvent) error) error {
	path := fmt.Sprintf("/api/v1/eval-runs/%s/events?from=%d", runID, fromIndex)
	return r.backend.StreamSSE(ctx, "GET", path, nil, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil // skip malformed events
		}
		return callback(event)
	})
}
