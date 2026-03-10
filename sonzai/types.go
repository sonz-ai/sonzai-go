package sonzai

// ---------------------------------------------------------------------------
// Chat
// ---------------------------------------------------------------------------

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatChoice represents a single choice in a streaming response.
type ChatChoice struct {
	Delta        map[string]string `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
	Index        int               `json:"index"`
}

// ChatUsage represents token usage for a chat completion.
type ChatUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// ChatStreamEvent represents a single SSE event from the chat stream.
type ChatStreamEvent struct {
	Choices []ChatChoice           `json:"choices,omitempty"`
	Usage   *ChatUsage             `json:"usage,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *struct{ Message string } `json:"error,omitempty"`
}

// Content returns the text content from the first choice delta.
func (e *ChatStreamEvent) Content() string {
	if len(e.Choices) > 0 {
		return e.Choices[0].Delta["content"]
	}
	return ""
}

// IsFinished returns true if the stream has finished.
func (e *ChatStreamEvent) IsFinished() bool {
	if len(e.Choices) > 0 && e.Choices[0].FinishReason != nil {
		return *e.Choices[0].FinishReason == "stop"
	}
	return false
}

// ChatResponse is the aggregated result of a non-streaming chat call.
type ChatResponse struct {
	Content   string     `json:"content"`
	SessionID string     `json:"session_id"`
	Usage     *ChatUsage `json:"usage,omitempty"`
}

// ChatOptions configures a chat request.
type ChatOptions struct {
	Messages   []ChatMessage `json:"messages"`
	UserID     string        `json:"user_id,omitempty"`
	SessionID  string        `json:"session_id,omitempty"`
	InstanceID string        `json:"instance_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Memory
// ---------------------------------------------------------------------------

// MemoryNode represents a node in the memory tree.
type MemoryNode struct {
	NodeID     string  `json:"node_id"`
	AgentID    string  `json:"agent_id"`
	UserID     string  `json:"user_id"`
	ParentID   string  `json:"parent_id"`
	Title      string  `json:"title"`
	Summary    string  `json:"summary"`
	Importance float64 `json:"importance"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
}

// AtomicFact represents a single atomic fact stored in memory.
type AtomicFact struct {
	FactID       string  `json:"fact_id"`
	AgentID      string  `json:"agent_id"`
	UserID       string  `json:"user_id"`
	NodeID       string  `json:"node_id"`
	AtomicText   string  `json:"atomic_text"`
	FactType     string  `json:"fact_type"`
	Importance   float64 `json:"importance"`
	SupersedesID string  `json:"supersedes_id"`
	SessionID    string  `json:"session_id"`
	CreatedAt    string  `json:"created_at,omitempty"`
}

// MemoryResponse is the response from the memory list endpoint.
type MemoryResponse struct {
	Nodes    []MemoryNode              `json:"nodes"`
	Contents map[string][]AtomicFact   `json:"contents"`
}

// MemorySearchResult represents a single search result.
type MemorySearchResult struct {
	FactID   string  `json:"fact_id"`
	Content  string  `json:"content"`
	FactType string  `json:"fact_type"`
	Score    float64 `json:"score"`
}

// MemorySearchResponse is the response from the memory search endpoint.
type MemorySearchResponse struct {
	Results []MemorySearchResult `json:"results"`
}

// TimelineSession represents a session in the memory timeline.
type TimelineSession struct {
	SessionID  string       `json:"session_id"`
	Facts      []AtomicFact `json:"facts"`
	FirstFactAt string      `json:"first_fact_at,omitempty"`
	LastFactAt  string      `json:"last_fact_at,omitempty"`
	FactCount   int         `json:"fact_count"`
}

// MemoryTimelineResponse is the response from the memory timeline endpoint.
type MemoryTimelineResponse struct {
	Sessions   []TimelineSession `json:"sessions"`
	TotalFacts int               `json:"total_facts"`
}

// MemoryListOptions configures a memory list request.
type MemoryListOptions struct {
	UserID          string
	InstanceID      string
	ParentID        string
	IncludeContents bool
	Limit           int
}

