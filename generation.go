package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
)

// GenerationResource provides AI content generation operations.
type GenerationResource struct {
	http *httpClient
}

// GenerateBioOptions configures a bio generation request.
type GenerateBioOptions struct {
	Name                string          `json:"name,omitempty"`
	Gender              string          `json:"gender,omitempty"`
	Description         string          `json:"description,omitempty"`
	UserID              string          `json:"user_id,omitempty"`
	EnrichedContextJSON json.RawMessage `json:"enriched_context_json,omitempty"`
	CurrentBio          string          `json:"current_bio,omitempty"`
	Style               string          `json:"style,omitempty"`
	InstanceID          string          `json:"instance_id,omitempty"`
}

// GenerateBioResponse is the response from bio generation.
type GenerateBioResponse struct {
	Bio        string  `json:"bio"`
	Tone       string  `json:"tone,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GenerateCharacterOptions configures a character generation request.
type GenerateCharacterOptions struct {
	// AgentID is an optional UUID. If empty, a deterministic ID is derived from Name.
	AgentID     string   `json:"agent_id,omitempty"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender,omitempty"`
	Description string   `json:"description,omitempty"`
	Fields      []string `json:"fields,omitempty"`
	// Provider selects the LLM backend ("gemini" | "openrouter" | "xai").
	// Leave empty to use the platform default (currently "gemini").
	Provider string `json:"provider,omitempty"`
	// Model optionally overrides the provider's default model.
	Model string `json:"model,omitempty"`
	// Regenerate forces re-generation even if an existing agent profile is found.
	Regenerate bool `json:"regenerate,omitempty"`
}

// GeneratedGoal represents a goal generated as part of character generation.
type GeneratedGoal struct {
	Type        string `json:"type,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority,omitempty"`
}

// GenerateCharacterResponse is the response from character generation.
type GenerateCharacterResponse struct {
	// AgentID is the resolved agent ID (provided or derived from name).
	AgentID string `json:"agent_id,omitempty"`
	// Existing is true when the agent already existed and the LLM was not called.
	Existing                 bool                       `json:"existing,omitempty"`
	Bio                      string                     `json:"bio"`
	PersonalityPrompt        string                     `json:"personality_prompt"`
	Big5                     *Big5Scores                `json:"big5,omitempty"`
	SpeechPatterns           []string                   `json:"speech_patterns,omitempty"`
	TrueInterests            []string                   `json:"true_interests,omitempty"`
	TrueDislikes             []string                   `json:"true_dislikes,omitempty"`
	PrimaryTraits            []string                   `json:"primary_traits,omitempty"`
	Dimensions               *SDKPersonalityDimensions  `json:"dimensions,omitempty"`
	Preferences              *SDKInteractionPreferences `json:"preferences,omitempty"`
	Behaviors                *SDKBehavioralTraits       `json:"behaviors,omitempty"`
	InitialGoals             []GeneratedGoal            `json:"initial_goals,omitempty"`
	WorldDescription         string                     `json:"world_description,omitempty"`
	OriginPromptInstructions string                     `json:"origin_prompt_instructions,omitempty"`
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

// GenerateSeedMemoriesOptions configures a seed memory generation request.
// Field names use camelCase to match the Platform API.
type GenerateSeedMemoriesOptions struct {
	UserID                       string                 `json:"user_id,omitempty"`
	AgentName                    string                 `json:"agentName,omitempty"`
	Big5                         *Big5Scores            `json:"big5,omitempty"`
	PersonalityPrompt            string                 `json:"personalityPrompt,omitempty"`
	TrueInterests                []string               `json:"trueInterests,omitempty"`
	TrueDislikes                 []string               `json:"trueDislikes,omitempty"`
	SpeechPatterns               []string               `json:"speechPatterns,omitempty"`
	CreatorDisplayName           string                 `json:"creatorDisplayName,omitempty"`
	StaticLoreMemories           []SeedMemory           `json:"staticLoreMemories,omitempty"`
	LoreGenerationContext        *LoreGenerationContext `json:"loreGenerationContext,omitempty"`
	IdentityMemoryTemplates      []IdentityMemory       `json:"identityMemoryTemplates,omitempty"`
	GenerateOriginStory          bool                   `json:"generateOriginStory,omitempty"`
	GeneratePersonalizedMemories bool                   `json:"generatePersonalizedMemories,omitempty"`
}

// LoreGenerationContext provides world context for LLM-based lore generation.
type LoreGenerationContext struct {
	WorldDescription         string            `json:"worldDescription"`
	EntityTerminology        map[string]string `json:"entityTerminology,omitempty"`
	OriginPromptInstructions string            `json:"originPromptInstructions,omitempty"`
}

// IdentityMemory represents a template for identity memory generation.
type IdentityMemory struct {
	Template   string   `json:"template"`
	FactType   string   `json:"factType,omitempty"`
	Importance float64  `json:"importance,omitempty"`
	Entities   []string `json:"entities,omitempty"`
}

// ModelConfig specifies LLM provider and model settings.
type ModelConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int32   `json:"max_tokens"`
}

