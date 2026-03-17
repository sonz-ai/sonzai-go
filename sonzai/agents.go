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
	Generation    *GenerationResource
	CustomStates  *CustomStatesResource
}

func newAgentsResource(http *httpClient) *AgentsResource {
	return &AgentsResource{
		http:          http,
		Memory:        &MemoryResource{http: http},
		Personality:   &PersonalityResource{http: http},
		Sessions:      &SessionsResource{http: http},
		Instances:     &InstancesResource{http: http},
		Notifications: &NotificationsResource{http: http},
		Generation:    &GenerationResource{http: http},
		CustomStates:  &CustomStatesResource{http: http},
	}
}

// Create creates a new agent.
func (a *AgentsResource) Create(ctx context.Context, params CreateAgentParams) (*CreateAgentResult, error) {
	var result CreateAgentResult
	if err := a.http.post(ctx, "/api/v1/agents", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves an agent profile by ID.
func (a *AgentsResource) Get(ctx context.Context, agentID string) (*AgentProfile, error) {
	var result AgentProfile
	if err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s", agentID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an agent's profile.
func (a *AgentsResource) Update(ctx context.Context, agentID string, params UpdateAgentParams) (*UpdateAgentResult, error) {
	var result UpdateAgentResult
	if err := a.http.patch(ctx, fmt.Sprintf("/api/v1/agents/%s", agentID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an agent.
func (a *AgentsResource) Delete(ctx context.Context, agentID string) error {
	return a.http.del(ctx, fmt.Sprintf("/api/v1/agents/%s", agentID), nil)
}

// Dialogue generates agent dialogue.
func (a *AgentsResource) Dialogue(ctx context.Context, agentID string, params AgentDialogueParams) (*AgentDialogueResult, error) {
	var result AgentDialogueResult
	if err := a.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/dialogue", agentID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TriggerGameEvent fires a game event for async processing.
func (a *AgentsResource) TriggerGameEvent(ctx context.Context, agentID string, params TriggerGameEventParams) (*TriggerGameEventResult, error) {
	var result TriggerGameEventResult
	if err := a.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/events", agentID), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConstellation returns constellation graph data for an agent.
func (a *AgentsResource) GetConstellation(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/constellation", agentID), params, &result)
	return result, err
}

// GetBreakthroughs returns breakthrough data for an agent.
func (a *AgentsResource) GetBreakthroughs(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/breakthroughs", agentID), params, &result)
	return result, err
}

// GetWakeups returns wakeup schedule data for an agent.
func (a *AgentsResource) GetWakeups(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/wakeups", agentID), params, &result)
	return result, err
}

// GetMoodAggregate returns aggregated mood data for an agent.
func (a *AgentsResource) GetMoodAggregate(ctx context.Context, agentID string, userID, instanceID string) (map[string]interface{}, error) {
	params := map[string]string{}
	if userID != "" {
		params["user_id"] = userID
	}
	if instanceID != "" {
		params["instance_id"] = instanceID
	}
	var result map[string]interface{}
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood/aggregate", agentID), params, &result)
	return result, err
}

// Chat sends a chat message and returns the aggregated response.
func (a *AgentsResource) Chat(ctx context.Context, agentID string, opts ChatOptions) (*ChatResponse, error) {
	var parts []string
	var usage *ChatUsage

	err := a.http.streamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", agentID), opts, func(raw json.RawMessage) error {
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
	return a.http.streamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/chat", agentID), opts, func(raw json.RawMessage) error {
		var event ChatStreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// Evaluate evaluates an agent against a template.
func (a *AgentsResource) Evaluate(ctx context.Context, agentID string, messages []ChatMessage, templateID string, configOverride map[string]interface{}) (*EvaluationResult, error) {
	body := map[string]interface{}{
		"messages":    messages,
		"template_id": templateID,
	}
	if configOverride != nil {
		body["config_override"] = configOverride
	}

	var result EvaluationResult
	if err := a.http.post(ctx, fmt.Sprintf("/api/v1/agents/%s/evaluate", agentID), body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Simulate runs a simulation and calls the callback for each streaming event.
func (a *AgentsResource) Simulate(ctx context.Context, agentID string, body map[string]interface{}, callback func(SimulationEvent) error) error {
	return a.http.streamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/simulate", agentID), body, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// RunEval runs simulation + evaluation combined.
func (a *AgentsResource) RunEval(ctx context.Context, agentID string, body map[string]interface{}, callback func(SimulationEvent) error) error {
	return a.http.streamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/run-eval", agentID), body, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
}

// EvalOnly re-evaluates an existing run.
func (a *AgentsResource) EvalOnly(ctx context.Context, agentID string, templateID, sourceRunID string, callback func(SimulationEvent) error) error {
	body := map[string]interface{}{
		"template_id":   templateID,
		"source_run_id": sourceRunID,
	}
	return a.http.streamSSE(ctx, "POST", fmt.Sprintf("/api/v1/agents/%s/eval-only", agentID), body, func(raw json.RawMessage) error {
		var event SimulationEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil
		}
		return callback(event)
	})
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/mood-history", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/relationships", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/habits", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/goals", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/interests", agentID), params, &result)
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
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/diary", agentID), params, &result)
	return result, err
}

// GetUsers returns users for an agent.
func (a *AgentsResource) GetUsers(ctx context.Context, agentID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := a.http.get(ctx, fmt.Sprintf("/api/v1/agents/%s/users", agentID), nil, &result)
	return result, err
}
