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
// Agent CRUD
// ---------------------------------------------------------------------------

// Big5Scores represents the Big Five personality trait scores (0.0-1.0).
type Big5Scores struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
	Confidence        float64 `json:"confidence,omitempty"`
}

// SDKPersonalityDimensions contains BFAS personality aspect scores (0-100 scale).
type SDKPersonalityDimensions struct {
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

// AgentToolCapabilities specifies which built-in Sonzai tools to enable.
type AgentToolCapabilities struct {
	WebSearch       bool `json:"web_search"`
	RememberName    bool `json:"remember_name"`
	ImageGeneration bool `json:"image_generation"`
}

// CreateAgentParams contains the parameters for creating a new agent.
type CreateAgentParams struct {
	AgentID           string                    `json:"agent_id,omitempty"`
	UserID            string                    `json:"user_id"`
	AgentName         string                    `json:"agent_name"`
	Gender            string                    `json:"gender"`
	Bio               string                    `json:"bio,omitempty"`
	AvatarURL         string                    `json:"avatar_url,omitempty"`
	Big5              Big5Scores                `json:"big5"`
	Language          string                    `json:"language,omitempty"`
	ModelTier         int                       `json:"model_tier,omitempty"`
	PersonalityPrompt string                    `json:"personality_prompt,omitempty"`
	SpeechPatterns    []string                  `json:"speech_patterns,omitempty"`
	TrueInterests     []string                  `json:"true_interests,omitempty"`
	TrueDislikes      []string                  `json:"true_dislikes,omitempty"`
	UserDisplayName   string                    `json:"user_display_name,omitempty"`
	Dimensions        *SDKPersonalityDimensions `json:"dimensions,omitempty"`
	ToolCapabilities  *AgentToolCapabilities    `json:"tool_capabilities,omitempty"`
}

// CreateAgentResult contains the result of agent creation.
type CreateAgentResult struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

// AgentProfile contains the retrieved agent profile.
type AgentProfile struct {
	AgentID     string     `json:"agent_id"`
	Name        string     `json:"name"`
	Bio         string     `json:"bio"`
	Gender      string     `json:"gender"`
	AvatarURL   string     `json:"avatar_url"`
	Big5        Big5Scores `json:"big5"`
	OwnerUserID string     `json:"owner_user_id"`
	CreatedAt   string     `json:"created_at,omitempty"`
}

// UpdateAgentParams contains the parameters for updating an agent.
type UpdateAgentParams struct {
	Name             string                    `json:"name,omitempty"`
	Bio              string                    `json:"bio,omitempty"`
	AvatarURL        string                    `json:"avatar_url,omitempty"`
	Big5             *Big5Scores               `json:"big5,omitempty"`
	Dimensions       *SDKPersonalityDimensions `json:"dimensions,omitempty"`
	ToolCapabilities *AgentToolCapabilities    `json:"tool_capabilities,omitempty"`
}

// UpdateAgentResult contains the result of an agent update.
type UpdateAgentResult struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------
// Memory Seed / Facts / Reset
// ---------------------------------------------------------------------------

// MemoryCandidate represents a candidate memory to store.
type MemoryCandidate struct {
	Content    string   `json:"content,omitempty"`
	FactType   string   `json:"fact_type,omitempty"`
	Importance float64  `json:"importance,omitempty"`
	Entities   []string `json:"entities,omitempty"`
}

// SeedMemoriesParams contains the parameters for seeding initial memories.
type SeedMemoriesParams struct {
	UserID     string            `json:"user_id"`
	Memories   []MemoryCandidate `json:"memories"`
	InstanceID string            `json:"instance_id,omitempty"`
}

// SeedMemoriesResult contains the result of memory seeding.
type SeedMemoriesResult struct {
	MemoriesCreated int `json:"memories_created"`
}

// ListFactsOptions configures a list facts request.
type ListFactsOptions struct {
	UserID     string
	Limit      int
	FactType   string
	InstanceID string
}

// StoredFact represents a fact stored in the platform's memory system.
type StoredFact struct {
	FactID       string  `json:"fact_id"`
	Content      string  `json:"content"`
	FactType     string  `json:"fact_type"`
	Importance   float64 `json:"importance"`
	Confidence   float64 `json:"confidence"`
	Entity       string  `json:"entity"`
	SourceType   string  `json:"source_type"`
	MentionCount int     `json:"mention_count"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
}

// ListFactsResult contains the result of listing facts.
type ListFactsResult struct {
	Facts      []StoredFact `json:"facts"`
	TotalCount int          `json:"total_count"`
}

// ResetMemoryParams contains the parameters for resetting an agent's memories.
type ResetMemoryParams struct {
	UserID     string `json:"user_id"`
	InstanceID string `json:"instance_id,omitempty"`
}

// ResetMemoryResult contains the result of a memory reset.
type ResetMemoryResult struct {
	Success      bool `json:"success"`
	FactsDeleted int  `json:"facts_deleted"`
	NodesDeleted int  `json:"nodes_deleted"`
}

// ---------------------------------------------------------------------------
// Personality Update
// ---------------------------------------------------------------------------

// UpdatePersonalityParams contains the parameters for updating personality.
type UpdatePersonalityParams struct {
	Big5       Big5Scores                `json:"big5"`
	Dimensions *SDKPersonalityDimensions `json:"dimensions,omitempty"`
}

// UpdatePersonalityResult contains the result of a personality update.
type UpdatePersonalityResult struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------
// Voice
// ---------------------------------------------------------------------------

// EmotionalContext provides emotional hints for TTS generation.
type EmotionalContext struct {
	Themes []string `json:"themes,omitempty"`
	Tone   string   `json:"tone,omitempty"`
}

// TextToSpeechParams contains the parameters for text-to-speech.
type TextToSpeechParams struct {
	AgentID          string            `json:"agent_id"`
	Text             string            `json:"text"`
	VoiceName        string            `json:"voice_name,omitempty"`
	Language         string            `json:"language,omitempty"`
	EmotionalContext *EmotionalContext `json:"emotional_context,omitempty"`
}

// TextToSpeechResult contains the TTS result.
type TextToSpeechResult struct {
	Audio       []byte `json:"audio"`
	ContentType string `json:"content_type"`
	VoiceName   string `json:"voice_name"`
	DurationMs  int    `json:"duration_ms,omitempty"`
}

// VoiceMatchParams contains the parameters for voice matching.
type VoiceMatchParams struct {
	AgentID         string     `json:"agent_id,omitempty"`
	Big5            Big5Scores `json:"big5"`
	PreferredGender string     `json:"preferred_gender,omitempty"`
}

// VoiceMatchResult contains the voice match result.
type VoiceMatchResult struct {
	VoiceID    string  `json:"voice_id"`
	VoiceName  string  `json:"voice_name"`
	MatchScore float64 `json:"match_score"`
	Reasoning  string  `json:"reasoning"`
}

// VoiceInfo represents a single available voice.
type VoiceInfo struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

// ListVoicesResult contains the available voices.
type ListVoicesResult struct {
	Voices []VoiceInfo `json:"voices"`
}

// VoiceChatParams contains the parameters for a single voice chat turn.
type VoiceChatParams struct {
	AgentID     string `json:"agent_id"`
	UserID      string `json:"user_id"`
	Audio       []byte `json:"audio"`
	AudioFormat string `json:"audio_format"`
	VoiceName   string `json:"voice_name,omitempty"`
	Language    string `json:"language,omitempty"`
	InstanceID  string `json:"instance_id,omitempty"`
}

// VoiceChatResult contains the voice chat result.
type VoiceChatResult struct {
	Transcript      string          `json:"transcript"`
	Response        string          `json:"response"`
	Audio           []byte          `json:"audio"`
	ContentType     string          `json:"content_type"`
	SideEffectsJSON json.RawMessage `json:"side_effects_json,omitempty"`
}

// ---------------------------------------------------------------------------
// Content Generation
// ---------------------------------------------------------------------------

// GenerateBioParams contains the parameters for bio generation.
type GenerateBioParams struct {
	UserID     string `json:"user_id,omitempty"`
	CurrentBio string `json:"current_bio,omitempty"`
	Style      string `json:"style,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
}

// GenerateBioResult contains the generated bio.
type GenerateBioResult struct {
	Bio        string  `json:"bio"`
	Tone       string  `json:"tone"`
	Confidence float64 `json:"confidence"`
}

// GenerateImageParams contains the parameters for image generation.
type GenerateImageParams struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Model          string `json:"model,omitempty"`
	Provider       string `json:"provider,omitempty"`
}

