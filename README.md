# Sonzai Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/sonz-ai/sonzai-go.svg)](https://pkg.go.dev/github.com/sonz-ai/sonzai-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonz-ai/sonzai-go)](https://goreportcard.com/report/github.com/sonz-ai/sonzai-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Go SDK for the [Sonzai Mind Layer API](https://sonz.ai). Build AI agents with persistent memory, evolving personality, proactive behaviors, and real-time voice.

Zero dependencies — Go standard library only.

## Requirements

- Go 1.25 or later

## Installation

```bash
go get github.com/sonz-ai/sonzai-go@v1.3.0
```

```go
import sonzai "github.com/sonz-ai/sonzai-go"
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
    // Reads SONZAI_API_KEY from env if the first arg is empty.
    client := sonzai.MustNewClient("")
    ctx := context.Background()

    // Stream a chat response
    err := client.Agents.ChatStream(ctx,
        sonzai.AgentChatParams{
            AgentID: "your-agent-id",
            ChatOptions: sonzai.ChatOptions{
                Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
                UserID:   "user-123",
            },
        },
        func(event sonzai.ChatStreamEvent) error {
            fmt.Print(event.Content())
            return nil
        },
    )
    if err != nil {
        panic(err)
    }
}
```

See the [examples/](examples/) directory for complete, runnable programs covering chat, agent lifecycle, memory, voice, and evaluation.

> **Using the OpenClaw plugin?** Your API key is stored in `openclaw.json` — no environment variables needed. See the [@sonzai-labs/openclaw-context](../sonzai-openclaw/) docs.

## Client & Configuration

```go
// Constructor with API key (falls back to SONZAI_API_KEY env var if empty)
client, err := sonzai.NewClient("")

// Or panic-on-error variant (for CLI tools and tests)
client := sonzai.MustNewClient("")

// Options
client := sonzai.MustNewClient("",
    sonzai.WithBaseURL("https://api.sonz.ai"),      // or SONZAI_BASE_URL env var
    sonzai.WithTimeout(60*time.Second),              // default 30s
    sonzai.WithHTTPClient(&http.Client{...}),        // custom transport / proxy / mTLS
)
```

Idempotent requests (GET, DELETE) automatically retry up to 3 times with exponential backoff on 429/502/503/504. POST/PUT/PATCH are not retried.

## Authentication