// MemorySearchOptions configures a memory search request.
type MemorySearchOptions struct {
	Query      string
	InstanceID string
	Limit      int
}

// MemoryTimelineOptions configures a memory timeline request.
type MemoryTimelineOptions struct {
	UserID     string
	InstanceID string
	Start      string
	End        string
}

// ---------------------------------------------------------------------------
// Personality
// ---------------------------------------------------------------------------

// Big5Trait represents a single Big Five personality trait.
type Big5Trait struct {
	Score      float64 `json:"score"`
	Percentile int     `json:"percentile"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Big5 represents the Big Five personality model.
type Big5 struct {
	Openness          Big5Trait `json:"openness"`
	Conscientiousness Big5Trait `json:"conscientiousness"`
	Extraversion      Big5Trait `json:"extraversion"`
	Agreeableness     Big5Trait `json:"agreeableness"`
	Neuroticism       Big5Trait `json:"neuroticism"`
}

// PersonalityDimensions represents personality dimension scores (1-10).
type PersonalityDimensions struct {
	Warmth         int `json:"warmth"`
	Energy         int `json:"energy"`
	Openness       int `json:"openness"`
	EmotionalDepth int `json:"emotional_depth"`
	Playfulness    int `json:"playfulness"`
	Supportiveness int `json:"supportiveness"`
	Curiosity      int `json:"curiosity"`
	Wisdom         int `json:"wisdom"`
}

// PersonalityPreferences represents interaction preferences.
type PersonalityPreferences struct {
	Pace                string `json:"pace"`
	Formality           string `json:"formality"`
	HumorStyle          string `json:"humor_style"`
	EmotionalExpression string `json:"emotional_expression"`
}

// PersonalityBehaviors represents behavioral traits.
type PersonalityBehaviors struct {
	Proactivity string `json:"proactivity"`
	Reliability string `json:"reliability"`
	Humor       string `json:"humor"`
}

// PersonalityProfile represents the full personality profile.
type PersonalityProfile struct {
	AgentID             string              `json:"agent_id"`
	Name                string              `json:"name"`
	Gender              string              `json:"gender"`
	Bio                 string              `json:"bio"`
	AvatarURL           string              `json:"avatar_url"`
	PersonalityPrompt   string              `json:"personality_prompt"`
	SpeechPatterns      []string            `json:"speech_patterns"`
	TrueInterests       []string            `json:"true_interests"`
	TrueDislikes        []string            `json:"true_dislikes"`
	PrimaryTraits       []string            `json:"primary_traits"`
	Temperature         float64             `json:"temperature"`
	Big5                Big5                `json:"big5"`
	Dimensions          PersonalityDimensions `json:"dimensions"`
	Preferences         PersonalityPreferences `json:"preferences"`
	Behaviors           PersonalityBehaviors  `json:"behaviors"`
	EmotionalTendencies map[string]float64  `json:"emotional_tendencies"`
	CreatedAt           string              `json:"created_at,omitempty"`
}

// PersonalityDelta represents a personality evolution event.
type PersonalityDelta struct {
	DeltaID   string `json:"delta_id"`
	Change    string `json:"change"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at,omitempty"`
}

// PersonalityResponse is the response from the personality endpoint.
type PersonalityResponse struct {
	Profile   PersonalityProfile  `json:"profile"`
	Evolution []PersonalityDelta  `json:"evolution"`
}

