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
}

func newAgentsResource(http *httpClient) *AgentsResource {
	return &AgentsResource{
		http:          http,
		Memory:        &MemoryResource{http: http},
		Personality:   &PersonalityResource{http: http},
		Sessions:      &SessionsResource{http: http},
		Instances:     &InstancesResource{http: http},
		Notifications: &NotificationsResource{http: http},
	}
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
