package sonzai

import "encoding/json"

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
	Choices []ChatChoice              `json:"choices,omitempty"`
	Usage   *ChatUsage                `json:"usage,omitempty"`
	Type    string                    `json:"type,omitempty"`
	Data    map[string]interface{}    `json:"data,omitempty"`
	Error   *struct{ Message string } `json:"error,omitempty"`

	// Rich event fields (populated based on Type)
	MessageIndex      int             `json:"message_index,omitempty"`
	IsFollowUp        bool            `json:"is_follow_up,omitempty"`
	Replacement       bool            `json:"replacement,omitempty"`
	FullContent       string          `json:"full_content,omitempty"`
	FinishReason      string          `json:"finish_reason,omitempty"`
	ContinuationToken string          `json:"continuation_token,omitempty"`
	ResponseCookie    string          `json:"response_cookie,omitempty"`
	MessageCount      int             `json:"message_count,omitempty"`
	SideEffectsJSON   json.RawMessage `json:"side_effects,omitempty"`
	EnrichedContext   json.RawMessage `json:"enriched_context,omitempty"`
	BuildDurationMs   int64           `json:"build_duration_ms,omitempty"`
	UsedFastPath      bool            `json:"used_fast_path,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
	ErrorCode         string          `json:"error_code,omitempty"`
	IsTokenError      bool            `json:"is_token_error,omitempty"`
}

// Content returns the text content from the first choice delta,
// or from the Data["content"] field for rich events.
func (e *ChatStreamEvent) Content() string {
	if len(e.Choices) > 0 {
		return e.Choices[0].Delta["content"]
	}
	if e.Data != nil {
		if c, ok := e.Data["content"].(string); ok {
			return c
		}
	}
	return ""
}

// IsFinished returns true if the stream has finished.
func (e *ChatStreamEvent) IsFinished() bool {
	if len(e.Choices) > 0 && e.Choices[0].FinishReason != nil {
		return *e.Choices[0].FinishReason == "stop"
	}
	return e.Type == "complete"
}

// ChatResponse is the aggregated result of a non-streaming chat call.
type ChatResponse struct {
	Content   string     `json:"content"`
	SessionID string     `json:"session_id"`
	Usage     *ChatUsage `json:"usage,omitempty"`
}

// AgentToolCapabilities specifies which built-in tools to enable for an agent.
type AgentToolCapabilities struct {
	WebSearch       bool `json:"web_search"`
	RememberName    bool `json:"remember_name"`
	ImageGeneration bool `json:"image_generation"`
	Inventory       bool `json:"inventory"`
}

// AgentChatParams is the single-struct params type for Chat, ChatStream, and ChatStreamChannel.
// AgentID is used as the URL path parameter; all other fields are sent as the request body.
type AgentChatParams struct {
	AgentID string `json:"-"`
	ChatOptions
}

// ChatOptions configures a chat request.
type ChatOptions struct {
	Messages             []ChatMessage         `json:"messages"`
	UserID               string                `json:"user_id,omitempty"`
	UserDisplayName      string                `json:"user_display_name,omitempty"`
	SessionID            string                `json:"session_id,omitempty"`
	InstanceID           string                `json:"instance_id,omitempty"`
	Provider             string                `json:"provider,omitempty"`
	Model                string                `json:"model,omitempty"`
	ContinuationToken    string                `json:"continuation_token,omitempty"`
	AiServiceCookie      string                `json:"ai_service_cookie,omitempty"`
	RequestType          string                `json:"request_type,omitempty"`
	Language             string                `json:"language,omitempty"`
	CompiledSystemPrompt string                `json:"compiled_system_prompt,omitempty"`
	InteractionRole      string                `json:"interaction_role,omitempty"`
	Timezone             string                `json:"timezone,omitempty"`
	ToolCapabilities     *AgentToolCapabilities `json:"tool_capabilities,omitempty"`
	ToolDefinitions      []ToolDefinition       `json:"tool_definitions,omitempty"`
	MaxTurns             int                    `json:"max_turns,omitempty"`
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
	FactID       string                 `json:"fact_id"`
	AgentID      string                 `json:"agent_id"`
	UserID       string                 `json:"user_id"`
	NodeID       string                 `json:"node_id"`
	AtomicText   string                 `json:"atomic_text"`
	FactType     string                 `json:"fact_type"`
	Importance   float64                `json:"importance"`
	SupersedesID string                 `json:"supersedes_id"`
	SessionID    string                 `json:"session_id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
}

