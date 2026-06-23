package sonzai

import "context"

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

// PipelineRun is the result of executing a pipeline end to end.
type PipelineRun struct {
	PipelineID    string               `json:"pipeline_id"`
	Steps         []PipelineStepResult `json:"steps"`
	FinalFindings interface{}          `json:"final_findings"`
	TotalCostUSD  float64              `json:"total_cost_usd"`
	Completed     bool                 `json:"completed"`
}

// PipelineListResponse is the response from listing pipelines.
type PipelineListResponse struct {
	Pipelines []Pipeline `json:"pipelines"`
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

// Run executes the pipeline end to end, threading each step's findings into the
// next, and returns the full run. Uses the long-running transport since a
// multi-step run can take minutes (each step bills like any invocation).
func (r *PipelinesResource) Run(ctx context.Context, pipelineID string, input map[string]interface{}) (*PipelineRun, error) {
	var out PipelineRun
	body := map[string]interface{}{"input": input}
	if err := r.http.PostLongRunning(ctx, "/api/v1/pipelines/"+pipelineID+"/run", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