Get your API key from the [Sonzai Dashboard](https://platform.sonz.ai) under **Projects > API Keys**. Set it via argument or environment variable:

```bash
export SONZAI_API_KEY=sk-...
```

## Resources

The `Client` exposes these top-level resource groups:

| Resource | Purpose |
|---|---|
| `Agents` | Agent CRUD, chat, dialogue, context engine state |
| `Knowledge` | Project-scoped knowledge base (docs, graph, facts, search) |
| `Eval` | Evaluations, simulations, eval runs, templates (sub-package) |
| `Voices` | Global voice catalog |
| `Webhooks` | Webhook registration, rotation, delivery inspection |
| `ProjectConfig` / `AccountConfig` | Scoped key-value configuration |
| `CustomLLM` | Bring-your-own-model (BYOM) configuration |
| `ProjectNotifications` | Project notification polling |
| `Tenants` | Multi-tenant lookup |
| `APIKeys` | Project API key management |
| `Analytics` | Cost, usage, and real-time analytics |
| `UserPersonas` | User persona CRUD |
| `Storefront` | Agent marketplace publishing |
| `Org` | Organization billing and usage |
| `Workbench` | Internal simulation and debugging |
| `Support` | Support tickets |

Sub-resources on `client.Agents`:

```go
client.Agents.Memory        // memory tree, search, facts, timeline
client.Agents.Personality   // Big5, dimensions, deltas, overlays
client.Agents.Sessions      // session lifecycle
client.Agents.Instances     // parallel agent instances
client.Agents.Notifications // proactive notifications
client.Agents.CustomState   // scoped key-value state
client.Agents.Image         // image generation
client.Agents.Voice         // TTS, STT, live WebSocket streaming
client.Agents.Wakeups       // wakeup scheduling
client.Agents.Generation    // bio, character, seed memory generation
client.Agents.Priming       // user priming & batch import
client.Agents.Inventory     // user inventory
client.Agents.Schedules     // user-scoped recurring events
```

## Usage

### Agent CRUD

```go
agent, _ := client.Agents.Create(ctx, sonzai.CreateAgentOptions{
    Name:   "Aria",
    Gender: "female",
    Bio:    "A curious explorer who loves stargazing.",
    Big5: &sonzai.Big5Scores{
        Openness: 85, Conscientiousness: 60, Extraversion: 70,
        Agreeableness: 75, Neuroticism: 25,
    },
    // Tool capabilities are configurable at creation time:
    ToolCapabilities: &sonzai.AgentToolCapabilities{
        WebSearch:     true,
        RememberName:  true,
        KnowledgeBase: sonzai.Ptr(true), // enable project-scoped KB search
        MemoryMode:    "async",           // "sync" (default) or "async"
    },
})

fetched, _ := client.Agents.Get(ctx, agent.AgentID)
agents, _ := client.Agents.List(ctx, sonzai.ListAgentsOptions{PageSize: 20})
client.Agents.Delete(ctx, agent.AgentID)
```

### Chat (aggregated response)

```go
resp, _ := client.Agents.Chat(ctx, sonzai.AgentChatParams{
    AgentID: "agent-id",
    ChatOptions: sonzai.ChatOptions{
        Messages: []sonzai.ChatMessage{
            {Role: "user", Content: "Tell me about yourself"},
        },
        UserID:    "user-123",
        SessionID: "session-456", // auto-created if omitted
        Provider:  "gemini",
        Model:     "gemini-3.1-flash-lite-preview",
    },
})
fmt.Println(resp.Content)
fmt.Println(resp.Usage.TotalTokens)
```

### Chat (streaming)

Two streaming patterns — callback and channel.

```go
// Callback
err := client.Agents.ChatStream(ctx, params, func(event sonzai.ChatStreamEvent) error {
    fmt.Print(event.Content())
    return nil
})

// Channel
events, errCh := client.Agents.ChatStreamChannel(ctx, params)
for event := range events {
    fmt.Print(event.Content())
}
if err := <-errCh; err != nil {
    log.Fatal(err)
}
```

`ChatStreamEvent` exposes `Content()`, `IsFinished()`, `Usage`, `FinishReason`, `ExternalToolCalls`, and `SideEffectsJSON` (facts, emotions, relationship updates).

### Sync vs async memory recall (`memoryMode`)

Supplementary memory recall can be **synchronous** (blocks on recall, facts always land in the current turn) or **asynchronous** (races a deadline, slow hits spill to the next turn for lower first-token latency). Default is `sync`.

`memoryMode` is an agent-wide capability. You can set it at creation time or flip it later with `UpdateCapabilities`:

```go
// At creation
client.Agents.Create(ctx, sonzai.CreateAgentOptions{
    Name: "Aria",
    ToolCapabilities: &sonzai.AgentToolCapabilities{
        MemoryMode: "async",
    },
})

// Or flip an existing agent
client.Agents.UpdateCapabilities(ctx, "agent-id", sonzai.UpdateCapabilitiesOptions{
    MemoryMode: "async", // or "sync"
})

// Read the current mode
caps, _ := client.Agents.GetCapabilities(ctx, "agent-id")
fmt.Println(caps.MemoryMode)
```

`UpdateCapabilities` is PATCH-style — omitted fields are left unchanged. To skip the context engine entirely on a single chat (e.g. test paths), set `SkipContextBuild: true` in the chat options.

### Advanced chat options

```go
ChatOptions{
    Messages:             []sonzai.ChatMessage{...},
    UserID:               "user-123",
    UserDisplayName:      "Alex",
    SessionID:            "session-456",
    InstanceID:           "instance-789",     // parallel branch
    Provider:             "gemini",            // gemini | zhipu | volcengine | openrouter | custom
    Model:                "gemini-3.1-flash-lite-preview",
    Language:             "en",
    Timezone:             "America/New_York",
    CompiledSystemPrompt: "You are a helpful assistant.",
    ToolCapabilities:     &sonzai.AgentToolCapabilities{WebSearch: true, RememberName: true},
    ToolDefinitions:      []sonzai.ToolDefinition{ /* external function calls */ },
    MaxTurns:             10,
    GameContext:          &sonzai.GameContext{ /* custom fields */ },
    SkillLevels:          map[string]int32{"negotiation": 5},
}
```

### Memory

```go
// Full memory tree
tree, _ := client.Agents.Memory.List(ctx, "agent-id", &sonzai.MemoryListOptions{
    UserID:          "user-123",
    IncludeContents: true,
    Limit:           100,
})
for _, node := range tree.Nodes {
    fmt.Printf("%s (importance: %.2f)\n", node.Title, node.Importance)
}

// Semantic search (cosine over fact embeddings when UserID is set)
results, _ := client.Agents.Memory.Search(ctx, "agent-id", sonzai.MemorySearchOptions{
    Query:  "favorite food",
    UserID: "user-123",
    Limit:  10,
})

// Timeline across a date range
timeline, _ := client.Agents.Memory.Timeline(ctx, "agent-id", &sonzai.MemoryTimelineOptions{
    UserID: "user-123",
    Start:  "2026-01-01",
    End:    "2026-03-01",
})

// Write a fact directly
fact, _ := client.Agents.Memory.CreateFact(ctx, "agent-id", sonzai.CreateFactOptions{
    UserID:     "user-123",
    Content:    "The user loves pizza",
    FactType:   "preference",
    Importance: 0.8,
    Entities:   []string{"pizza"},
})

// Update, delete, audit
client.Agents.Memory.UpdateFact(ctx, "agent-id", fact.FactID, sonzai.UpdateFactOptions{Importance: 0.9})
history, _ := client.Agents.Memory.FactHistory(ctx, "agent-id", fact.FactID)
client.Agents.Memory.DeleteFact(ctx, "agent-id", fact.FactID)

// Seed memories in bulk at agent creation
client.Agents.Memory.Seed(ctx, "agent-id", sonzai.SeedMemoryOptions{
    UserID:   "user-123",
    Memories: []sonzai.SeedMemory{ /* ... */ },
})

// Reset all memory for a user
client.Agents.Memory.Reset(ctx, "agent-id", sonzai.MemoryResetOptions{UserID: "user-123"})
```

### Personality

```go
profile, _ := client.Agents.Personality.Get(ctx, "agent-id")
fmt.Println(profile.Profile.Big5.Openness)
fmt.Println(profile.Profile.Dimensions.Warmth)

// Recent personality shifts
shifts, _ := client.Agents.Personality.RecentShifts(ctx, "agent-id")

// Per-user overlays (how the agent perceives a specific user)
overlay, _ := client.Agents.Personality.GetOverlay(ctx, "agent-id", "user-123")
```

### Sessions & instances

```go
client.Agents.Sessions.Start(ctx, "agent-id", sonzai.SessionStartOptions{
    UserID:    "user-123",
    SessionID: "session-456",
})

client.Agents.Sessions.End(ctx, "agent-id", sonzai.SessionEndOptions{
    UserID:          "user-123",
    SessionID:       "session-456",
    TotalMessages:   10,
    DurationSeconds: 300,
})

// Parallel agent instances for A/B testing or sandboxed forks
instance, _ := client.Agents.Instances.Create(ctx, "agent-id", sonzai.InstanceCreateOptions{Name: "Beta"})
client.Agents.Instances.Reset(ctx, "agent-id", instance.InstanceID)
client.Agents.Instances.Delete(ctx, "agent-id", instance.InstanceID)
```

### Context engine introspection

```go
mood, _          := client.Agents.GetMood(ctx, "agent-id", sonzai.GetMoodOptions{UserID: "user-123"})
relationships, _ := client.Agents.GetRelationships(ctx, "agent-id")
habits, _        := client.Agents.ListHabits(ctx, "agent-id")
goals, _         := client.Agents.ListGoals(ctx, "agent-id")
interests, _     := client.Agents.GetInterests(ctx, "agent-id")
diary, _         := client.Agents.GetDiary(ctx, "agent-id")
users, _         := client.Agents.GetUsers(ctx, "agent-id")
breakthroughs, _ := client.Agents.GetBreakthroughs(ctx, "agent-id")

// Point-in-time personality snapshot
snapshot, _ := client.Agents.GetTimeMachine(ctx, "agent-id", "2026-01-15T00:00:00Z")
```

### Custom state (scoped key-value)

```go
state, _ := client.Agents.CustomState.Create(ctx, "agent-id", sonzai.CustomStateCreateOptions{
    Key:         "player_level",
    Value:       map[string]any{"level": 15, "xp": 2400},
    Scope:       "user",       // or "global"
    ContentType: "json",
    UserID:      "user-123",
})

// Upsert by composite key
client.Agents.CustomState.Upsert(ctx, "agent-id", sonzai.CustomStateUpsertOptions{
    Key: "player_level", Scope: "user", UserID: "user-123",
    Value: map[string]any{"level": 16},
})

client.Agents.CustomState.GetByKey(ctx, "agent-id", sonzai.CustomStateGetByKeyOptions{
    Key: "player_level", Scope: "user", UserID: "user-123",
})
client.Agents.CustomState.DeleteByKey(ctx, "agent-id", sonzai.CustomStateDeleteByKeyOptions{
    Key: "player_level", Scope: "user", UserID: "user-123",
})
```

### Notifications

```go
pending, _ := client.Agents.Notifications.List(ctx, "agent-id", sonzai.NotificationListOptions{
    Status: "pending",
    UserID: "user-123",
    Limit:  20,
})
for _, n := range pending.Notifications {
    fmt.Printf("[%s] %s\n", n.CheckType, n.GeneratedMessage)
}

client.Agents.Notifications.Consume(ctx, "agent-id", "message-id")
history, _ := client.Agents.Notifications.History(ctx, "agent-id", 50)
```

### Knowledge base

```go
// Upload a document (PDF, markdown, etc.)
file, _ := os.Open("document.pdf")
defer file.Close()
doc, _ := client.Knowledge.UploadDocument(ctx, "project-id", "document.pdf", file, "application/pdf")

docs, _ := client.Knowledge.ListDocuments(ctx, "project-id", 100)
client.Knowledge.GetDocument(ctx, "project-id", doc.DocumentID)
client.Knowledge.DeleteDocument(ctx, "project-id", doc.DocumentID)

// Direct fact/graph insertion
client.Knowledge.InsertFacts(ctx, "project-id", sonzai.KnowledgeFactBatch{
    Entities:      []sonzai.KnowledgeEntity{ /* ... */ },
    Relationships: []sonzai.KnowledgeRelationship{ /* ... */ },
})

// Cross-document semantic search
results, _ := client.Knowledge.Search(ctx, "project-id", sonzai.KBSearchOptions{
    Query: "how do I configure X?",
    Limit: 10,
})
for _, r := range results.Results {
    fmt.Printf("[%.2f] %s (type: %s)\n", r.Score, r.Label, r.NodeType)
}
```

### Voice (TTS, STT, live streaming)

```go
// Text-to-Speech
audio, _ := client.Agents.Voice.TTS(ctx, "agent-id", sonzai.TTSOptions{
    Text:      "Hello, how are you?",
    VoiceName: "Kore",
    Language:  "en-US",
})

// Speech-to-Text
transcript, _ := client.Agents.Voice.STT(ctx, "agent-id", sonzai.STTOptions{
    Audio:       base64.StdEncoding.EncodeToString(pcmBytes),
    AudioFormat: "pcm",
    Language:    "en-US",
})

// Live bidirectional voice (WebSocket, Gemini Live)
token, _ := client.Agents.Voice.GetToken(ctx, "agent-id", sonzai.VoiceTokenOptions{
    VoiceName: "Kore",
    Language:  "en-US",
    UserID:    "user-123",
})

stream, _ := client.Agents.Voice.Stream(ctx, token)
defer stream.Close()

stream.SendText("Hello!")
// or: stream.SendAudio(pcm16kHzMonoBytes)

for {
    event, err := stream.Recv()
    if err == io.EOF { break }
    switch event.Type {
    case "input_transcript":
        fmt.Println("User:", event.Text)
    case "output_transcript":
        fmt.Println("Agent:", event.Text)
    case "audio":
        playPCM(event.Audio) // 24kHz PCM
    case "turn_complete":
    case "session_ended":
        return
    }
}
```

### Evaluation & simulation

```go
import "github.com/sonz-ai/sonzai-go/eval"

// One-off evaluation over a conversation
result, _ := client.Eval.Evaluate(ctx, "agent-id", eval.EvaluateOptions{
    Messages:   []eval.Message{{Role: "user", Content: "Hello"}},
    TemplateID: "template-uuid",
})
fmt.Printf("Score: %.2f\n", result.Score)

// Streaming simulation
client.Eval.Simulate(ctx, "agent-id", eval.SimulateOptions{
    UserPersona: map[string]any{"name": "Alex", "background": "Student"},
    Config:      map[string]any{"max_sessions": 2, "max_turns_per_session": 5},
}, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})

// Fire-and-forget
ref, _ := client.Eval.SimulateAsync(ctx, "agent-id", eval.SimulateOptions{
    Config: map[string]any{"max_sessions": 2},
})
// Reconnect later (from_index supports resume)
client.Eval.Runs.StreamEvents(ctx, ref.RunID, 0, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})

// Combined simulation + evaluation
client.Eval.Run(ctx, "agent-id", eval.RunEvalOptions{
    TemplateID:       "template-uuid",
    SimulationConfig: map[string]any{"max_sessions": 2},
}, handler)

// Re-evaluate an existing run with a different template
client.Eval.ReEval(ctx, "agent-id", eval.ReEvalOptions{
    TemplateID:  "new-template-uuid",
    SourceRunID: "existing-run-uuid",
}, handler)

// Template & run management
tmpl, _ := client.Eval.Templates.Create(ctx, eval.CreateTemplateOptions{Name: "Empathy", ScoringRubric: "..."})
runs, _ := client.Eval.Runs.List(ctx, eval.ListRunsOptions{AgentID: "agent-id", Limit: 20})
run, _ := client.Eval.Runs.Get(ctx, "run-id")
client.Eval.Runs.Delete(ctx, "run-id")
```

### Webhooks

```go
// Register/update a webhook
resp, _ := client.Webhooks.Register(ctx, "agent.message.created", sonzai.WebhookRegisterOptions{
    WebhookURL: "https://example.com/hook",
    AuthHeader: "Bearer your-secret", // optional header added to deliveries
})
fmt.Println("signing secret:", resp.SigningSecret)

list, _       := client.Webhooks.List(ctx)
attempts, _   := client.Webhooks.ListDeliveryAttempts(ctx, "agent.message.created")
rotated, _    := client.Webhooks.RotateSecret(ctx, "agent.message.created")
client.Webhooks.Delete(ctx, "agent.message.created")

// Project-scoped variants
client.Webhooks.RegisterForProject(ctx, "project-id", "event.type", opts)
client.Webhooks.ListForProject(ctx, "project-id")
client.Webhooks.DeleteForProject(ctx, "project-id", "event.type")
```

### Webhook signature verification

Signatures are HMAC-SHA256 over `"{timestamp}.{payload}"`. Multiple `v1=` signatures are supported for secret rotation.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("Sonzai-Signature")

    if err := sonzai.VerifyWebhookSignature(body, sig, webhookSecret); err != nil {
        switch {
        case errors.Is(err, sonzai.ErrInvalidSignature):
            http.Error(w, "bad signature", 401)
        case errors.Is(err, sonzai.ErrTimestampExpired):
            http.Error(w, "expired", 401)
        default:
            http.Error(w, "unauthorized", 401)
        }
        return
    }

    // Accept tighter/looser timestamp windows:
    sonzai.VerifyWebhookSignatureWithTolerance(body, sig, secret, 1*time.Minute)
}
```

## Error handling

```go
resp, err := client.Agents.Chat(ctx, params)
if err != nil {
    var (
        authErr      *sonzai.AuthenticationError
        notFoundErr  *sonzai.NotFoundError
        badReqErr    *sonzai.BadRequestError
        permErr      *sonzai.PermissionDeniedError
        rateErr      *sonzai.RateLimitError
        serverErr    *sonzai.InternalServerError
    )
    switch {
    case errors.As(err, &authErr):
        log.Fatal("invalid API key")
    case errors.As(err, &notFoundErr):
        log.Fatal("agent not found")
    case errors.As(err, &rateErr):
        if rateErr.RetryAfter != nil {
            log.Printf("rate limited, retry after %ds", *rateErr.RetryAfter)
        }
    case errors.As(err, &serverErr):
        log.Fatal("server error")
    }
}
```

All typed errors embed `sonzai.SonzaiError` (exposes `StatusCode` and `Message`).

## Staying in sync with the production API

This SDK tracks `https://api.sonz.ai/docs/openapi.json`. A git pre-push hook
checks for drift; run `just install-hooks` once after cloning. To refresh the
committed spec snapshot, run `just sync-spec` and commit the diff.

## Documentation

- [Sonzai docs](https://sonz.ai/docs)
- [API reference](api.md)
- [Package docs on pkg.go.dev](https://pkg.go.dev/github.com/sonz-ai/sonzai-go)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](LICENSE) for details.