// GenerateSeedMemoriesResponse is the response from seed memory generation.
type GenerateSeedMemoriesResponse struct {
	Memories       []SeedMemory `json:"memories"`
	MemoriesStored int32        `json:"memories_stored,omitempty"`
}

// SeedMemoriesOptions configures a bulk memory seeding request.
type SeedMemoriesOptions struct {
	UserID     string       `json:"user_id"`
	Memories   []SeedMemory `json:"memories"`
	InstanceID string       `json:"instance_id,omitempty"`
}

// SeedMemoriesResponse is the response from memory seeding.
type SeedMemoriesResponse struct {
	MemoriesCreated int32 `json:"memories_created"`
}

// GenerateBio generates a bio for an agent using AI.
func (g *GenerationResource) GenerateBio(ctx context.Context, agentID string, opts GenerateBioOptions) (*GenerateBioResponse, error) {
	var result GenerateBioResponse
	err := g.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/bio/generate", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateSeedMemories generates seed memories for an agent using AI.
func (g *GenerationResource) GenerateSeedMemories(ctx context.Context, agentID string, opts GenerateSeedMemoriesOptions) (*GenerateSeedMemoriesResponse, error) {
	var result GenerateSeedMemoriesResponse
	err := g.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/seed", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateCharacter generates a full character profile from a description.
func (g *GenerationResource) GenerateCharacter(ctx context.Context, opts GenerateCharacterOptions) (*GenerateCharacterResponse, error) {
	var result GenerateCharacterResponse
	err := g.http.Post(ctx, "/api/v1/agents/generate-character", opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateAndCreateOptions configures a combined generate + create request.
type GenerateAndCreateOptions struct {
	// AgentID is an optional UUID. If empty, a deterministic ID is derived from Name.
	AgentID     string   `json:"agent_id,omitempty"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender,omitempty"`
	Description string   `json:"description,omitempty"`
	Fields      []string `json:"fields,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	Language    string   `json:"language,omitempty"`
	// Provider selects the LLM backend ("gemini" | "openrouter" | "xai").
	// Leave empty to use the platform default (currently "gemini").
	Provider string `json:"provider,omitempty"`
	// Model optionally overrides the provider's default model.
	Model string `json:"model,omitempty"`
}

// GenerateAndCreateResponse is the response from the combined generate + create endpoint.
type GenerateAndCreateResponse struct {
	AgentID   string         `json:"agent_id"`
	Name      string         `json:"name"`
	Existing  bool           `json:"existing"`
	Generated map[string]any `json:"generated,omitempty"`
	Usage     struct {
		PromptTokens     int64  `json:"promptTokens"`
		CompletionTokens int64  `json:"completionTokens"`
		TotalTokens      int64  `json:"totalTokens"`
		Model            string `json:"model,omitempty"`
	} `json:"usage"`
}

// GenerateAndCreate generates a character and creates the agent in one idempotent call.
// If the agent already exists, the LLM is skipped and the existing agent is returned.
// Safe to call on every app startup.
func (g *GenerationResource) GenerateAndCreate(ctx context.Context, opts GenerateAndCreateOptions) (*GenerateAndCreateResponse, error) {
	var result GenerateAndCreateResponse
	err := g.http.Post(ctx, "/api/v1/agents/generate-and-create", opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SeedMemories bulk imports initial memories for an agent during setup.
func (g *GenerationResource) SeedMemories(ctx context.Context, agentID string, opts SeedMemoriesOptions) (*SeedMemoriesResponse, error) {
	var result SeedMemoriesResponse
	err := g.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/seed", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RegenerateAvatarOptions configures an avatar regeneration request.
type RegenerateAvatarOptions struct {
	Style string `json:"style,omitempty"`
}

// RegenerateAvatarResponse is the response from avatar regeneration.
type RegenerateAvatarResponse struct {
	Success          bool   `json:"success"`
	AvatarURL        string `json:"avatar_url"`
	Prompt           string `json:"prompt"`
	GenerationTimeMs int64  `json:"generation_time_ms"`
}

// RegenerateAvatar regenerates the avatar image for an agent using AI.
func (g *GenerationResource) RegenerateAvatar(ctx context.Context, agentID string, opts RegenerateAvatarOptions) (*RegenerateAvatarResponse, error) {
	var result RegenerateAvatarResponse
	err := g.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/avatar/generate", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
