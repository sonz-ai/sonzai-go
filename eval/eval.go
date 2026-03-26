// Package eval provides evaluation, simulation, and benchmarking operations
// for the Sonzai Character Engine API.
//
// This package is a distinct domain from the core chat/memory/personality
// surface. It provides tools for testing and measuring agent quality.
//
// # Getting Started
//
// The eval client is accessed through the root sonzai.Client:
//
//	client := sonzai.NewClient("sk-your-api-key")
//
//	// Run evaluation
//	result, err := client.Eval.Evaluate(ctx, "agent-id", eval.EvaluateOptions{
//	    Messages:   []eval.Message{{Role: "user", Content: "Hello"}},
//	    TemplateID: "template-id",
//	})
//
//	// Manage templates
//	templates, err := client.Eval.Templates.List(ctx, "")
package eval

import (
	"context"
	"encoding/json"
	"fmt"
)

// Backend is the HTTP transport interface used by eval resources.
// It is implemented by the internal httpClient type in the root sonzai package.
type Backend interface {
	Get(ctx context.Context, path string, params map[string]string, result interface{}) error
	Post(ctx context.Context, path string, body interface{}, result interface{}) error
	Put(ctx context.Context, path string, body interface{}, result interface{}) error
	Patch(ctx context.Context, path string, body interface{}, result interface{}) error
	Delete(ctx context.Context, path string, result interface{}) error
	StreamSSE(ctx context.Context, method, path string, body interface{}, callback func(json.RawMessage) error) error
}

// Client provides evaluation and simulation operations.
type Client struct {
	backend   Backend
	Templates *TemplatesResource
	Runs      *RunsResource
}

// New creates a new eval client with the given backend.
func New(backend Backend) *Client {
	return &Client{
		backend:   backend,
		Templates: &TemplatesResource{backend: backend},
		Runs:      &RunsResource{backend: backend},
	}
}

// Evaluate evaluates an agent's responses against a template.
func (c *Client) Evaluate(ctx context.Context, agentID string, opts EvaluateOptions) (*EvaluationResult, error) {
	body := map[string]interface{}{
		"messages":    opts.Messages,
		"template_id": opts.TemplateID,
	}
	if opts.ConfigOverride != nil {
		body["config_override"] = opts.ConfigOverride
	}

	var result EvaluationResult
	if err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/evaluate", agentID), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Simulate launches a multi-turn simulation, streams events via callback.
// Returns a RunRef so the caller can reconnect or poll if the stream drops.
func (c *Client) Simulate(ctx context.Context, agentID string, opts SimulateOptions, callback func(SimulationEvent) error) (*RunRef, error) {
	var ref RunRef
	if err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/simulate", agentID), opts, &ref); err != nil {
		return nil, err
	}
	if err := c.Runs.StreamEvents(ctx, ref.RunID, 0, callback); err != nil {
		return &ref, err
	}
	return &ref, nil
}

// SimulateAsync launches a simulation without streaming. Returns a RunRef for polling.
func (c *Client) SimulateAsync(ctx context.Context, agentID string, opts SimulateOptions) (*RunRef, error) {
	var ref RunRef
	err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/simulate", agentID), opts, &ref)
	return &ref, err
}

// Run launches simulation + evaluation combined, streams events via callback.
// Returns a RunRef so the caller can reconnect or poll if the stream drops.
func (c *Client) Run(ctx context.Context, agentID string, opts RunEvalOptions, callback func(SimulationEvent) error) (*RunRef, error) {
	var ref RunRef
	if err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/run-eval", agentID), opts, &ref); err != nil {
		return nil, err
	}
	if err := c.Runs.StreamEvents(ctx, ref.RunID, 0, callback); err != nil {
		return &ref, err
	}
	return &ref, nil
}

// RunAsync launches a run-eval without streaming. Returns a RunRef for polling.
func (c *Client) RunAsync(ctx context.Context, agentID string, opts RunEvalOptions) (*RunRef, error) {
	var ref RunRef
	err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/run-eval", agentID), opts, &ref)
	return &ref, err
}

// ReEval re-evaluates an existing run with a different template, streams events via callback.
func (c *Client) ReEval(ctx context.Context, agentID string, opts ReEvalOptions, callback func(SimulationEvent) error) (*RunRef, error) {
	var ref RunRef
	if err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/eval-only", agentID), opts, &ref); err != nil {
		return nil, err
	}
	if err := c.Runs.StreamEvents(ctx, ref.RunID, 0, callback); err != nil {
		return &ref, err
	}
	return &ref, nil
}

// ReEvalAsync launches a re-evaluation without streaming. Returns a RunRef for polling.
func (c *Client) ReEvalAsync(ctx context.Context, agentID string, opts ReEvalOptions) (*RunRef, error) {
	var ref RunRef
	err := c.backend.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/eval-only", agentID), opts, &ref)
	return &ref, err
}
