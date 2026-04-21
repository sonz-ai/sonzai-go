package sonzai

import (
	"encoding/json"
)

// GameContext provides game-specific context for chat requests.
type GameContext struct {
	CustomFields  map[string]string `json:"custom_fields,omitempty"`
	GameStateJSON json.RawMessage   `json:"game_state_json,omitempty"`
}

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
	MessageIndex      int                `json:"message_index,omitempty"`
	IsFollowUp        bool               `json:"is_follow_up,omitempty"`
	Replacement       bool               `json:"replacement,omitempty"`
	FullContent       string             `json:"full_content,omitempty"`
	FinishReason      string             `json:"finish_reason,omitempty"`
	ContinuationToken string             `json:"continuation_token,omitempty"`
	ResponseCookie    string             `json:"response_cookie,omitempty"`
	MessageCount      int                `json:"message_count,omitempty"`
	SideEffectsJSON   json.RawMessage    `json:"side_effects,omitempty"`
	ExternalToolCalls []ExternalToolCall `json:"external_tool_calls,omitempty"`
	EnrichedContext   json.RawMessage    `json:"enriched_context,omitempty"`
	BuildDurationMs   int64              `json:"build_duration_ms,omitempty"`
	UsedFastPath      bool               `json:"used_fast_path,omitempty"`
	ErrorMessage      string             `json:"error_message,omitempty"`
	ErrorCode         string             `json:"error_code,omitempty"`
	IsTokenError      bool               `json:"is_token_error,omitempty"`
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

