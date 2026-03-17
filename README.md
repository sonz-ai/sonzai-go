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

### Agent Lifecycle

```go
// Create an agent
result, err := client.Agents.Create(ctx, sonzai.CreateAgentParams{
	UserID:    "user-123",
	AgentName: "Luna",
	Gender:    "female",
	Big5:      sonzai.Big5Scores{Openness: 0.8, Conscientiousness: 0.6, Extraversion: 0.7, Agreeableness: 0.85, Neuroticism: 0.3},
})

// Get agent profile
agent, err := client.Agents.Get(ctx, "agent-id")

// Update agent
_, err := client.Agents.Update(ctx, "agent-id", sonzai.UpdateAgentParams{
	Name: "Luna v2",
	Bio:  "An updated bio",
})

// Delete agent
err := client.Agents.Delete(ctx, "agent-id")
```

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

// Seed memories
seedResult, err := client.Agents.Memory.Seed(ctx, "agent-id", sonzai.SeedMemoriesParams{
	UserID: "user-123",
	Memories: []sonzai.MemoryCandidate{
		{Content: "Likes coffee", FactType: "preference", Importance: 0.7},
		{Content: "Lives in Singapore", FactType: "location", Importance: 0.9},
	},
})

// List facts
facts, err := client.Agents.Memory.ListFacts(ctx, "agent-id", &sonzai.ListFactsOptions{
	UserID: "user-123", FactType: "preference", Limit: 50,
})

// Reset memories
resetResult, err := client.Agents.Memory.Reset(ctx, "agent-id", sonzai.ResetMemoryParams{
	UserID: "user-123",
})
```

### Personality

```go
// Get personality
personality, err := client.Agents.Personality.Get(ctx, "agent-id", nil)
fmt.Printf("Name: %s\n", personality.Profile.Name)
fmt.Printf("Openness: %.2f\n", personality.Profile.Big5.Openness.Score)
fmt.Printf("Warmth: %d/10\n", personality.Profile.Dimensions.Warmth)

// Update personality
_, err := client.Agents.Personality.Update(ctx, "agent-id", sonzai.UpdatePersonalityParams{
	Big5: sonzai.Big5Scores{Openness: 0.9, Conscientiousness: 0.7},
})
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

### Voice

```go
// Text-to-Speech
tts, err := client.Voice.TextToSpeech(ctx, sonzai.TextToSpeechParams{
	AgentID: "agent-id", Text: "Hello world!", VoiceName: "alloy",
})

// Match personality to voice
match, err := client.Voice.VoiceMatch(ctx, sonzai.VoiceMatchParams{
	Big5: sonzai.Big5Scores{Openness: 0.8, Extraversion: 0.7},
})

// List available voices
voices, err := client.Voice.ListVoices(ctx, "") // or "female", "male"

// Single-turn voice chat (STT -> LLM -> TTS)
voiceResult, err := client.Voice.VoiceChat(ctx, sonzai.VoiceChatParams{
	AgentID: "agent-id", UserID: "user-123", AudioFormat: "wav",
	Audio: audioBytes,
})
```

### Content Generation

```go
// Generate bio
bio, err := client.Agents.Generation.GenerateBio(ctx, "agent-id", sonzai.GenerateBioParams{
	UserID: "user-123", Style: "poetic",
})

// Generate image
img, err := client.Agents.Generation.GenerateImage(ctx, "agent-id", sonzai.GenerateImageParams{
	Prompt: "A cute cat in a garden",
})

// Generate full character profile
char, err := client.Agents.Generation.GenerateCharacter(ctx, "agent-id", sonzai.GenerateCharacterParams{
	Name: "Luna", Gender: "female", Description: "A warm and curious companion",
})

// Generate seed memories via LLM
seeds, err := client.Agents.Generation.GenerateSeedMemories(ctx, "agent-id", sonzai.GenerateSeedMemoriesParams{
	AgentName: "Luna", Big5: sonzai.Big5Scores{Openness: 0.8},
	GenerateOriginStory: true, StoreMemories: true,
})
```

### Dialogue

```go
dialogue, err := client.Agents.Dialogue(ctx, "agent-id", sonzai.AgentDialogueParams{
	UserID: "user-123", SceneGuidance: "casual greeting at a cafe",
})
fmt.Println(dialogue.Response)
```

### Game Events

```go
event, err := client.Agents.TriggerGameEvent(ctx, "agent-id", sonzai.TriggerGameEventParams{
	UserID:           "user-123",
	EventType:        "achievement",
	EventDescription: "Player reached level 10",
	Metadata:         map[string]string{"level": "10"},
})
```

### Custom States

```go
// List
states, err := client.Agents.CustomStates.List(ctx, "agent-id")

// Create
state, err := client.Agents.CustomStates.Create(ctx, "agent-id", sonzai.CustomStateCreateParams{
	Key: "player_level", Value: 10,
})

// Update
_, err := client.Agents.CustomStates.Update(ctx, "agent-id", "state-id", sonzai.CustomStateUpdateParams{
	Value: 15,
})

// Delete
err := client.Agents.CustomStates.Delete(ctx, "agent-id", "state-id")
```

### Context Engine Data

```go
mood, err := client.Agents.GetMood(ctx, "agent-id", "user-123", "")
moodHistory, err := client.Agents.GetMoodHistory(ctx, "agent-id", "user-123", "")
moodAggregate, err := client.Agents.GetMoodAggregate(ctx, "agent-id", "user-123", "")
relationships, err := client.Agents.GetRelationships(ctx, "agent-id", "user-123", "")
habits, err := client.Agents.GetHabits(ctx, "agent-id", "", "")
goals, err := client.Agents.GetGoals(ctx, "agent-id", "", "")
interests, err := client.Agents.GetInterests(ctx, "agent-id", "", "")
diary, err := client.Agents.GetDiary(ctx, "agent-id", "", "")
users, err := client.Agents.GetUsers(ctx, "agent-id")
constellation, err := client.Agents.GetConstellation(ctx, "agent-id", "user-123", "")
breakthroughs, err := client.Agents.GetBreakthroughs(ctx, "agent-id", "user-123", "")
wakeups, err := client.Agents.GetWakeups(ctx, "agent-id", "user-123", "")
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
