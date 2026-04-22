package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync/atomic"
)

// AgentsResource provides agent-scoped operations.
type AgentsResource struct {
	http          *httpClient
	Memory        *MemoryResource
	Personality   *PersonalityResource
	Sessions      *SessionsResource
	Instances     *InstancesResource
	Notifications *NotificationsResource
	CustomState   *CustomStateResource
	Image         *ImageResource
	Voice         *VoiceResource
	Wakeups       *WakeupResource
	Generation    *GenerationResource
	Priming       *PrimingResource
	Inventory     *InventoryResource
	Schedules     *SchedulesResource
}

func newAgentsResource(http *httpClient) *AgentsResource {
	return &AgentsResource{
		http:          http,
		Memory:        &MemoryResource{http: http},
		Personality:   &PersonalityResource{http: http},
		Sessions:      &SessionsResource{http: http},
		Instances:     &InstancesResource{http: http},
		Notifications: &NotificationsResource{http: http},
		CustomState:   &CustomStateResource{http: http},
		Image:         &ImageResource{http: http},
		Voice:         &VoiceResource{http: http},
		Wakeups:       &WakeupResource{http: http},
		Generation:    &GenerationResource{http: http},
		Priming:       &PrimingResource{http: http},
		Inventory:     &InventoryResource{http: http},
		Schedules:     &SchedulesResource{http: http},
	}
}

// Chat sends a chat message and returns the aggregated response.
func (a *AgentsResource) Chat(ctx context.Context, params AgentChatParams) (*ChatResponse, error) {
	var parts []string
	var usage *ChatUsage
	var totalEvents, malformedEvents int64

	err := a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", params.AgentID), params.ChatOptions, func(raw json.RawMessage) error {
		totalEvents++
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			malformedEvents++
			slog.Warn("skipping malformed SSE event in Chat",
				"error", err,
				"agent_id", params.AgentID,
				"malformed_count", malformedEvents,
				"total_count", totalEvents,
			)
			if totalEvents >= 4 && malformedEvents*2 > totalEvents {
				return fmt.Errorf("too many malformed SSE events: %d/%d", malformedEvents, totalEvents)
			}
			return nil
		}
		if c := event.Content(); c != "" {
			parts = append(parts, c)
		}
		if event.Usage != nil {
			usage = event.Usage
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Content: strings.Join(parts, ""),
		Usage:   usage,
	}, nil
}

// ChatStream sends a chat message and calls the callback for each streaming event.
func (a *AgentsResource) ChatStream(ctx context.Context, params AgentChatParams, callback func(ChatStreamEvent) error) error {
	var totalEvents, malformedEvents atomic.Int64
	return a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", params.AgentID), params.ChatOptions, func(raw json.RawMessage) error {
		total := totalEvents.Add(1)
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			malformed := malformedEvents.Add(1)
			slog.Warn("skipping malformed SSE event in ChatStream",
				"error", err,
				"agent_id", params.AgentID,
				"malformed_count", malformed,
				"total_count", total,
			)
			if total >= 4 && malformed*2 > total {
				return fmt.Errorf("too many malformed SSE events: %d/%d", malformed, total)
			}
			return nil
		}
		return callback(event)
	})
}

