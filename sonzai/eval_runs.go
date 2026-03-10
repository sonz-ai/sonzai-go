package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// EvalRunsResource provides eval run operations.
type EvalRunsResource struct {
	http *httpClient
}

func newEvalRunsResource(http *httpClient) *EvalRunsResource {
	return &EvalRunsResource{http: http}
}

// List returns eval runs.
func (e *EvalRunsResource) List(ctx context.Context, agentID string, limit, offset int) (*EvalRunListResponse, error) {
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

	var result EvalRunListResponse
	err := e.http.get(ctx, "/api/v1/eval-runs", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a specific eval run.
func (e *EvalRunsResource) Get(ctx context.Context, runID string) (*EvalRun, error) {
	var result EvalRun
	err := e.http.get(ctx, fmt.Sprintf("/api/v1/eval-runs/%s", runID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an eval run.
func (e *EvalRunsResource) Delete(ctx context.Context, runID string) error {
	return e.http.del(ctx, fmt.Sprintf("/api/v1/eval-runs/%s", runID), nil)
}