// GenerateImageResult contains the generated image details.
type GenerateImageResult struct {
	Success          bool   `json:"success"`
	ImageID          string `json:"image_id"`
	PublicURL        string `json:"public_url"`
	MimeType         string `json:"mime_type"`
	GenerationTimeMs int64  `json:"generation_time_ms"`
	Error            string `json:"error,omitempty"`
}

// GenerateCharacterParams contains the parameters for full character generation.
type GenerateCharacterParams struct {
	UserID      string   `json:"user_id,omitempty"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender"`
	Description string   `json:"description"`
	Fields      []string `json:"fields,omitempty"`
}

// SDKInteractionPreferences contains conversation style preferences.
type SDKInteractionPreferences struct {
	ConversationPace    string `json:"conversation_pace"`
	Formality           string `json:"formality"`
	HumorStyle          string `json:"humor_style"`
	EmotionalExpression string `json:"emotional_expression"`
}

// SDKBehavioralTraits contains behavioral response patterns.
type SDKBehavioralTraits struct {
	ResponseLength    string `json:"response_length"`
	QuestionFrequency string `json:"question_frequency"`
	EmpathyStyle      string `json:"empathy_style"`
	ConflictApproach  string `json:"conflict_approach"`
}

// GenerateCharacterUsage contains token usage for generation.
type GenerateCharacterUsage struct {
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	Model            string `json:"model"`
}