type ExternalToolCall struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type ToolCallResponseOptions struct {
	SessionID  string `json:"session_id"`
	UserID     string `json:"user_id,omitempty"`
	ToolCallID string `json:"tool_call_id"`
	Result     any    `json:"result"`
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
// AgentID may be a UUID or an agent name — names are resolved to deterministic UUIDs server-side.
type AgentChatParams struct {
	AgentID string `json:"-"`
	ChatOptions
}

// ChatOptions configures a chat request.
type ChatOptions struct {
	Messages             []ChatMessage          `json:"messages"`
	UserID               string                 `json:"user_id,omitempty"`
	UserDisplayName      string                 `json:"user_display_name,omitempty"`
	SessionID            string                 `json:"session_id,omitempty"`
	InstanceID           string                 `json:"instance_id,omitempty"`
	Provider             string                 `json:"provider,omitempty"`
	Model                string                 `json:"model,omitempty"`
	ContinuationToken    string                 `json:"continuation_token,omitempty"`
	AiServiceCookie      string                 `json:"ai_service_cookie,omitempty"`
	RequestType          string                 `json:"request_type,omitempty"`
	Language             string                 `json:"language,omitempty"`
	CompiledSystemPrompt string                 `json:"compiled_system_prompt,omitempty"`
	InteractionRole      string                 `json:"interaction_role,omitempty"`
	Timezone             string                 `json:"timezone,omitempty"`
	ToolCapabilities     *AgentToolCapabilities `json:"tool_capabilities,omitempty"`
	ToolDefinitions      []ToolDefinition       `json:"tool_definitions,omitempty"`
	MaxTurns             int                    `json:"max_turns,omitempty"`
	SkipContextBuild     bool                   `json:"skip_context_build,omitempty"`
	GameContext          *GameContext            `json:"game_context,omitempty"`
	Capabilities         []string               `json:"capabilities,omitempty"`
	SkillLevels          map[string]int32        `json:"skill_levels,omitempty"`
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
	FactID               string                 `json:"fact_id"`
	AgentID              string                 `json:"agent_id"`
	UserID               string                 `json:"user_id"`
	NodeID               string                 `json:"node_id"`
	AtomicText           string                 `json:"atomic_text"`
	FactType             string                 `json:"fact_type"`
	Importance           float64                `json:"importance"`
	Confidence           float64                `json:"confidence,omitempty"`
	SupersedesID         string                 `json:"supersedes_id"`
	SessionID            string                 `json:"session_id"`
	SourceID             string                 `json:"source_id,omitempty"`
	SourceType           string                 `json:"source_type,omitempty"`
	Sentiment            string                 `json:"sentiment,omitempty"`
	Entities             []string               `json:"entities,omitempty"`
	InferredEntities     []string               `json:"inferred_entities,omitempty"`
	TopicTags            []string               `json:"topic_tags,omitempty"`
	AgentFraming         string                 `json:"agent_framing,omitempty"`
	CharacterSalience    float64                `json:"character_salience,omitempty"`
	EmotionalIntensity   float64                `json:"emotional_intensity,omitempty"`
	RelationshipRelevance float64               `json:"relationship_relevance,omitempty"`
	RetentionStrength    float64                `json:"retention_strength,omitempty"`
	TemporalRelevance    string                 `json:"temporal_relevance,omitempty"`
	TimeSensitiveAt      string                 `json:"time_sensitive_at,omitempty"`
	EpisodeID            string                 `json:"episode_id,omitempty"`
	EventTime            string                 `json:"event_time,omitempty"`
	EvidenceMessageIDs   []string               `json:"evidence_message_ids,omitempty"`
	PolarityGroupID      string                 `json:"polarity_group_id,omitempty"`
	HitCount             int                    `json:"hit_count,omitempty"`
	MissCount            int                    `json:"miss_count,omitempty"`
	MentionCount         int                    `json:"mention_count,omitempty"`
	LastConfirmed        string                 `json:"last_confirmed,omitempty"`
	LastRetrievedAt      string                 `json:"last_retrieved_at,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt            string                 `json:"created_at,omitempty"`
	UpdatedAt            string                 `json:"updated_at,omitempty"`
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

// PersonalityDimensions contains BFAS personality aspect scores (0-100 scale).
type PersonalityDimensions struct {
	Intellect       float64 `json:"intellect"`
	Aesthetic       float64 `json:"aesthetic"`
	Industriousness float64 `json:"industriousness"`
	Orderliness     float64 `json:"orderliness"`
	Enthusiasm      float64 `json:"enthusiasm"`
	Assertiveness   float64 `json:"assertiveness"`
	Compassion      float64 `json:"compassion"`
	Politeness      float64 `json:"politeness"`
	Withdrawal      float64 `json:"withdrawal"`
	Volatility      float64 `json:"volatility"`
}

// PersonalityPreferences represents interaction preferences.
type PersonalityPreferences struct {
	ConversationPace    string `json:"conversation_pace"`
	Formality           string `json:"formality"`
	HumorStyle          string `json:"humor_style"`
	EmotionalExpression string `json:"emotional_expression"`
}

// PersonalityBehaviors represents behavioral traits.
type PersonalityBehaviors struct {
	ResponseLength    string `json:"response_length"`
	QuestionFrequency string `json:"question_frequency"`
	EmpathyStyle      string `json:"empathy_style"`
	ConflictApproach  string `json:"conflict_approach"`
}

// TraitPrecision represents the precision and observation count for a personality trait.
type TraitPrecision struct {
	Precision        float64 `json:"precision"`
	ObservationCount int     `json:"observation_count"`
	LastUpdatedAt    string  `json:"last_updated_at,omitempty"`
}

// PersonalityProfile represents the full personality profile.
type PersonalityProfile struct {
	AgentID             string                    `json:"agent_id"`
	Name                string                    `json:"name"`
	Gender              string                    `json:"gender"`
	Bio                 string                    `json:"bio"`
	AvatarURL           string                    `json:"avatar_url"`
	PersonalityPrompt   string                    `json:"personality_prompt"`
	SpeechPatterns      []string                  `json:"speech_patterns"`
	TrueInterests       []string                  `json:"true_interests"`
	TrueDislikes        []string                  `json:"true_dislikes"`
	PrimaryTraits       []string                  `json:"primary_traits"`
	Temperature         float64                   `json:"temperature"`
	Big5                Big5                      `json:"big5"`
	Dimensions          PersonalityDimensions     `json:"dimensions"`
	Preferences         PersonalityPreferences    `json:"preferences"`
	Behaviors           PersonalityBehaviors      `json:"behaviors"`
	EmotionalTendencies map[string]float64        `json:"emotional_tendencies"`
	TraitPrecisions     map[string]TraitPrecision `json:"trait_precisions,omitempty"`
	CreatedAt           string                    `json:"created_at,omitempty"`
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
	FactType string // "relationship", "preference", "event", "interest"
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
	// Messages carries the raw conversation that triggered this event (e.g. the
	// chat session that just ended). When present, Platform API uses these as
	// the LLM's conversation context for context-sensitive generation (diary,
	// summaries) instead of reconstructing from lossy consolidation summaries.
	// Typically set by the orchestrator after a chat session ends; omit when
	// triggering from cron jobs or other non-conversation sources. Older
	// Platform API servers that don't know this field will ignore it.
	Messages []ChatMessage `json:"messages,omitempty"`
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
	ID               string `json:"id"`
	AgentID          string `json:"agent_id,omitempty"`
	TenantID         string `json:"tenant_id,omitempty"`
	Name             string `json:"name"`
	Bio              string `json:"bio,omitempty"`
	Gender           string `json:"gender,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	Status           string `json:"status,omitempty"`
	IsActive         bool   `json:"is_active,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`
	InstanceCount    int    `json:"instance_count,omitempty"`
	LastSeenAt       string `json:"last_seen_at,omitempty"`
	OwnerUserID      string `json:"owner_user_id,omitempty"`
	OwnerDisplayName string `json:"owner_display_name,omitempty"`
	OwnerEmail       string `json:"owner_email,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

// AgentListResponse is the response from listing agents.
type AgentListResponse struct {
	Items      []AgentIndex `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
	TotalCount int          `json:"total_count,omitempty"`
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

// PendingCapability represents a capability the agent knows is coming.
type PendingCapability struct {
	Capability string `json:"capability"`
	Context    string `json:"context,omitempty"`
}

// AgentCapabilities represents the capabilities enabled for an agent.
type AgentCapabilities struct {
	WebSearch           bool                   `json:"webSearch"`
	RememberName        bool                   `json:"rememberName"`
	ImageGeneration     bool                   `json:"imageGeneration"`
	Inventory           bool                   `json:"inventory"`
	KnowledgeBase       bool                   `json:"knowledgeBase,omitempty"`
	KnowledgeBaseProjectID string              `json:"knowledgeBaseProjectId,omitempty"`
	VoiceGeneration     bool                   `json:"voiceGeneration"`
	VoiceID             string                 `json:"voiceId,omitempty"`
	VoiceTier           int                    `json:"voiceTier,omitempty"`
	VoiceUnlockedAt     string                 `json:"voiceUnlockedAt,omitempty"`
	ImageUnlockedAt     string                 `json:"imageUnlockedAt,omitempty"`
	MusicGeneration     bool                   `json:"musicGeneration"`
	MusicUnlockedAt     string                 `json:"musicUnlockedAt,omitempty"`
	VideoGeneration     bool                   `json:"videoGeneration"`
	VideoUnlockedAt     string                 `json:"videoUnlockedAt,omitempty"`
	PendingCapabilities []PendingCapability    `json:"pendingCapabilities,omitempty"`
	CustomTools         []CustomToolDefinition `json:"customTools,omitempty"`
}

// UpdateCapabilitiesOptions configures a capabilities update request.
type UpdateCapabilitiesOptions struct {
	WebSearch       *bool `json:"webSearch,omitempty"`
	RememberName    *bool `json:"rememberName,omitempty"`
	ImageGeneration *bool `json:"imageGeneration,omitempty"`
	Inventory       *bool `json:"inventory,omitempty"`
	KnowledgeBase   *bool `json:"knowledgeBase,omitempty"`
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
// Constellation (knowledge graph)
// ---------------------------------------------------------------------------

// ConstellationNode represents a node in the knowledge graph.
type ConstellationNode struct {
	NodeID       string                 `json:"node_id"`
	AgentID      string                 `json:"agent_id"`
	UserID       string                 `json:"user_id,omitempty"`
	Label        string                 `json:"label"`
	Type         string                 `json:"type"`
	NodeType     string                 `json:"node_type,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Significance float64                `json:"significance,omitempty"`
	Weight       float64                `json:"weight"`
	MentionCount int                    `json:"mention_count,omitempty"`
	Brightness   float64                `json:"brightness,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
}

// ConstellationEdge represents an edge in the knowledge graph.
type ConstellationEdge struct {
	EdgeID   string                 `json:"edge_id"`
	AgentID  string                 `json:"agent_id"`
	SourceID string                 `json:"source_id"`
	TargetID string                 `json:"target_id"`
	Relation string                 `json:"relation"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConstellationInsight represents an insight derived from the knowledge graph.
type ConstellationInsight struct {
	InsightID string                 `json:"insight_id"`
	AgentID   string                 `json:"agent_id"`
	UserID    string                 `json:"user_id,omitempty"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Surfaced  bool                   `json:"surfaced"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
}

// ConstellationResponse is the response from the constellation endpoint.
type ConstellationResponse struct {
	Nodes    []ConstellationNode    `json:"nodes"`
	Edges    []ConstellationEdge    `json:"edges"`
	Insights []ConstellationInsight `json:"insights"`
}

// ---------------------------------------------------------------------------
// Breakthroughs
// ---------------------------------------------------------------------------

// Breakthrough represents a relationship breakthrough event.
type Breakthrough struct {
	BreakthroughID      string   `json:"breakthrough_id"`
	AgentID             string   `json:"agent_id"`
	UserID              string   `json:"user_id"`
	BreakthroughNumber  int      `json:"breakthrough_number"`
	LevelAtBreakthrough int      `json:"level_at_breakthrough"`
	Narrative           string   `json:"narrative"`
	PersonalityShifts   []string `json:"personality_shifts"`
	TraitEvolved        string   `json:"trait_evolved,omitempty"`
	NewGoals            []string `json:"new_goals"`
	AchievedGoals       []string `json:"achieved_goals"`
	SkillPointsAwarded  int      `json:"skill_points_awarded"`
	Acknowledged        bool     `json:"acknowledged"`
	CreatedAt           string   `json:"created_at"`
}

// BreakthroughsResponse is the response from the breakthroughs endpoint.
type BreakthroughsResponse struct {
	Breakthroughs []Breakthrough `json:"breakthroughs"`
}

// ---------------------------------------------------------------------------
// Wakeups (list)
// ---------------------------------------------------------------------------

// WakeupListOptions configures a list wakeups request.
type WakeupListOptions struct {
	UserID string // Filter by user (optional; empty means all users on the agent).
	Limit  int    // Maximum number of wakeups to return (default 50, max 500).
	Status string // Filter by status (e.g. "pending", "executed").
}

// WakeupsResponse is the response from the list wakeups endpoint.
type WakeupsResponse struct {
	Wakeups []ScheduledWakeup `json:"wakeups"`
}

// ---------------------------------------------------------------------------
// Process (full pipeline)
// ---------------------------------------------------------------------------

// ProcessOptions configures a process request.
type ProcessOptions struct {
	UserID             string        `json:"userId"`
	SessionID          string        `json:"sessionId,omitempty"`
	InstanceID         string        `json:"instanceId,omitempty"`
	Messages           []ChatMessage `json:"messages"`
	Provider           string        `json:"provider,omitempty"`
	Model              string        `json:"model,omitempty"`
	IncludeExtractions bool          `json:"include_extractions,omitempty"`
}

// ProcessSideEffectsSummary summarises behavioral side effects.
type ProcessSideEffectsSummary struct {
	MoodUpdated        bool `json:"mood_updated"`
	PersonalityUpdated bool `json:"personality_updated"`
	HabitsObserved     int  `json:"habits_observed"`
	InterestsDetected  int  `json:"interests_detected"`
}

// ProcessResponse is the response from the process endpoint.
type ProcessResponse struct {
	Success         bool                      `json:"success"`
	MemoriesCreated int                       `json:"memories_created"`
	FactsExtracted  int                       `json:"facts_extracted"`
	SideEffects     ProcessSideEffectsSummary `json:"side_effects"`
	Extractions     *SideEffectExtraction     `json:"extractions,omitempty"`
}

// ---------------------------------------------------------------------------
// Side-Effect Extraction Types
// ---------------------------------------------------------------------------

// SideEffectExtraction contains the full extracted side effects from conversation processing.
type SideEffectExtraction struct {
	MemoryFacts          []ExtractionFact         `json:"memory_facts"`
	PersonalityDeltas    []ExtractionPDelta       `json:"personality_deltas"`
	DimensionDeltas      []ExtractionDDelta       `json:"dimension_deltas"`
	MoodDelta            *ExtractionMoodDelta     `json:"mood_delta"`
	HabitObservations    []ExtractionHabit        `json:"habit_observations"`
	InterestsDetected    []ExtractionInterest     `json:"interests_detected"`
	RelationshipDelta    *ExtractionRelDelta      `json:"relationship_delta"`
	ProactiveSuggestions []ExtractionProactive    `json:"proactive_suggestions"`
	RecurringEvents      []ExtractionRecurring    `json:"recurring_events"`
	InnerThoughts        *ExtractionInnerThoughts `json:"inner_thoughts"`
	EmotionalThemes      []string                 `json:"emotional_themes"`
}

// ExtractionFact represents a single extracted memory fact.
type ExtractionFact struct {
	Text       string   `json:"text"`
	FactType   string   `json:"fact_type"`
	Importance float64  `json:"importance"`
	Entities   []string `json:"entities"`
	Sentiment  string   `json:"sentiment"`
	TopicTags  []string `json:"topic_tags"`
}

// ExtractionPDelta represents a personality trait delta.
type ExtractionPDelta struct {
	Trait  string  `json:"trait"`
	Delta  float64 `json:"delta"`
	Reason string  `json:"reason"`
}

// ExtractionDDelta represents a personality dimension delta.
type ExtractionDDelta struct {
	Dimension string  `json:"dimension"`
	Delta     float64 `json:"delta"`
	Reason    string  `json:"reason"`
}

// ExtractionMoodDelta represents a mood state delta.
type ExtractionMoodDelta struct {
	Happiness float64 `json:"happiness"`
	Energy    float64 `json:"energy"`
	Calmness  float64 `json:"calmness"`
	Affection float64 `json:"affection"`
	Reason    string  `json:"reason"`
}

// ExtractionHabit represents an observed habit.
type ExtractionHabit struct {
	Name            string `json:"name"`
	Category        string `json:"category"`
	Description     string `json:"description"`
	IsReinforcement bool   `json:"is_reinforcement"`
}

// ExtractionInterest represents a detected interest.
type ExtractionInterest struct {
	Topic           string  `json:"topic"`
	Category        string  `json:"category"`
	Confidence      float64 `json:"confidence"`
	EngagementLevel float64 `json:"engagement_level"`
}

// ExtractionRelDelta represents a relationship score delta.
type ExtractionRelDelta struct {
	ScoreChange int    `json:"score_change"`
	Reason      string `json:"reason"`
}

// ExtractionProactive represents a proactive suggestion.
type ExtractionProactive struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	DelayHours  int    `json:"delay_hours"`
	Intent      string `json:"intent"`
}

// ExtractionRecurring represents a detected recurring event.
type ExtractionRecurring struct {
	Description string  `json:"description"`
	Pattern     string  `json:"pattern"`
	Confidence  float64 `json:"confidence"`
}

// ExtractionInnerThoughts represents the agent's inner thoughts.
type ExtractionInnerThoughts struct {
	Diary      string `json:"diary"`
	Reflection string `json:"reflection"`
}

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

// ModelVariant represents a single model variant offered by a provider.
type ModelVariant struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ModelsProviderEntry represents a single LLM provider.
type ModelsProviderEntry struct {
	Provider     string         `json:"provider"`
	ProviderName string         `json:"provider_name"`
	DefaultModel string         `json:"default_model"`
	Models       []ModelVariant `json:"models"`
}

// ModelsResponse is the response from the models endpoint.
type ModelsResponse struct {
	DefaultProvider string                `json:"default_provider"`
	DefaultModel    string                `json:"default_model"`
	Providers       []ModelsProviderEntry `json:"providers"`
}

// PlatformModelsResponse is the response from GET /api/v1/models.
// It lists all LLM providers and model variants enabled on this deployment.
type PlatformModelsResponse struct {
	DefaultModel string                `json:"default_model"`
	Providers    []ModelsProviderEntry `json:"providers"`
}

// ---------------------------------------------------------------------------
// Context (single-call enriched context)
// ---------------------------------------------------------------------------

// GetContextOptions configures a get context request.
type GetContextOptions struct {
	UserID     string `json:"user_id"`
	SessionID  string `json:"session_id,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
	Query      string `json:"query,omitempty"`
	Language   string `json:"language,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

// EnrichedContextResponse is the response from the context endpoint.
// Fields are intentionally loosely typed for forward compatibility.
type EnrichedContextResponse struct {
	// Layer 1: Core Identity
	Bio               string   `json:"bio,omitempty"`
	PersonalityPrompt string   `json:"personality_prompt,omitempty"`
	SpeechPatterns    []string `json:"speech_patterns,omitempty"`
	TrueInterests     []string `json:"true_interests,omitempty"`
	TrueDislikes      []string `json:"true_dislikes,omitempty"`
	PrimaryTraits     []string `json:"primary_traits,omitempty"`

	// Layer 2: Personality
	Big5        map[string]interface{} `json:"big5,omitempty"`
	Dimensions  map[string]interface{} `json:"dimensions,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	Behaviors   map[string]interface{} `json:"behaviors,omitempty"`

	// Layer 3: Evolution
	RecentPersonalityShifts []interface{} `json:"recent_personality_shifts,omitempty"`
	SignificantMoments      []interface{} `json:"significant_moments,omitempty"`
	ActiveGoals             []interface{} `json:"active_goals,omitempty"`
	Habits                  []interface{} `json:"habits,omitempty"`
	BreakthroughCount       int           `json:"breakthrough_count,omitempty"`

	// Layer 4: Relationship
	RelationshipNarrative string  `json:"relationship_narrative,omitempty"`
	SharedMemorySummary   string  `json:"shared_memory_summary,omitempty"`
	ChemistryScore        float64 `json:"chemistry_score,omitempty"`
	LoveFromAgent         float64 `json:"love_from_agent,omitempty"`
	LoveFromUser          float64 `json:"love_from_user,omitempty"`
	RelationshipStatus    string  `json:"relationship_status,omitempty"`
	DaysSinceLastChat     int     `json:"days_since_last_chat,omitempty"`

	// Layer 5: Current State
	CurrentMood    map[string]interface{} `json:"current_mood,omitempty"`
	EmotionalState string                 `json:"emotional_state,omitempty"`
	Capabilities   map[string]interface{} `json:"capabilities,omitempty"`

	// Layer 6: Memory
	LoadedFacts       []map[string]interface{} `json:"loaded_facts,omitempty"`
	LongTermSummaries []interface{}            `json:"long_term_summaries,omitempty"`

	// Layer 6b: Proactive
	ProactiveMemories []interface{} `json:"proactive_memories,omitempty"`

	// Layer 6c: Constellation
	ConstellationPatterns []interface{} `json:"constellation_patterns,omitempty"`

	// Layer 7: Backend Context
	BackendContext map[string]interface{} `json:"game_context,omitempty"`
}

// ---------------------------------------------------------------------------
// Avatar Generation
// ---------------------------------------------------------------------------

// GenerateAvatarOptions configures an avatar generation request.
type GenerateAvatarOptions struct {
	Style string `json:"style,omitempty"`
}

// GenerateAvatarResponse is the response from the avatar generation endpoint.
type GenerateAvatarResponse struct {
	Success          bool   `json:"success"`
	AvatarURL        string `json:"avatar_url"`
	Prompt           string `json:"prompt"`
	GenerationTimeMs int    `json:"generation_time_ms"`
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

// ---------------------------------------------------------------------------
// Mood
// ---------------------------------------------------------------------------

// MoodState represents agent mood dimensions (valence-arousal-tension-affiliation model).
type MoodState struct {
	Valence     float64 `json:"valence"`
	Arousal     float64 `json:"arousal"`
	Tension     float64 `json:"tension"`
	Affiliation float64 `json:"affiliation"`
	Label       string  `json:"label,omitempty"`
}

// MoodResponse wraps the mood state from the API.
type MoodResponse struct {
	Mood MoodState `json:"mood"`
}

// MoodHistoryEntry represents a single mood history data point.
type MoodHistoryEntry struct {
	Valence          float64 `json:"valence"`
	Arousal          float64 `json:"arousal"`
	Tension          float64 `json:"tension"`
	Affiliation      float64 `json:"affiliation"`
	Label            string  `json:"label,omitempty"`
	TriggerType      string  `json:"trigger_type,omitempty"`
	TriggerReason    string  `json:"trigger_reason,omitempty"`
	DeltaValence     float64 `json:"delta_valence,omitempty"`
	DeltaArousal     float64 `json:"delta_arousal,omitempty"`
	DeltaTension     float64 `json:"delta_tension,omitempty"`
	DeltaAffiliation float64 `json:"delta_affiliation,omitempty"`
	Timestamp        string  `json:"timestamp"`
}

// MoodHistoryResponse wraps mood history from the API.
type MoodHistoryResponse struct {
	Entries []MoodHistoryEntry `json:"entries"`
}

// MoodAggregateResponse wraps aggregated mood statistics.
type MoodAggregateResponse struct {
	Valence     float64 `json:"valence"`
	Arousal     float64 `json:"arousal"`
	Tension     float64 `json:"tension"`
	Affiliation float64 `json:"affiliation"`
	Label       string  `json:"label"`
	UserCount   int     `json:"user_count"`
	DaysWindow  int     `json:"days_window"`
}

// ---------------------------------------------------------------------------
// Relationships
// ---------------------------------------------------------------------------

// RelationshipData represents a relationship with a user.
type RelationshipData struct {
	UserID         string  `json:"user_id"`
	LoveScore      float64 `json:"love_score"`
	ChemistryScore float64 `json:"chemistry_score,omitempty"`
	Narrative      string  `json:"narrative,omitempty"`
	LastUpdate     string  `json:"last_update,omitempty"`
	UpdatedAt      string  `json:"updated_at,omitempty"`
}

// RelationshipsResponse wraps relationship data.
type RelationshipsResponse struct {
	Relationships []RelationshipData `json:"relationships"`
}

// ---------------------------------------------------------------------------
// Habits
// ---------------------------------------------------------------------------

// HabitData represents an agent habit.
type HabitData struct {
	Name            string  `json:"name"`
	Strength        float64 `json:"strength"`
	Category        string  `json:"category,omitempty"`
	Description     string  `json:"description,omitempty"`
	DisplayName     string  `json:"display_name,omitempty"`
	Formed          bool    `json:"formed,omitempty"`
	DailyReinforced float64 `json:"daily_reinforced,omitempty"`
	LastUpdate      string  `json:"last_update,omitempty"`
}

// HabitsResponse wraps habit data.
type HabitsResponse struct {
	Habits []HabitData `json:"habits"`
}

// ---------------------------------------------------------------------------
// Interests
// ---------------------------------------------------------------------------

// InterestData represents an agent interest.
type InterestData struct {
	Topic            string  `json:"topic"`
	Score            float64 `json:"score,omitempty"` // Deprecated: use Confidence instead.
	Category         string  `json:"category,omitempty"`
	AgentID          string  `json:"agent_id,omitempty"`
	UserID           string  `json:"user_id,omitempty"`
	Confidence       float64 `json:"confidence,omitempty"`
	EngagementLevel  float64 `json:"engagement_level,omitempty"`
	MentionCount     int     `json:"mention_count,omitempty"`
	ResearchStatus   string  `json:"research_status,omitempty"`
	ResearchFindings string  `json:"research_findings,omitempty"`
	LastMentionedAt  string  `json:"last_mentioned_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// InterestsResponse wraps interest data.
type InterestsResponse struct {
	Interests []InterestData `json:"interests"`
}

// ---------------------------------------------------------------------------
// Diary
// ---------------------------------------------------------------------------

// DiaryEntry represents a single diary entry.
type DiaryEntry struct {
	EntryID     string   `json:"entry_id"`
	AgentID     string   `json:"agent_id"`
	UserID      string   `json:"user_id,omitempty"`
	Date        string   `json:"date"`
	Content     string   `json:"content"`
	Title       string   `json:"title,omitempty"`
	BodyLines   []string `json:"body_lines,omitempty"`
	Body        string   `json:"body,omitempty"` // Deprecated: use Content instead.
	Mood        string   `json:"mood,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	TriggerType string   `json:"trigger_type,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// DiaryResponse wraps diary entries.
type DiaryResponse struct {
	Entries []DiaryEntry `json:"entries"`
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

// UsersResponse wraps agent users data.
type UsersResponse struct {
	Users []map[string]interface{} `json:"users"`
}

// ---------------------------------------------------------------------------
// Agent Knowledge Search (tool endpoint)
// ---------------------------------------------------------------------------

// ForkAgentOptions configures a fork agent request.
type ForkAgentOptions struct {
	Name *string `json:"name,omitempty"`
}

// ForkResponse is the response from forking an agent.
type ForkResponse struct {
	AgentID       string `json:"agent_id"`
	SourceAgentID string `json:"source_agent_id"`
	Status        string `json:"status"`
	Name          string `json:"name"`
}

// ForkStatusResponse is the response from checking fork status.
type ForkStatusResponse struct {
	Status        string  `json:"status"`
	SourceAgentID string  `json:"source_agent_id"`
	StartedAt     *string `json:"started_at,omitempty"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	TablesCopied  int     `json:"tables_copied"`
	TablesTotal   int     `json:"tables_total"`
	ErrorMessage  string  `json:"error_message,omitempty"`
}

// DeleteWisdomResponse is the response from deleting a wisdom fact.
type DeleteWisdomResponse struct {
	Success bool   `json:"success"`
	FactID  string `json:"fact_id"`
}

// WisdomAuditResponse is the response from the wisdom audit endpoint.
type WisdomAuditResponse struct {
	FactID              string   `json:"fact_id"`
	Content             string   `json:"content"`
	TargetPath          string   `json:"target_path,omitempty"`
	DerivedFromHashes   []string `json:"derived_from_hashes,omitempty"`
	SourceUserCount     int      `json:"source_user_count"`
	PromotionConfidence float64  `json:"promotion_confidence"`
	PromotedAt          string   `json:"promoted_at,omitempty"`
}

// AgentKBSearchOptions configures an agent-scoped knowledge search request.
type AgentKBSearchOptions struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// AgentKBSearchResult represents a single result from agent knowledge search.
type AgentKBSearchResult struct {
	Content  string  `json:"content"`
	Label    string  `json:"label"`
	NodeType string  `json:"type"`
	Source   string  `json:"source"`
	Score    float64 `json:"score"`
}

// AgentKBSearchResponse is the response from the agent knowledge search endpoint.
type AgentKBSearchResponse struct {
	Query   string                `json:"query"`
	Results []AgentKBSearchResult `json:"results"`
}

// ---------------------------------------------------------------------------
// Tool Schemas (BYO-LLM)
// ---------------------------------------------------------------------------

// ToolSchema describes a single tool available for an agent (BYO-LLM integrations).
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Endpoint    string         `json:"endpoint"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// ToolSchemasResponse is the response from the GetTools endpoint.
type ToolSchemasResponse struct {
	Tools []ToolSchema `json:"tools"`
}