// PersonalityGetOptions configures a personality get request.
type PersonalityGetOptions struct {
	HistoryLimit int
	Since        string
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// SessionResponse is the response from session start/end.
type SessionResponse struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------
// Instances
// ---------------------------------------------------------------------------

// AgentInstance represents an agent instance.
type AgentInstance struct {
	InstanceID  string `json:"instance_id"`
	AgentID     string `json:"agent_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IsDefault   bool   `json:"is_default"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// InstanceListResponse is the response from listing instances.
type InstanceListResponse struct {
	Instances []AgentInstance `json:"instances"`
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

// Notification represents a proactive notification.
type Notification struct {
	MessageID        string `json:"message_id"`
	AgentID          string `json:"agent_id"`
	UserID           string `json:"user_id"`
	WakeupID         string `json:"wakeup_id"`
	CheckType        string `json:"check_type"`
	Intent           string `json:"intent"`
	GeneratedMessage string `json:"generated_message"`
	Status           string `json:"status"`
	ConsumedAt       string `json:"consumed_at,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

// NotificationListResponse is the response from listing notifications.
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
}

// ---------------------------------------------------------------------------
// Evaluation
// ---------------------------------------------------------------------------

// EvalCategory represents a scored evaluation category.
type EvalCategory struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

// EvaluationResult is the response from agent evaluation.
type EvaluationResult struct {
	Score      float64        `json:"score"`
	Feedback   string         `json:"feedback"`
	Categories []EvalCategory `json:"categories"`
}

// ---------------------------------------------------------------------------
// Simulation
// ---------------------------------------------------------------------------

// SimulationEvent represents a single SSE event from simulation.
type SimulationEvent struct {
	Type             string                 `json:"type"`
	SessionIndex     int                    `json:"session_index,omitempty"`
	TotalSessions    int                    `json:"total_sessions,omitempty"`
	TotalTurns       int                    `json:"total_turns,omitempty"`
	TotalCostUSD     float64                `json:"total_cost_usd,omitempty"`
	Message          string                 `json:"message,omitempty"`
	EvalResult       map[string]interface{} `json:"eval_result,omitempty"`
	AdaptationResult map[string]interface{} `json:"adaptation_result,omitempty"`
	Error            *struct{ Message string } `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Eval Templates
// ---------------------------------------------------------------------------

// EvalTemplateCategory represents a category in an eval template.
type EvalTemplateCategory struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Criteria string  `json:"criteria"`
}

// EvalTemplate represents an evaluation template.
type EvalTemplate struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	TemplateType  string                 `json:"template_type"`
	JudgeModel    string                 `json:"judge_model"`
	Temperature   float64                `json:"temperature"`
	MaxTokens     int                    `json:"max_tokens"`
	ScoringRubric string                 `json:"scoring_rubric"`
	Categories    []EvalTemplateCategory `json:"categories"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	UpdatedAt     string                 `json:"updated_at,omitempty"`
}

// EvalTemplateListResponse is the response from listing eval templates.
type EvalTemplateListResponse struct {
	Templates []EvalTemplate `json:"templates"`
}

// ---------------------------------------------------------------------------
// Eval Runs
// ---------------------------------------------------------------------------

// EvalRun represents a completed evaluation run.
type EvalRun struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	AgentID          string                 `json:"agent_id"`
	AgentName        string                 `json:"agent_name"`
	Status           string                 `json:"status"`
	CharacterConfig  map[string]interface{} `json:"character_config"`
	TemplateID       string                 `json:"template_id"`
	TemplateSnapshot map[string]interface{} `json:"template_snapshot"`
	SimulationConfig map[string]interface{} `json:"simulation_config"`
	SimulationModel  string                 `json:"simulation_model"`
	UserPersona      map[string]interface{} `json:"user_persona"`
	Transcript       []interface{}          `json:"transcript"`
	EvaluationResult map[string]interface{} `json:"evaluation_result"`
	AdaptationResult map[string]interface{} `json:"adaptation_result"`
	SimulationState  map[string]interface{} `json:"simulation_state"`
	TotalSessions    int                    `json:"total_sessions"`
	TotalTurns       int                    `json:"total_turns"`
	SimulatedMinutes int                    `json:"simulated_minutes"`
	TotalCostUSD     float64                `json:"total_cost_usd"`
	CreatedAt        string                 `json:"created_at,omitempty"`
	CompletedAt      string                 `json:"completed_at,omitempty"`
}

// EvalRunListResponse is the response from listing eval runs.
type EvalRunListResponse struct {
	Runs       []EvalRun `json:"runs"`
	TotalCount int       `json:"total_count"`
}