// GenerateCharacterResult contains the generated character profile.
type GenerateCharacterResult struct {
	Bio               string                    `json:"bio"`
	PersonalityPrompt string                    `json:"personality_prompt"`
	Big5              *Big5Scores               `json:"big5,omitempty"`
	SpeechPatterns    []string                  `json:"speech_patterns,omitempty"`
	TrueInterests     []string                  `json:"true_interests,omitempty"`
	TrueDislikes      []string                  `json:"true_dislikes,omitempty"`
	PrimaryTraits     []string                  `json:"primary_traits,omitempty"`
	Preferences       *SDKInteractionPreferences `json:"preferences,omitempty"`
	Behaviors         *SDKBehavioralTraits       `json:"behaviors,omitempty"`
	Dimensions        *SDKPersonalityDimensions  `json:"dimensions,omitempty"`
	Usage             *GenerateCharacterUsage    `json:"usage,omitempty"`
}

// ModelConfig specifies LLM provider and model settings for generation.
type ModelConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// LoreGenerationContext provides world context for LLM-based lore generation.
type LoreGenerationContext struct {
	WorldDescription         string            `json:"world_description"`
	EntityTerminology        map[string]string `json:"entity_terminology,omitempty"`
	OriginPromptInstructions string            `json:"origin_prompt_instructions,omitempty"`
}

// IdentityMemoryTemplate defines a template for identity memories.
type IdentityMemoryTemplate struct {
	Template   string   `json:"template"`
	FactType   string   `json:"fact_type"`
	Importance float64  `json:"importance"`
	Entities   []string `json:"entities,omitempty"`
}

