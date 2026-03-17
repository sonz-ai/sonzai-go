package sonzai

import (
	"context"
	"fmt"
)

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

// CreateAgentOptions configures an agent creation request.
type CreateAgentOptions struct {
	AgentID                      string                    `json:"agent_id,omitempty"`
	UserID                       string                    `json:"user_id,omitempty"`
	UserDisplayName              string                    `json:"user_display_name,omitempty"`
	Name                         string                    `json:"name"`
	Gender                       string                    `json:"gender,omitempty"`
	Bio                          string                    `json:"bio,omitempty"`
	AvatarURL                    string                    `json:"avatar_url,omitempty"`
	ProjectID                    string                    `json:"project_id,omitempty"`
	PersonalityPrompt            string                    `json:"personality_prompt,omitempty"`
	SpeechPatterns               []string                  `json:"speech_patterns,omitempty"`
	TrueInterests                []string                  `json:"true_interests,omitempty"`
	TrueDislikes                 []string                  `json:"true_dislikes,omitempty"`
	PrimaryTraits                []string                  `json:"primary_traits,omitempty"`
	Big5                         *Big5Scores               `json:"big5,omitempty"`
	Dimensions                   *SDKPersonalityDimensions `json:"dimensions,omitempty"`
	Preferences                  map[string]string         `json:"preferences,omitempty"`
	Behaviors                    map[string]string         `json:"behaviors,omitempty"`
	ToolCapabilities             *AgentToolCapabilities    `json:"tool_capabilities,omitempty"`
	Language                     string                    `json:"language,omitempty"`
	SeedMemories                 []SeedMemory              `json:"seed_memories,omitempty"`
	LoreContext                  map[string]interface{}    `json:"lore_generation_context,omitempty"`
	GenerateOriginStory          bool                      `json:"generate_origin_story,omitempty"`
	GeneratePersonalizedMemories bool                      `json:"generate_personalized_memories,omitempty"`
}

// SeedMemory represents a memory to seed during agent creation.
type SeedMemory struct {
	Content    string   `json:"content"`
	FactType   string   `json:"fact_type,omitempty"`
	Importance float64  `json:"importance,omitempty"`
	Entities   []string `json:"entities,omitempty"`
}

// Agent represents an agent returned from the API.
type Agent struct {
	AgentID           string   `json:"agent_id"`
	Name              string   `json:"name"`
	Bio               string   `json:"bio,omitempty"`
	Gender            string   `json:"gender,omitempty"`
	AvatarURL         string   `json:"avatar_url,omitempty"`
	Status            string   `json:"status,omitempty"`
	PersonalityPrompt string   `json:"personality_prompt,omitempty"`
	SpeechPatterns    []string `json:"speech_patterns,omitempty"`
	TrueInterests     []string `json:"true_interests,omitempty"`
	TrueDislikes      []string `json:"true_dislikes,omitempty"`
	CreatedAt         string   `json:"created_at,omitempty"`
}

// UpdateAgentOptions configures an agent profile update request.
type UpdateAgentOptions struct {
	Name             string                    `json:"name,omitempty"`
	Bio              string                    `json:"bio,omitempty"`
	AvatarURL        string                    `json:"avatar_url,omitempty"`
	PersonalityPrompt string                   `json:"personality_prompt,omitempty"`
	SpeechPatterns   []string                  `json:"speech_patterns,omitempty"`
	TrueInterests    []string                  `json:"true_interests,omitempty"`
	TrueDislikes     []string                  `json:"true_dislikes,omitempty"`
	Big5             *Big5Scores               `json:"big5,omitempty"`
	Dimensions       *SDKPersonalityDimensions `json:"dimensions,omitempty"`
	ToolCapabilities *AgentToolCapabilities    `json:"tool_capabilities,omitempty"`
}

// Create creates a new agent.
func (a *AgentsResource) Create(ctx context.Context, opts CreateAgentOptions) (*Agent, error) {
	var result Agent
	err := a.http.Post(ctx, "/api/v1/agents", opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns an agent by ID.
func (a *AgentsResource) Get(ctx context.Context, agentID string) (*Agent, error) {
	var result Agent
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s", agentID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an agent's profile.
func (a *AgentsResource) Update(ctx context.Context, agentID string, opts UpdateAgentOptions) (*Agent, error) {
	var result Agent
	err := a.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/profile", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an agent.
func (a *AgentsResource) Delete(ctx context.Context, agentID string) error {
	return a.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s", agentID), nil)
}

// UpdatePersonalityOptions configures a personality update request.
type UpdatePersonalityOptions struct {
	Big5       *Big5Scores               `json:"big5"`
	Dimensions *SDKPersonalityDimensions `json:"dimensions,omitempty"`
}

// UpdatePersonalityResponse is the response from updating personality.
type UpdatePersonalityResponse struct {
	Success bool `json:"success"`
}

// UpdatePersonality updates an agent's Big5 personality scores.
func (a *AgentsResource) UpdatePersonality(ctx context.Context, agentID string, opts UpdatePersonalityOptions) (*UpdatePersonalityResponse, error) {
	var result UpdatePersonalityResponse
	err := a.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/personality", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
