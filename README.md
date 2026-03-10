# Sonzai Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/sonz-ai/sonzai-go/sonzai.svg)](https://pkg.go.dev/github.com/sonz-ai/sonzai-go/sonzai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Go SDK for the [Sonzai Character Engine API](https://sonz.ai). Build AI characters with persistent memory, evolving personality, and proactive behaviors.

**Zero dependencies.** Uses only the Go standard library.

## Installation

```bash
go get github.com/sonz-ai/sonzai-go/sonzai
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/sonz-ai/sonzai-go/sonzai"
)

func main() {
	client := sonzai.NewClient("your-api-key")

	resp, err := client.Agents.Chat(context.Background(), "your-agent-id", sonzai.ChatOptions{
		Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
		UserID:   "user-123",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Content)
}
```

## Authentication

Get your API key from the [Sonzai Dashboard](https://platform.sonz.ai) under **Projects > API Keys**.

```go
// Pass directly
client := sonzai.NewClient("sk-...")

// Or set the environment variable
// export SONZAI_API_KEY=sk-...
client := sonzai.NewClient("")
```

## Usage

### Chat (Streaming)

```go
err := client.Agents.ChatStream(ctx, "agent-id", sonzai.ChatOptions{
	Messages: []sonzai.ChatMessage{{Role: "user", Content: "Tell me a story"}},
}, func(event sonzai.ChatStreamEvent) error {
	fmt.Print(event.Content())
	return nil
})
```

### Chat (Non-streaming)

```go
resp, err := client.Agents.Chat(ctx, "agent-id", sonzai.ChatOptions{
	Messages:  []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
	UserID:    "user-123",
	SessionID: "session-456",
})
fmt.Println(resp.Content)
fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
```

### Memory

```go
// Get memory tree
memory, err := client.Agents.Memory.List(ctx, "agent-id", &sonzai.MemoryListOptions{
	UserID: "user-123",
})
for _, node := range memory.Nodes {
	fmt.Printf("%s (importance: %.2f)\n", node.Title, node.Importance)
}

// Search memories
results, err := client.Agents.Memory.Search(ctx, "agent-id", sonzai.MemorySearchOptions{
	Query: "favorite food",
})
for _, fact := range results.Results {
	fmt.Printf("%s (score: %.2f)\n", fact.Content, fact.Score)
}

// Timeline
timeline, err := client.Agents.Memory.Timeline(ctx, "agent-id", &sonzai.MemoryTimelineOptions{
	UserID: "user-123",
	Start:  "2026-01-01",
	End:    "2026-03-01",
})
```

### Personality

```go
personality, err := client.Agents.Personality.Get(ctx, "agent-id", nil)
fmt.Printf("Name: %s\n", personality.Profile.Name)
fmt.Printf("Openness: %.2f\n", personality.Profile.Big5.Openness.Score)
fmt.Printf("Warmth: %d/10\n", personality.Profile.Dimensions.Warmth)
```

### Sessions

```go
// Start
_, err := client.Agents.Sessions.Start(ctx, "agent-id", sonzai.SessionStartOptions{
	UserID:    "user-123",
	SessionID: "session-456",
})

// End
_, err := client.Agents.Sessions.End(ctx, "agent-id", sonzai.SessionEndOptions{
	UserID:          "user-123",
	SessionID:       "session-456",
	TotalMessages:   10,
	DurationSeconds: 300,
})
```

### Agent Instances

```go
// List
instances, err := client.Agents.Instances.List(ctx, "agent-id")

// Create
instance, err := client.Agents.Instances.Create(ctx, "agent-id", "Test Instance", "")

// Reset
_, err := client.Agents.Instances.Reset(ctx, "agent-id", instance.InstanceID)

// Delete
err := client.Agents.Instances.Delete(ctx, "agent-id", instance.InstanceID)
```

### Notifications

```go
// List pending
notifications, err := client.Agents.Notifications.List(ctx, "agent-id", &sonzai.NotificationListOptions{
	Status: "pending",
})
for _, n := range notifications.Notifications {
	fmt.Printf("[%s] %s\n", n.CheckType, n.GeneratedMessage)
}

// Consume
_, err := client.Agents.Notifications.Consume(ctx, "agent-id", "msg-id")

// History
history, err := client.Agents.Notifications.History(ctx, "agent-id", 50)
```

### Context Engine Data

```go
mood, err := client.Agents.GetMood(ctx, "agent-id", "user-123", "")
relationships, err := client.Agents.GetRelationships(ctx, "agent-id", "user-123", "")
habits, err := client.Agents.GetHabits(ctx, "agent-id", "", "")
goals, err := client.Agents.GetGoals(ctx, "agent-id", "", "")
interests, err := client.Agents.GetInterests(ctx, "agent-id", "", "")
diary, err := client.Agents.GetDiary(ctx, "agent-id", "", "")
users, err := client.Agents.GetUsers(ctx, "agent-id")
```

### Evaluation

```go
result, err := client.Agents.Evaluate(ctx, "agent-id",
	[]sonzai.ChatMessage{
		{Role: "user", Content: "I'm feeling sad today"},
		{Role: "assistant", Content: "I'm sorry to hear that..."},
	},
	"template-uuid",
	nil,
)
fmt.Printf("Score: %.2f\nFeedback: %s\n", result.Score, result.Feedback)
```

### Simulation

```go
err := client.Agents.Simulate(ctx, "agent-id", map[string]interface{}{
	"user_persona": map[string]interface{}{
		"name":                "Alex",
		"background":          "College student",
		"personality_traits":  []string{"curious", "friendly"},
		"communication_style": "casual",
	},
	"config": map[string]interface{}{
		"max_sessions":          3,
		"max_turns_per_session": 10,
	},
}, func(event sonzai.SimulationEvent) error {
	fmt.Printf("[%s] %s\n", event.Type, event.Message)
	return nil
})
```

### Eval Templates

```go
// List
templates, err := client.EvalTemplates.List(ctx, "")

// Create
template, err := client.EvalTemplates.Create(ctx, sonzai.EvalTemplateCreateOptions{
	Name:          "Empathy Check",
	ScoringRubric: "Evaluate emotional awareness",
	Categories: []sonzai.EvalTemplateCategory{
		{Name: "Awareness", Weight: 0.5, Criteria: "..."},
		{Name: "Response", Weight: 0.5, Criteria: "..."},
	},
})

// Delete
err := client.EvalTemplates.Delete(ctx, template.ID)
```

### Eval Runs

```go
runs, err := client.EvalRuns.List(ctx, "agent-id", 20, 0)
run, err := client.EvalRuns.Get(ctx, "run-id")
err := client.EvalRuns.Delete(ctx, "run-id")
```

## Configuration

```go
client := sonzai.NewClient("sk-...",
	sonzai.WithBaseURL("https://api.sonz.ai"), // or SONZAI_BASE_URL env var
	sonzai.WithTimeout(30 * time.Second),
)
```

## Error Handling

```go
resp, err := client.Agents.Chat(ctx, "agent-id", opts)
if err != nil {
	switch err.(type) {
	case *sonzai.AuthenticationError:
		log.Fatal("Invalid API key")
	case *sonzai.NotFoundError:
		log.Fatal("Agent not found")
	case *sonzai.RateLimitError:
		log.Fatal("Rate limit exceeded")
	case *sonzai.BadRequestError:
		log.Fatal("Bad request")
	default:
		log.Fatalf("API error: %v", err)
	}
}
```

## Development

```bash
# Run tests
go test ./sonzai/ -v

# Run tests with coverage
go test ./sonzai/ -cover
```

## License

MIT License - see [LICENSE](LICENSE) for details.
