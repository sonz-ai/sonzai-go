package eval

// ---------------------------------------------------------------------------
// Shared
// ---------------------------------------------------------------------------

// Message represents a single message in a conversation for evaluation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ---------------------------------------------------------------------------
// Evaluate
// ---------------------------------------------------------------------------

// EvaluateOptions configures an evaluation request.
type EvaluateOptions struct {
	Messages       []Message              `json:"messages"`
	TemplateID     string                 `json:"template_id"`
	ConfigOverride map[string]interface{} `json:"config_override,omitempty"`
}

// EvaluationResult is the response from agent evaluation.
type EvaluationResult struct {
	Score      float64    `json:"score"`
	Feedback   string     `json:"feedback"`
	Categories []Category `json:"categories"`
}

// Category represents a scored evaluation category.
type Category struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

// ---------------------------------------------------------------------------
// Simulation
// ---------------------------------------------------------------------------

// SessionConfig configures a single session within a simulation.
type SessionConfig struct {
	Greeting    string `json:"greeting,omitempty"`
	Scenario    string `json:"scenario,omitempty"`
	Instruction string `json:"instruction,omitempty"`
}

// SimulateOptions configures a simulation request (POST /simulate).
type SimulateOptions struct {
	Sessions       []SessionConfig        `json:"sessions,omitempty"`
	UserPersona    interface{}            `json:"user_persona,omitempty"`
	Config         interface{}            `json:"config,omitempty"`
	Model          string                 `json:"model,omitempty"`
	ConfigOverride map[string]interface{} `json:"config_override,omitempty"`
}

// RunEvalOptions configures a run-eval request (POST /run-eval).
type RunEvalOptions struct {
	TemplateID           string                 `json:"template_id"`
	AdaptationTemplateID string                 `json:"adaptation_template_id,omitempty"`
	QualityOnly          bool                   `json:"quality_only,omitempty"`
	Sessions             []SessionConfig        `json:"sessions,omitempty"`
	UserPersona          interface{}            `json:"user_persona,omitempty"`
	SimulationConfig     interface{}            `json:"simulation_config,omitempty"`
	Model                string                 `json:"model,omitempty"`
	ConfigOverride       map[string]interface{} `json:"config_override,omitempty"`
}

// SimulationEvent represents a single SSE event from a simulation.
type SimulationEvent struct {
	Type             string                    `json:"type"`
	SessionIndex     int                       `json:"session_index,omitempty"`
	TotalSessions    int                       `json:"total_sessions,omitempty"`
	TotalTurns       int                       `json:"total_turns,omitempty"`
	TotalCostUSD     float64                   `json:"total_cost_usd,omitempty"`
	Message          string                    `json:"message,omitempty"`
	EvalResult       map[string]interface{}    `json:"eval_result,omitempty"`
	AdaptationResult map[string]interface{}    `json:"adaptation_result,omitempty"`
	Error            *struct{ Message string } `json:"error,omitempty"`
}

// ReEvalOptions configures a re-evaluation request.
type ReEvalOptions struct {
	TemplateID           string `json:"template_id"`
	SourceRunID          string `json:"source_run_id"`
	AdaptationTemplateID string `json:"adaptation_template_id,omitempty"`
	QualityOnly          bool   `json:"quality_only,omitempty"`
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

// TemplateCategory represents a category in an eval template.
type TemplateCategory struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Criteria string  `json:"criteria"`
}

// Template represents an evaluation template.
type Template struct {
	ID            string             `json:"id"`
	TenantID      string             `json:"tenant_id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	TemplateType  string             `json:"template_type"`
	JudgeModel    string             `json:"judge_model"`
	Temperature   float64            `json:"temperature"`
	MaxTokens     int                `json:"max_tokens"`
	ScoringRubric string             `json:"scoring_rubric"`
	Categories    []TemplateCategory `json:"categories"`
	CreatedAt     string             `json:"created_at,omitempty"`
	UpdatedAt     string             `json:"updated_at,omitempty"`
}

// TemplateListResponse is the response from listing eval templates.
type TemplateListResponse struct {
	Templates []Template `json:"templates"`
}

// TemplateCreateOptions configures a template creation request.
type TemplateCreateOptions struct {
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	TemplateType  string             `json:"template_type,omitempty"`
	JudgeModel    string             `json:"judge_model,omitempty"`
	Temperature   float64            `json:"temperature,omitempty"`
	MaxTokens     int                `json:"max_tokens,omitempty"`
	ScoringRubric string             `json:"scoring_rubric,omitempty"`
	Categories    []TemplateCategory `json:"categories,omitempty"`
}

// TemplateUpdateOptions configures a template update request.
type TemplateUpdateOptions struct {
	Name          *string            `json:"name,omitempty"`
	Description   *string            `json:"description,omitempty"`
	TemplateType  *string            `json:"template_type,omitempty"`
	JudgeModel    *string            `json:"judge_model,omitempty"`
	Temperature   *float64           `json:"temperature,omitempty"`
	MaxTokens     *int               `json:"max_tokens,omitempty"`
	ScoringRubric *string            `json:"scoring_rubric,omitempty"`
	Categories    []TemplateCategory `json:"categories,omitempty"`
}

// ---------------------------------------------------------------------------
// Runs
// ---------------------------------------------------------------------------

// RunRef is the response from launching a run (fire-and-forget).
type RunRef struct {
	RunID  string `json:"run_id"`
	Status string `json:"status"`
}

// Run represents a completed evaluation run.
type Run struct {
	ID                         string                 `json:"id"`
	TenantID                   string                 `json:"tenant_id"`
	AgentID                    string                 `json:"agent_id"`
	AgentName                  string                 `json:"agent_name"`
	Status                     string                 `json:"status"`
	ErrorReason                string                 `json:"error_reason,omitempty"`
	CharacterConfig            map[string]interface{} `json:"character_config"`
	TemplateID                 string                 `json:"template_id"`
	TemplateSnapshot           map[string]interface{} `json:"template_snapshot"`
	SimulationConfig           map[string]interface{} `json:"simulation_config"`
	SimulationModel            string                 `json:"simulation_model"`
	UserPersona                map[string]interface{} `json:"user_persona"`
	Transcript                 []interface{}          `json:"transcript"`
	EvaluationResult           map[string]interface{} `json:"evaluation_result"`
	AdaptationResult           map[string]interface{} `json:"adaptation_result"`
	SimulationState            map[string]interface{} `json:"simulation_state"`
	TotalSessions              int                    `json:"total_sessions"`
	TotalTurns                 int                    `json:"total_turns"`
	SimulatedMinutes           int                    `json:"simulated_minutes"`
	TotalCostUSD               float64                `json:"total_cost_usd"`
	SimulationCostUSD          float64                `json:"simulation_cost_usd,omitempty"`
	EvaluationCostUSD          float64                `json:"evaluation_cost_usd,omitempty"`
	AdaptationTemplateID       string                 `json:"adaptation_template_id,omitempty"`
	AdaptationTemplateSnapshot map[string]interface{} `json:"adaptation_template_snapshot,omitempty"`
	CreatedAt                  string                 `json:"created_at,omitempty"`
	CompletedAt                string                 `json:"completed_at,omitempty"`
}

// RunListResponse is the response from listing eval runs.
type RunListResponse struct {
	Runs       []Run `json:"runs"`
	TotalCount int   `json:"total_count"`
}