// MemoryResponse is the response from the memory list endpoint.
type MemoryResponse struct {
	Nodes    []MemoryNode            `json:"nodes"`
	Contents map[string][]AtomicFact `json:"contents"`
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
	SessionID   string       `json:"session_id"`
	Facts       []AtomicFact `json:"facts"`
	FirstFactAt string       `json:"first_fact_at,omitempty"`
	LastFactAt  string       `json:"last_fact_at,omitempty"`
	FactCount   int          `json:"fact_count"`
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
	AgentID             string                 `json:"agent_id"`
	Name                string                 `json:"name"`
	Gender              string                 `json:"gender"`
	Bio                 string                 `json:"bio"`
	AvatarURL           string                 `json:"avatar_url"`
	PersonalityPrompt   string                 `json:"personality_prompt"`
	SpeechPatterns      []string               `json:"speech_patterns"`
	TrueInterests       []string               `json:"true_interests"`
	TrueDislikes        []string               `json:"true_dislikes"`
	PrimaryTraits       []string               `json:"primary_traits"`
	Temperature         float64                `json:"temperature"`
	Big5                Big5                   `json:"big5"`
	Dimensions          PersonalityDimensions  `json:"dimensions"`
	Preferences         PersonalityPreferences `json:"preferences"`
	Behaviors           PersonalityBehaviors   `json:"behaviors"`
	EmotionalTendencies map[string]float64     `json:"emotional_tendencies"`
	CreatedAt           string                 `json:"created_at,omitempty"`
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
	Profile   PersonalityProfile `json:"profile"`
	Evolution []PersonalityDelta `json:"evolution"`
}

// PersonalityGetOptions configures a personality get request.
type PersonalityGetOptions struct {
	HistoryLimit int
	Since        string
}

// Big5Scores represents raw Big5 personality scores for updates (0-100 scale).
type Big5Scores struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
	Confidence        float64 `json:"confidence,omitempty"`
}

// PersonalityUpdateOptions configures a personality update request.
type PersonalityUpdateOptions struct {
	Big5             *Big5Scores `json:"big5"`
	AssessmentMethod string      `json:"assessment_method,omitempty"` // "quiz", "conversation", "llm_analysis"
	TotalExchanges   int         `json:"total_exchanges,omitempty"`
}

// PersonalityUpdateResponse is the response from updating personality.
type PersonalityUpdateResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
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
// Memory Facts
// ---------------------------------------------------------------------------

// Fact represents a single extracted fact from agent memory.
type Fact struct {
	FactID          string   `json:"fact_id"`
	AgentID         string   `json:"agent_id"`
	UserID          string   `json:"user_id,omitempty"`
	Content         string   `json:"content"`
	Category        string   `json:"category"` // "relationship", "preference", "event", "interest"
	Confidence      float64  `json:"confidence"`
	MentionCount    int      `json:"mention_count"`
	CreatedAt       string   `json:"created_at,omitempty"`
	LastMentionedAt string   `json:"last_mentioned_at,omitempty"`
	ContextExamples []string `json:"context_examples,omitempty"`
}

// FactListResponse is the response from listing facts.
type FactListResponse struct {
	Facts      []Fact `json:"facts"`
	TotalCount int    `json:"total_count"`
	HasMore    bool   `json:"has_more"`
}

// FactListOptions configures a fact listing request.
type FactListOptions struct {
	UserID   string
	Category string // "relationship", "preference", "event", "interest"
	Limit    int
	Offset   int
}

// MemoryResetOptions configures a memory reset request.
type MemoryResetOptions struct {
	UserID     string // scope to single user; empty = reset all
	InstanceID string // scope to single instance
}

// MemoryResetResponse is the response from resetting memory.
type MemoryResetResponse struct {
	AgentID              string `json:"agent_id"`
	UserID               string `json:"user_id,omitempty"`
	Status               string `json:"status"`
	FactsDeleted         int    `json:"facts_deleted"`
	RelationshipsDeleted int    `json:"relationships_deleted"`
}

// ---------------------------------------------------------------------------
// Events
// ---------------------------------------------------------------------------

// TriggerEventOptions configures an event trigger request.
type TriggerEventOptions struct {
	UserID           string            `json:"user_id"`
	EventType        string            `json:"event_type"`                  // e.g., "achievement", "milestone", "level_up"
	EventDescription string            `json:"event_description,omitempty"` // Human-readable context for the AI
	Metadata         map[string]string `json:"metadata,omitempty"`
	Language         string            `json:"language,omitempty"`
	InstanceID       string            `json:"instance_id,omitempty"`
}

// TriggerEventResponse is the response from triggering an event.
type TriggerEventResponse struct {
	Accepted bool   `json:"accepted"`
	EventID  string `json:"event_id"`
}