// GenerateSeedMemoriesParams contains parameters for generating seed memories via LLM.
type GenerateSeedMemoriesParams struct {
	UserID                       string                   `json:"user_id,omitempty"`
	AgentName                    string                   `json:"agent_name"`
	Big5                         Big5Scores               `json:"big5"`
	PersonalityPrompt            string                   `json:"personality_prompt,omitempty"`
	TrueInterests                []string                 `json:"true_interests,omitempty"`
	TrueDislikes                 []string                 `json:"true_dislikes,omitempty"`
	SpeechPatterns               []string                 `json:"speech_patterns,omitempty"`
	CreatorDisplayName           string                   `json:"creator_display_name,omitempty"`
	StaticLoreMemories           []MemoryCandidate        `json:"static_lore_memories,omitempty"`
	LoreGenerationContext        *LoreGenerationContext   `json:"lore_generation_context,omitempty"`
	IdentityMemoryTemplates      []IdentityMemoryTemplate `json:"identity_memory_templates,omitempty"`
	GenerateOriginStory          bool                     `json:"generate_origin_story,omitempty"`
	GeneratePersonalizedMemories bool                     `json:"generate_personalized_memories,omitempty"`
	ModelConfig                  *ModelConfig             `json:"model_config,omitempty"`
	StoreMemories                bool                     `json:"store_memories,omitempty"`
}

// GenerateSeedMemoriesResult contains the result of seed memory generation.
type GenerateSeedMemoriesResult struct {
	Memories       []MemoryCandidate `json:"memories"`
	MemoriesStored int               `json:"memories_stored"`
}

// ---------------------------------------------------------------------------
// Dialogue
// ---------------------------------------------------------------------------

// AgentDialogueParams contains the parameters for agent dialogue generation.
type AgentDialogueParams struct {
	UserID        string        `json:"user_id,omitempty"`
	Messages      []ChatMessage `json:"messages,omitempty"`
	RequestType   string        `json:"request_type,omitempty"`
	SceneGuidance string        `json:"scene_guidance,omitempty"`
	InstanceID    string        `json:"instance_id,omitempty"`
}

// AgentDialogueResult contains the dialogue generation result.
type AgentDialogueResult struct {
	Response        string          `json:"response,omitempty"`
	SideEffectsJSON json.RawMessage `json:"side_effects_json,omitempty"`
}

// ---------------------------------------------------------------------------
// Game Events
// ---------------------------------------------------------------------------

// TriggerGameEventParams contains the parameters for triggering a game event.
type TriggerGameEventParams struct {
	UserID           string            `json:"user_id,omitempty"`
	EventType        string            `json:"event_type"`
	EventDescription string            `json:"event_description,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Language         string            `json:"language,omitempty"`
	InstanceID       string            `json:"instance_id,omitempty"`
}

// TriggerGameEventResult contains the result of a game event trigger.
type TriggerGameEventResult struct {
	Accepted bool   `json:"accepted"`
	EventID  string `json:"event_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Custom States
// ---------------------------------------------------------------------------

// CustomState represents a custom state key-value entry.
type CustomState struct {
	StateID   string          `json:"state_id"`
	AgentID   string          `json:"agent_id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	CreatedAt string          `json:"created_at,omitempty"`
	UpdatedAt string          `json:"updated_at,omitempty"`
}

// CustomStateListResponse is the response from listing custom states.
type CustomStateListResponse struct {
	States []CustomState `json:"states"`
}

// CustomStateCreateParams contains the parameters for creating a custom state.
type CustomStateCreateParams struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// CustomStateUpdateParams contains the parameters for updating a custom state.
type CustomStateUpdateParams struct {
	Value interface{} `json:"value"`
}

// ---------------------------------------------------------------------------
// Constellation / Breakthroughs / Wakeups
// ---------------------------------------------------------------------------

// MoodState represents an agent's emotional state in PAD dimensions.
type MoodState struct {
	Valence     float64 `json:"valence"`
	Arousal     float64 `json:"arousal"`
	Tension     float64 `json:"tension"`
	Affiliation float64 `json:"affiliation"`
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
