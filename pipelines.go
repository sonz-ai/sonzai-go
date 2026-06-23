package sonzai

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Custom agents — a company's OWN backend agents (own prompt/model/schema/tools).
// They provision, run, and bill identically to Sonzai built-ins and can be
// chained into pipelines alongside them.
// ---------------------------------------------------------------------------

// CustomAgentsResource manages tenant-defined backend agents.
type CustomAgentsResource struct {
	http *httpClient
}

// CustomAgent is a tenant-defined backend agent.
type CustomAgent struct {
	AgentID        string                 `json:"agent_id"`
	ProjectID      string                 `json:"project_id"`
	Slug           string                 `json:"slug"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Model          string                 `json:"model"`
	System         string                 `json:"system"`
	FindingsSchema map[string]interface{} `json:"findings_schema,omitempty"`
	Tools          []string               `json:"tools"`
	DisableTools   bool                   `json:"disable_tools"`
	MaxToolRounds  int                    `json:"max_tool_rounds"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// CustomAgentInput is the create/update payload. Model must be an allowed
// Anthropic model (claude-sonnet-4-6 | claude-haiku-4-5).
type CustomAgentInput struct {
	Slug           string                 `json:"slug"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Model          string                 `json:"model"`
	System         string                 `json:"system"`
	FindingsSchema map[string]interface{} `json:"findings_schema,omitempty"`
	Tools          []string               `json:"tools,omitempty"`
	DisableTools   bool                   `json:"disable_tools,omitempty"`
	MaxToolRounds  int                    `json:"max_tool_rounds,omitempty"`
}

// CustomAgentListResponse is the response from listing custom agents.
type CustomAgentListResponse struct {
	Agents []CustomAgent `json:"agents"`
}

func (r *CustomAgentsResource) List(ctx context.Context) (*CustomAgentListResponse, error) {
	var out CustomAgentListResponse
	if err := r.http.Get(ctx, "/api/v1/custom-agents", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomAgentsResource) Create(ctx context.Context, in CustomAgentInput) (*CustomAgent, error) {
	var out CustomAgent
	if err := r.http.Post(ctx, "/api/v1/custom-agents", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomAgentsResource) Get(ctx context.Context, agentID string) (*CustomAgent, error) {
	var out CustomAgent
	if err := r.http.Get(ctx, "/api/v1/custom-agents/"+agentID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomAgentsResource) Update(ctx context.Context, agentID string, in CustomAgentInput) (*CustomAgent, error) {
	var out CustomAgent
	if err := r.http.Put(ctx, "/api/v1/custom-agents/"+agentID, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *CustomAgentsResource) Delete(ctx context.Context, agentID string) error {
	return r.http.Delete(ctx, "/api/v1/custom-agents/"+agentID, nil)
}

// ---------------------------------------------------------------------------
// Pipelines — an ordered chain of agent slugs (built-in or custom) that run
// one after another, each step fed the prior steps' findings.
// ---------------------------------------------------------------------------

// PipelinesResource manages tenant-defined agent pipelines.
type PipelinesResource struct {
	http *httpClient
}

// PipelineStep references an agent to run (built-in slug or custom agent slug).
type PipelineStep struct {
	Slug  string `json:"slug"`
	Title string `json:"title,omitempty"`
}

// Pipeline is an ordered chain of agent steps.
type Pipeline struct {
	PipelineID  string         `json:"pipeline_id"`
	ProjectID   string         `json:"project_id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Steps       []PipelineStep `json:"steps"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// PipelineInput is the create/update payload.
type PipelineInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Steps       []PipelineStep `json:"steps,omitempty"`
}

// PipelineStepResult is one step's output in a run.
type PipelineStepResult struct {
	Slug     string      `json:"slug"`
	Title    string      `json:"title,omitempty"`
	Findings interface{} `json:"findings"`
	Summary  string      `json:"summary,omitempty"`
	CostUSD  float64     `json:"cost_usd"`
	Error    string      `json:"error,omitempty"`
}

// PipelineRun is an async execution of a pipeline. Starting a run returns this
// with Status "queued" and a RunID; poll GetRun for progress. Status is one of
// "queued" | "running" | "completed" | "failed".
type PipelineRun struct {
	RunID         string               `json:"run_id"`
	PipelineID    string               `json:"pipeline_id"`
	Status        string               `json:"status"`
	Steps         []PipelineStepResult `json:"steps"`
	FinalFindings interface{}          `json:"final_findings"`
	TotalCostUSD  float64              `json:"total_cost_usd"`
	Error         string               `json:"error,omitempty"`
	Completed     bool                 `json:"completed"`
	CreatedAt     string               `json:"created_at,omitempty"`
	UpdatedAt     string               `json:"updated_at,omitempty"`
}

// Terminal reports whether the run has finished (completed or failed).
func (r *PipelineRun) Terminal() bool {
	return r.Status == "completed" || r.Status == "failed"
}

// PipelineListResponse is the response from listing pipelines.
type PipelineListResponse struct {
	Pipelines []Pipeline `json:"pipelines"`
}

// PipelineRunListResponse is the response from listing runs.
type PipelineRunListResponse struct {
	Runs []PipelineRun `json:"runs"`
}

func (r *PipelinesResource) List(ctx context.Context) (*PipelineListResponse, error) {
	var out PipelineListResponse
	if err := r.http.Get(ctx, "/api/v1/pipelines", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PipelinesResource) Create(ctx context.Context, in PipelineInput) (*Pipeline, error) {
	var out Pipeline
	if err := r.http.Post(ctx, "/api/v1/pipelines", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PipelinesResource) Get(ctx context.Context, pipelineID string) (*Pipeline, error) {
	var out Pipeline
	if err := r.http.Get(ctx, "/api/v1/pipelines/"+pipelineID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PipelinesResource) Update(ctx context.Context, pipelineID string, in PipelineInput) (*Pipeline, error) {
	var out Pipeline
	if err := r.http.Put(ctx, "/api/v1/pipelines/"+pipelineID, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PipelinesResource) Delete(ctx context.Context, pipelineID string) error {
	return r.http.Delete(ctx, "/api/v1/pipelines/"+pipelineID, nil)
}

// AppendStep adds a step to the end of an existing pipeline.
func (r *PipelinesResource) AppendStep(ctx context.Context, pipelineID string, step PipelineStep) (*Pipeline, error) {
	var out Pipeline
	if err := r.http.Post(ctx, "/api/v1/pipelines/"+pipelineID+"/steps", step, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Run starts an asynchronous pipeline run and returns immediately with the
// queued run (RunID + Status "queued"). Poll GetRun for progress + results, or
// use RunAndWait to block until the run finishes.
func (r *PipelinesResource) Run(ctx context.Context, pipelineID string, input map[string]interface{}) (*PipelineRun, error) {
	var out PipelineRun
	body := map[string]interface{}{"input": input}
	if err := r.http.Post(ctx, "/api/v1/pipelines/"+pipelineID+"/run", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRun returns a single run — poll this for status and accumulated results.
func (r *PipelinesResource) GetRun(ctx context.Context, pipelineID, runID string) (*PipelineRun, error) {
	var out PipelineRun
	if err := r.http.Get(ctx, "/api/v1/pipelines/"+pipelineID+"/runs/"+runID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRuns returns recent runs for a pipeline, newest first.
func (r *PipelinesResource) ListRuns(ctx context.Context, pipelineID string) (*PipelineRunListResponse, error) {
	var out PipelineRunListResponse
	if err := r.http.Get(ctx, "/api/v1/pipelines/"+pipelineID+"/runs", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RunAndWait starts a run and polls until it finishes (completed or failed) or
// the context is cancelled. pollInterval defaults to 2s. Convenience over
// Run + GetRun for callers that just want the final result.
func (r *PipelinesResource) RunAndWait(ctx context.Context, pipelineID string, input map[string]interface{}, pollInterval time.Duration) (*PipelineRun, error) {
	run, err := r.Run(ctx, pipelineID, input)
	if err != nil {
		return nil, err
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	for !run.Terminal() {
		select {
		case <-ctx.Done():
			return run, ctx.Err()
		case <-time.After(pollInterval):
		}
		got, err := r.GetRun(ctx, pipelineID, run.RunID)
		if err != nil {
			return run, err
		}
		run = got
	}
	return run, nil
}