// ---------------------------------------------------------------------------
// Dialogue
// ---------------------------------------------------------------------------

// DialogueOptions configures a dialogue request.
type DialogueOptions struct {
	UserID              string          `json:"user_id,omitempty"`
	EnrichedContextJSON json.RawMessage `json:"enriched_context,omitempty"`
	Messages            []ChatMessage   `json:"messages,omitempty"`
	RequestType         string          `json:"request_type,omitempty"`
	SceneGuidance       string          `json:"scene_guidance,omitempty"`
	ToolConfigJSON      json.RawMessage `json:"tool_config,omitempty"`
	InstanceID          string          `json:"instance_id,omitempty"`
}

// DialogueResponse is the response from a dialogue.
type DialogueResponse struct {
	Response        string          `json:"response"`
	SideEffectsJSON json.RawMessage `json:"side_effects,omitempty"`
}

// DialogueMessage represents a single message in a dialogue.
type DialogueMessage struct {
	MessageID string `json:"message_id"`
	AgentID   string `json:"agent_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

// ---------------------------------------------------------------------------
// Agent List
// ---------------------------------------------------------------------------

// AgentListOptions configures an agent list request.
type AgentListOptions struct {
	PageSize  int    `json:"page_size,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Search    string `json:"search,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// AgentIndex represents a summary of an agent in a list.
type AgentIndex struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio,omitempty"`
	Gender    string `json:"gender,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Status    string `json:"status,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// AgentListResponse is the response from listing agents.
type AgentListResponse struct {
	Items      []AgentIndex `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
}

// ---------------------------------------------------------------------------
// Agent Status
// ---------------------------------------------------------------------------

// SetStatusResponse is the response from setting agent status.
type SetStatusResponse struct {
	Success  bool   `json:"success"`
	AgentID  string `json:"agent_id"`
	IsActive bool   `json:"is_active"`
}

// ---------------------------------------------------------------------------
// Agent Project
// ---------------------------------------------------------------------------

// UpdateProjectResponse is the response from updating agent project.
type UpdateProjectResponse struct {
	Success   bool   `json:"success"`
	AgentID   string `json:"agent_id"`
	ProjectID string `json:"project_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Agent Capabilities
// ---------------------------------------------------------------------------

// CustomToolDefinition represents a custom tool definition for an agent.
type CustomToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AgentCapabilities represents the capabilities enabled for an agent.
type AgentCapabilities struct {
	WebSearch       bool                   `json:"webSearch"`
	RememberName    bool                   `json:"rememberName"`
	ImageGeneration bool                   `json:"imageGeneration"`
	Inventory       bool                   `json:"inventory"`
	CustomTools     []CustomToolDefinition `json:"customTools,omitempty"`
}

// UpdateCapabilitiesOptions configures a capabilities update request.
type UpdateCapabilitiesOptions struct {
	WebSearch       *bool `json:"webSearch,omitempty"`
	RememberName    *bool `json:"rememberName,omitempty"`
	ImageGeneration *bool `json:"imageGeneration,omitempty"`
	Inventory       *bool `json:"inventory,omitempty"`
}

// CustomToolListResponse is the response from listing custom tools.
type CustomToolListResponse struct {
	Tools []CustomToolDefinition `json:"tools"`
}

// CreateCustomToolOptions configures a custom tool creation request.
type CreateCustomToolOptions struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// UpdateCustomToolOptions configures a custom tool update request.
type UpdateCustomToolOptions struct {
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ---------------------------------------------------------------------------
// Consolidation & Summaries
// ---------------------------------------------------------------------------

// ConsolidateOptions configures a consolidation request.
type ConsolidateOptions struct {
	Period string `json:"period,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// SummariesOptions configures a summaries request.
type SummariesOptions struct {
	Period string
	Limit  int
}

// MemorySummary represents a memory consolidation summary.
type MemorySummary struct {
	AgentID     string  `json:"agent_id,omitempty"`
	Stage       string  `json:"stage,omitempty"`
	SummaryText string  `json:"summary_text,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
	FactCount   int     `json:"fact_count,omitempty"`
	Confidence  float64 `json:"confidence,omitempty"`
}

// SummariesResponse is the response from the summaries endpoint.
type SummariesResponse struct {
	Summaries []MemorySummary `json:"summaries"`
}

// ---------------------------------------------------------------------------
// Time Machine
// ---------------------------------------------------------------------------

// TimeMachineOptions configures a time machine request.
type TimeMachineOptions struct {
	At         string `json:"at"`
	UserID     string `json:"user_id,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
}

// TimeMachineMoodSnapshot represents a mood snapshot at a point in time.
type TimeMachineMoodSnapshot struct {
	Valence     float64 `json:"valence"`
	Arousal     float64 `json:"arousal"`
	Tension     float64 `json:"tension"`
	Affiliation float64 `json:"affiliation"`
	Label       string  `json:"label"`
}

// TimeMachineResponse is the response from the time machine endpoint.
type TimeMachineResponse struct {
	PersonalityAt      map[string]interface{}   `json:"personality_at,omitempty"`
	CurrentPersonality map[string]interface{}   `json:"current_personality,omitempty"`
	EvolutionEvents    []PersonalityShift       `json:"evolution_events,omitempty"`
	MoodAt             *TimeMachineMoodSnapshot `json:"mood_at,omitempty"`
	RequestedAt        string                   `json:"requested_at,omitempty"`
}

// ---------------------------------------------------------------------------
// Batch Personality
// ---------------------------------------------------------------------------

// BatchPersonalityEntry represents a single entry in a batch personality response.
type BatchPersonalityEntry struct {
	Profile        PersonalityProfile `json:"profile"`
	EvolutionCount int                `json:"evolution_count"`
}

// BatchPersonalityResponse is the response from the batch personality endpoint.
type BatchPersonalityResponse struct {
	Personalities map[string]BatchPersonalityEntry `json:"personalities"`
}

// ---------------------------------------------------------------------------
// Personality Extensions
// ---------------------------------------------------------------------------

// SignificantMoment represents a significant moment in agent history.
type SignificantMoment struct {
	AgentID           string  `json:"agent_id,omitempty"`
	MomentID          string  `json:"moment_id,omitempty"`
	Timestamp         string  `json:"timestamp,omitempty"`
	Description       string  `json:"description,omitempty"`
	SignificanceScore float64 `json:"significance_score,omitempty"`
}

// SignificantMomentsResponse is the response from the significant moments endpoint.
type SignificantMomentsResponse struct {
	Moments []SignificantMoment `json:"moments"`
}

// PersonalityShift represents a personality trait shift event.
type PersonalityShift struct {
	AgentID       string  `json:"agent_id,omitempty"`
	TraitName     string  `json:"trait_name,omitempty"`
	TraitCategory string  `json:"trait_category,omitempty"`
	OldValue      float64 `json:"old_value,omitempty"`
	NewValue      float64 `json:"new_value,omitempty"`
	Delta         float64 `json:"delta,omitempty"`
	Timestamp     string  `json:"timestamp,omitempty"`
	Reason        string  `json:"reason,omitempty"`
}

// RecentShiftsResponse is the response from the recent shifts endpoint.
type RecentShiftsResponse struct {
	Shifts []PersonalityShift `json:"shifts"`
}

// UserPersonalityOverlay represents a per-user personality overlay.
type UserPersonalityOverlay struct {
	AgentID       string                  `json:"agent_id"`
	UserID        string                  `json:"user_id"`
	Big5          *Big5                   `json:"big5,omitempty"`
	Dimensions    *PersonalityDimensions  `json:"dimensions,omitempty"`
	Preferences   *PersonalityPreferences `json:"preferences,omitempty"`
	Behaviors     *PersonalityBehaviors   `json:"behaviors,omitempty"`
	PrimaryTraits []string                `json:"primary_traits,omitempty"`
	CreatedAt     string                  `json:"created_at,omitempty"`
	UpdatedAt     string                  `json:"updated_at,omitempty"`
}

// UserOverlaysListResponse is the response from listing user overlays.
type UserOverlaysListResponse struct {
	Overlays []UserPersonalityOverlay `json:"overlays"`
}

// UserOverlayDetailResponse is the response from getting a user overlay detail.
type UserOverlayDetailResponse struct {
	Overlay   UserPersonalityOverlay `json:"overlay"`
	Base      PersonalityProfile     `json:"base"`
	Evolution []PersonalityShift     `json:"evolution"`
}

// UserOverlayOptions configures a user overlay request.
type UserOverlayOptions struct {
	InstanceID string `json:"instance_id,omitempty"`
	Since      string `json:"since,omitempty"`
}

// ---------------------------------------------------------------------------
// Fact History
// ---------------------------------------------------------------------------

// FactHistoryResponse is the response from the fact history endpoint.
type FactHistoryResponse struct {
	Current          AtomicFact   `json:"current"`
	PreviousVersions []AtomicFact `json:"previous_versions"`
}

// ---------------------------------------------------------------------------
// Update Instance
// ---------------------------------------------------------------------------

// UpdateInstanceOptions configures an instance update request.
type UpdateInstanceOptions struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}
