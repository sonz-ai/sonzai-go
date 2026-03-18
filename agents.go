package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	}
}

// Chat sends a chat message and returns the aggregated response.
func (a *AgentsResource) Chat(ctx context.Context, agentID string, opts ChatOptions) (*ChatResponse, error) {
	var parts []string
	var usage *ChatUsage

	err := a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", agentID), opts, func(raw json.RawMessage) error {
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil // skip malformed events
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
func (a *AgentsResource) ChatStream(ctx context.Context, agentID string, opts ChatOptions, callback func(ChatStreamEvent) error) error {
	return a.http.StreamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", agentID), opts, func(raw json.RawMessage) error {
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// ChatStreamChannel sends a chat message and returns a channel of streaming events.
// The channel is closed when the stream ends or the context is cancelled.
func (a *AgentsResource) ChatStreamChannel(ctx context.Context, agentID string, opts ChatOptions) (<-chan ChatStreamEvent, <-chan error) {
	ch := make(chan ChatStreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		err := a.ChatStream(ctx, agentID, opts, func(event ChatStreamEvent) error {
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
func (a *AgentsResource) GetMood(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood", agentID), params, &result)
	return result, err
}

// GetMoodHistory returns mood history for an agent.
func (a *AgentsResource) GetMoodHistory(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood-history", agentID), params, &result)
	return result, err
}

// GetMoodAggregate returns aggregated mood statistics for an agent.
func (a *AgentsResource) GetMoodAggregate(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood/aggregate", agentID), params, &result)
	return result, err
}

// GetRelationships returns relationship data for an agent.
func (a *AgentsResource) GetRelationships(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/relationships", agentID), params, &result)
	return result, err
}

// GetHabits returns habit data for an agent.
func (a *AgentsResource) GetHabits(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/habits", agentID), params, &result)
	return result, err
}

// GetGoals returns goal data for an agent.
func (a *AgentsResource) GetGoals(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/goals", agentID), params, &result)
	return result, err
}

// GetInterests returns interest data for an agent.
func (a *AgentsResource) GetInterests(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/interests", agentID), params, &result)
	return result, err
}

// GetDiary returns diary entries for an agent.
func (a *AgentsResource) GetDiary(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/diary", agentID), params, &result)
	return result, err
}

// GetUsers returns users for an agent.
func (a *AgentsResource) GetUsers(ctx context.Context, agentID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users", agentID), nil, &result)
	return result, err
}

// TriggerEvent triggers a game event / activity for an agent.
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
	err := a.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/capabilities", agentID), opts, &result)
	return &result, err
}

// Consolidate triggers memory consolidation for an agent.
func (a *AgentsResource) Consolidate(ctx context.Context, agentID string, opts ConsolidateOptions) error {
	return a.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/consolidate", agentID), opts, nil)
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
	err := a.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/summaries", agentID), params, &result)
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
