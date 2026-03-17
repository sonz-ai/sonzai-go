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
