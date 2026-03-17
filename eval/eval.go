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

// Simulate runs a multi-turn simulation and calls the callback for each event.
func (c *Client) Simulate(ctx context.Context, agentID string, opts SimulateOptions, callback func(SimulationEvent) error) error {
	return c.backend.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/simulate", agentID), opts, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// Run executes simulation + evaluation combined and calls the callback for each event.
func (c *Client) Run(ctx context.Context, agentID string, opts SimulateOptions, callback func(SimulationEvent) error) error {
	return c.backend.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/run-eval", agentID), opts, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// ReEval re-evaluates an existing run with a different template.
func (c *Client) ReEval(ctx context.Context, agentID string, opts ReEvalOptions, callback func(SimulationEvent) error) error {
	return c.backend.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/eval-only", agentID), opts, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}