// ChatStreamChannel sends a chat message and returns a channel of streaming events.
// The channel is closed when the stream ends or the context is cancelled.
func (a *AgentsResource) ChatStreamChannel(ctx context.Context, params AgentChatParams) (<-chan ChatStreamEvent, <-chan error) {
	ch := make(chan ChatStreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		err := a.ChatStream(ctx, params, func(event ChatStreamEvent) error {
			select {
			case ch <- event:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil {
			errCh <- err
		}
	}()

	return ch, errCh
}

// GetMood returns the current mood for an agent.
func (a *AgentsResource) GetMood(ctx context.Context, agentID string, userID, instanceID string) (*MoodResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result MoodResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood", agentID), params, &result)
	return &result, err
}

// GetMoodHistory returns mood history for an agent.
func (a *AgentsResource) GetMoodHistory(ctx context.Context, agentID string, userID, instanceID string) (*MoodHistoryResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result MoodHistoryResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood-history", agentID), params, &result)
	return &result, err
}

// GetMoodAggregate returns aggregated mood statistics for an agent.
func (a *AgentsResource) GetMoodAggregate(ctx context.Context, agentID string, userID, instanceID string) (*MoodAggregateResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result MoodAggregateResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood/aggregate", agentID), params, &result)
	return &result, err
}

// GetRelationships returns relationship data for an agent.
func (a *AgentsResource) GetRelationships(ctx context.Context, agentID string, userID, instanceID string) (*RelationshipsResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result RelationshipsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/relationships", agentID), params, &result)
	return &result, err
}

// ListHabits returns habit data for an agent.
func (a *AgentsResource) ListHabits(ctx context.Context, agentID string, userID, instanceID string) (*HabitsResponse, error) {
	return a.getHabitsImpl(ctx, agentID, userID, instanceID)
}

// Deprecated: Use ListHabits instead.
func (a *AgentsResource) GetHabits(ctx context.Context, agentID string, userID, instanceID string) (*HabitsResponse, error) {
	return a.getHabitsImpl(ctx, agentID, userID, instanceID)
}

func (a *AgentsResource) getHabitsImpl(ctx context.Context, agentID string, userID, instanceID string) (*HabitsResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result HabitsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/habits", agentID), params, &result)
	return &result, err
}

// Habit represents a full agent habit returned from the API.
type Habit struct {
	ID               string  `json:"id,omitempty"`
	AgentID          string  `json:"agent_id"`
	UserID           string  `json:"user_id,omitempty"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	Description      string  `json:"description"`
	DisplayName      string  `json:"display_name,omitempty"`
	Strength         float64 `json:"strength"`
	Formed           bool    `json:"formed"`
	ObservationCount int     `json:"observation_count"`
	LastReinforcedAt string  `json:"last_reinforced_at,omitempty"`
	FormedAt         string  `json:"formed_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// CreateHabitOptions configures a habit creation request.
type CreateHabitOptions struct {
	UserID      string  `json:"user_id,omitempty"`
	Name        string  `json:"name"`
	Category    string  `json:"category,omitempty"`
	Description string  `json:"description,omitempty"`
	DisplayName string  `json:"display_name,omitempty"`
	Strength    float64 `json:"strength,omitempty"`
}

// UpdateHabitOptions configures a habit update request.
type UpdateHabitOptions struct {
	UserID      string   `json:"user_id,omitempty"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	Strength    *float64 `json:"strength,omitempty"`
}

// CreateHabit creates a new habit for an agent. Set UserID in opts for a per-user habit.
func (a *AgentsResource) CreateHabit(ctx context.Context, agentID string, opts CreateHabitOptions) (*Habit, error) {
	var result Habit
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/habits", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateHabit updates an existing habit by name.
func (a *AgentsResource) UpdateHabit(ctx context.Context, agentID, habitName string, opts UpdateHabitOptions) (*Habit, error) {
	var result Habit
	err := a.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/habits/%s", agentID, url.PathEscape(habitName)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteHabit removes a habit. Pass userID for per-user habits.
func (a *AgentsResource) DeleteHabit(ctx context.Context, agentID, habitName, userID string) error {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	return a.http.DeleteWithParams(ctx, fmt.Sprintf("/api/v1/agents/%s/habits/%s", agentID, url.PathEscape(habitName)), params, nil)
}

// Goal represents an agent goal returned from the API.
type Goal struct {
	GoalID        string   `json:"goal_id"`
	AgentID       string   `json:"agent_id"`
	UserID        string   `json:"user_id,omitempty"`
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Priority      int      `json:"priority"`
	Status        string   `json:"status"`
	RelatedTraits []string `json:"related_traits,omitempty"`
	CreatedAt     string   `json:"created_at"`
	AchievedAt    string   `json:"achieved_at,omitempty"`
	UpdatedAt     string   `json:"updated_at"`
}

// GoalsResponse wraps a list of goals returned by the API.
type GoalsResponse struct {
	Goals []Goal `json:"goals"`
}

// CreateGoalOptions configures a goal creation request.
type CreateGoalOptions struct {
	UserID        string   `json:"user_id,omitempty"`
	Type          string   `json:"type,omitempty"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Priority      int      `json:"priority,omitempty"`
	RelatedTraits []string `json:"related_traits,omitempty"`
}

// UpdateGoalOptions configures a goal update request.
type UpdateGoalOptions struct {
	UserID        string   `json:"user_id,omitempty"`
	Title         string   `json:"title,omitempty"`
	Description   string   `json:"description,omitempty"`
	Priority      *int     `json:"priority,omitempty"`
	Status        string   `json:"status,omitempty"`
	RelatedTraits []string `json:"related_traits,omitempty"`
}

// ListGoals returns goal data for an agent. Pass userID to get combined agent-global + per-user goals.
func (a *AgentsResource) ListGoals(ctx context.Context, agentID string, userID, instanceID string) (*GoalsResponse, error) {
	return a.getGoalsImpl(ctx, agentID, userID, instanceID)
}

// Deprecated: Use ListGoals instead.
func (a *AgentsResource) GetGoals(ctx context.Context, agentID string, userID, instanceID string) (*GoalsResponse, error) {
	return a.getGoalsImpl(ctx, agentID, userID, instanceID)
}

func (a *AgentsResource) getGoalsImpl(ctx context.Context, agentID string, userID, instanceID string) (*GoalsResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result GoalsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/goals", agentID), params, &result)
	return &result, err
}

// CreateGoal creates a new goal for an agent. Set UserID in opts for a per-user goal.
func (a *AgentsResource) CreateGoal(ctx context.Context, agentID string, opts CreateGoalOptions) (*Goal, error) {
	var result Goal
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/goals", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGoal updates an existing goal. Set UserID in opts for per-user goals.
func (a *AgentsResource) UpdateGoal(ctx context.Context, agentID, goalID string, opts UpdateGoalOptions) (*Goal, error) {
	var result Goal
	err := a.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/goals/%s", agentID, goalID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGoal soft-deletes (abandons) a goal. Pass userID for per-user goals.
func (a *AgentsResource) DeleteGoal(ctx context.Context, agentID, goalID, userID string) error {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	return a.http.DeleteWithParams(ctx, fmt.Sprintf("/api/v1/agents/%s/goals/%s", agentID, goalID), params, nil)
}

// GetInterests returns interest data for an agent.
func (a *AgentsResource) GetInterests(ctx context.Context, agentID string, userID, instanceID string) (*InterestsResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result InterestsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/interests", agentID), params, &result)
	return &result, err
}

// GetDiary returns diary entries for an agent.
func (a *AgentsResource) GetDiary(ctx context.Context, agentID string, userID, instanceID string) (*DiaryResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result DiaryResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/diary", agentID), params, &result)
	return &result, err
}

// GetUsers returns users for an agent.
func (a *AgentsResource) GetUsers(ctx context.Context, agentID string, opts *GetUsersOptions) (*UsersResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", opts.Offset)
		}
		if opts.SortBy != "" {
			params["sort_by"] = opts.SortBy
		}
		if opts.SortOrder != "" {
			params["sort_order"] = opts.SortOrder
		}
	}
	var result UsersResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users", agentID), params, &result)
	return &result, err
}

// TriggerEvent triggers a backend event / activity for an agent.
func (a *AgentsResource) TriggerEvent(ctx context.Context, agentID string, opts TriggerEventOptions) (*TriggerEventResponse, error) {
	var result TriggerEventResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/events", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Dialogue initiates a dialogue with an agent.
func (a *AgentsResource) Dialogue(ctx context.Context, agentID string, opts DialogueOptions) (*DialogueResponse, error) {
	var result DialogueResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/dialogue", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns a paginated list of agents.
func (a *AgentsResource) List(ctx context.Context, opts AgentListOptions) (*AgentListResponse, error) {
	params := map[string]string{}
	if opts.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", opts.PageSize)
	}
	if opts.Cursor != "" {
		params["cursor"] = opts.Cursor
	}
	if opts.Search != "" {
		params["search"] = opts.Search
	}
	if opts.ProjectID != "" {
		params["project_id"] = opts.ProjectID
	}
	var result AgentListResponse
	err := a.http.Get(ctx, "/api/v1/agents", params, &result)
	return &result, err
}

// SetStatus sets the active status for an agent.
func (a *AgentsResource) SetStatus(ctx context.Context, agentID string, isActive bool) (*SetStatusResponse, error) {
	var result SetStatusResponse
	err := a.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/status", agentID), map[string]interface{}{"is_active": isActive}, &result)
	return &result, err
}

// UpdateProject updates the project assignment for an agent.
func (a *AgentsResource) UpdateProject(ctx context.Context, agentID string, projectID string) (*UpdateProjectResponse, error) {
	var result UpdateProjectResponse
	err := a.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/project", agentID), map[string]interface{}{"project_id": projectID}, &result)
	return &result, err
}

// GetCapabilities returns the capabilities for an agent.
func (a *AgentsResource) GetCapabilities(ctx context.Context, agentID string) (*AgentCapabilities, error) {
	var result AgentCapabilities
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/capabilities", agentID), nil, &result)
	return &result, err
}

// UpdateCapabilities updates the capabilities for an agent.
func (a *AgentsResource) UpdateCapabilities(ctx context.Context, agentID string, opts UpdateCapabilitiesOptions) (*AgentCapabilities, error) {
	var result AgentCapabilities
	err := a.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/capabilities", agentID), opts, &result)
	return &result, err
}

// Consolidate triggers memory consolidation for an agent.
func (a *AgentsResource) Consolidate(ctx context.Context, agentID string, opts ConsolidateOptions) error {
	return a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/consolidate", agentID), opts, nil)
}

// GetSummaries returns memory summaries for an agent.
func (a *AgentsResource) GetSummaries(ctx context.Context, agentID string, opts SummariesOptions) (*SummariesResponse, error) {
	params := map[string]string{}
	if opts.Period != "" {
		params["period"] = opts.Period
	}
	if opts.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", opts.Limit)
	}
	var result SummariesResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/memory/summaries", agentID), params, &result)
	return &result, err
}

// GetConstellation returns the knowledge graph (nodes, edges, insights) for an agent.
func (a *AgentsResource) GetConstellation(ctx context.Context, agentID string, userID, instanceID string) (*ConstellationResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result ConstellationResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/constellation", agentID), params, &result)
	return &result, err
}

// CreateConstellationNodeOptions configures a constellation node creation request.
type CreateConstellationNodeOptions struct {
	UserID       string  `json:"user_id,omitempty"`
	NodeType     string  `json:"node_type,omitempty"`
	Label        string  `json:"label"`
	Description  string  `json:"description,omitempty"`
	Significance float64 `json:"significance,omitempty"`
}

// UpdateConstellationNodeOptions configures a constellation node update request.
type UpdateConstellationNodeOptions struct {
	Label        string   `json:"label,omitempty"`
	Description  string   `json:"description,omitempty"`
	Significance *float64 `json:"significance,omitempty"`
	NodeType     string   `json:"node_type,omitempty"`
}

// CreateConstellationNode creates a new constellation node (lore) for an agent.
func (a *AgentsResource) CreateConstellationNode(ctx context.Context, agentID string, opts CreateConstellationNodeOptions) (*ConstellationNode, error) {
	var result ConstellationNode
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/constellation/nodes", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateConstellationNode updates an existing constellation node.
func (a *AgentsResource) UpdateConstellationNode(ctx context.Context, agentID, nodeID string, opts UpdateConstellationNodeOptions) (*ConstellationNode, error) {
	var result ConstellationNode
	err := a.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/constellation/nodes/%s", agentID, nodeID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteConstellationNode removes a constellation node.
func (a *AgentsResource) DeleteConstellationNode(ctx context.Context, agentID, nodeID string) error {
	return a.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/constellation/nodes/%s", agentID, nodeID), nil)
}

// ListBreakthroughs returns breakthroughs for an agent.
func (a *AgentsResource) ListBreakthroughs(ctx context.Context, agentID string, userID, instanceID string) (*BreakthroughsResponse, error) {
	return a.getBreakthroughsImpl(ctx, agentID, userID, instanceID)
}

// Deprecated: Use ListBreakthroughs instead.
func (a *AgentsResource) GetBreakthroughs(ctx context.Context, agentID string, userID, instanceID string) (*BreakthroughsResponse, error) {
	return a.getBreakthroughsImpl(ctx, agentID, userID, instanceID)
}

func (a *AgentsResource) getBreakthroughsImpl(ctx context.Context, agentID string, userID, instanceID string) (*BreakthroughsResponse, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result BreakthroughsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/breakthroughs", agentID), params, &result)
	return &result, err
}

// GetTimeMachine returns a point-in-time snapshot of agent personality and mood.
func (a *AgentsResource) GetTimeMachine(ctx context.Context, agentID string, opts TimeMachineOptions) (*TimeMachineResponse, error) {
	params := map[string]string{"at": opts.At}
	if opts.UserID != "" {
		params["user_id"] = opts.UserID
	}
	if opts.InstanceID != "" {
		params["instance_id"] = opts.InstanceID
	}
	var result TimeMachineResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/timemachine", agentID), params, &result)
	return &result, err
}

// ListCustomTools returns the custom tools for an agent.
func (a *AgentsResource) ListCustomTools(ctx context.Context, agentID string) (*CustomToolListResponse, error) {
	var result CustomToolListResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/tools", agentID), nil, &result)
	return &result, err
}

// CreateCustomTool creates a custom tool for an agent.
func (a *AgentsResource) CreateCustomTool(ctx context.Context, agentID string, opts CreateCustomToolOptions) (*CustomToolDefinition, error) {
	var result CustomToolDefinition
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/tools", agentID), opts, &result)
	return &result, err
}

// UpdateCustomTool updates a custom tool for an agent.
func (a *AgentsResource) UpdateCustomTool(ctx context.Context, agentID string, toolName string, opts UpdateCustomToolOptions) error {
	return a.http.Put(ctx, fmt.Sprintf("/api/v1/agents/%s/tools/%s", agentID, toolName), opts, nil)
}

// DeleteCustomTool deletes a custom tool for an agent.
func (a *AgentsResource) DeleteCustomTool(ctx context.Context, agentID string, toolName string) error {
	return a.http.Delete(ctx, fmt.Sprintf("/api/v1/agents/%s/tools/%s", agentID, toolName), nil)
}

// Process runs the full Context Engine pipeline on conversation messages
// without generating a chat response.
func (a *AgentsResource) Process(ctx context.Context, agentID string, opts ProcessOptions) (*ProcessResponse, error) {
	var result ProcessResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/process", agentID), opts, &result)
	return &result, err
}

// GetModels returns available LLM providers and models.
func (a *AgentsResource) GetModels(ctx context.Context, agentID string) (*ModelsResponse, error) {
	var result ModelsResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/models", agentID), nil, &result)
	return &result, err
}

// GetContext returns the full enriched agent context in a single call.
func (a *AgentsResource) GetContext(ctx context.Context, agentID string, opts GetContextOptions) (*EnrichedContextResponse, error) {
	params := map[string]string{"userId": opts.UserID}
	if opts.SessionID != "" {
		params["sessionId"] = opts.SessionID
	}
	if opts.InstanceID != "" {
		params["instanceId"] = opts.InstanceID
	}
	if opts.Query != "" {
		params["query"] = opts.Query
	}
	if opts.Language != "" {
		params["language"] = opts.Language
	}
	if opts.Timezone != "" {
		params["timezone"] = opts.Timezone
	}
	var result EnrichedContextResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/context", agentID), params, &result)
	return &result, err
}

// KnowledgeSearch searches the knowledge base for an agent using the tool endpoint.
func (a *AgentsResource) KnowledgeSearch(ctx context.Context, agentID string, opts AgentKBSearchOptions) (*AgentKBSearchResponse, error) {
	var result AgentKBSearchResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/tools/kb-search", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTools returns the tool schemas available for an agent (for BYO-LLM integrations).
func (a *AgentsResource) GetTools(ctx context.Context, agentID string) (*ToolSchemasResponse, error) {
	var result ToolSchemasResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/tools", agentID), nil, &result)
	return &result, err
}

// Fork creates a copy of an agent with a new ID.
func (a *AgentsResource) Fork(ctx context.Context, agentID string, opts *ForkAgentOptions) (*ForkResponse, error) {
	var body interface{}
	if opts != nil {
		body = opts
	}
	var result ForkResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/fork", agentID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetForkStatus checks the status of a fork operation.
func (a *AgentsResource) GetForkStatus(ctx context.Context, agentID string) (*ForkStatusResponse, error) {
	var result ForkStatusResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/fork/status", agentID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PlaygroundChat sends a chat message via the playground endpoint (SSE streaming).
// This is the same as Chat but uses the playground path for dashboard testing.
func (a *AgentsResource) PlaygroundChat(ctx context.Context, params AgentChatParams) (*ChatResponse, error) {
	var parts []string
	var usage *ChatUsage

	err := a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/playground/chat", params.AgentID), params.ChatOptions, func(raw json.RawMessage) error {
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			slog.Warn("skipping malformed SSE event", "error", err)
			return nil
		}
		if c := event.Content(); c != "" {
			parts = append(parts, c)
		}
		if event.Usage != nil {
			usage = event.Usage
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Content: strings.Join(parts, ""),
		Usage:   usage,
	}, nil
}

// PlaygroundChatStream sends a chat message via the playground endpoint and calls the callback for each streaming event.
func (a *AgentsResource) PlaygroundChatStream(ctx context.Context, params AgentChatParams, callback func(ChatStreamEvent) error) error {
	return a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/playground/chat", params.AgentID), params.ChatOptions, func(raw json.RawMessage) error {
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			slog.Warn("skipping malformed SSE event", "error", err)
			return nil
		}
		return callback(event)
	})
}

// KnowledgeSearchGet searches the knowledge base for an agent using a GET request with query parameters.
func (a *AgentsResource) KnowledgeSearchGet(ctx context.Context, agentID string, query string, limit int) (*AgentKBSearchResponse, error) {
	params := map[string]string{
		"q": query,
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	var result AgentKBSearchResponse
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/tools/kb-search", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateAvatar triggers avatar generation for an agent.
func (a *AgentsResource) GenerateAvatar(ctx context.Context, agentID string, opts *GenerateAvatarOptions) (*GenerateAvatarResponse, error) {
	var body interface{}
	if opts != nil {
		body = opts
	}
	var result GenerateAvatarResponse
	err := a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/avatar/generate", agentID), body, &result)
	return &result, err
}
