# Sonzai Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/sonz-ai/sonzai-go.svg)](https://pkg.go.dev/github.com/sonz-ai/sonzai-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonz-ai/sonzai-go)](https://goreportcard.com/report/github.com/sonz-ai/sonzai-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Go SDK for the [Sonzai Mind Layer API](https://sonz.ai). Build AI agents with persistent memory, evolving personality, and real-time voice.

Zero dependencies. Uses only the Go standard library.

## Staying in sync with the production API

This SDK tracks `https://api.sonz.ai/docs/openapi.json`. A git pre-push hook
checks for drift; run `just install-hooks` once after cloning. To refresh the
committed spec snapshot, run `just sync-spec` and commit the diff.

## Documentation

Full API documentation is available at [sonz.ai/docs](https://sonz.ai/docs).

See the [api reference](api.md) for a complete list of methods and types.

## Installation

```go
import (
    sonzai "github.com/sonz-ai/sonzai-go" // imported as sonzai
)
```

```bash
go get github.com/sonz-ai/sonzai-go@v1.2.3
```

## Getting Started

```go
package main

import (
    "context"
    "fmt"

    sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
    client := sonzai.NewClient("my-api-key") // defaults to os.LookupEnv("SONZAI_API_KEY")

    // Stream a chat response
    err := client.Agents.ChatStream(
        context.Background(),
        "agent-id",
        sonzai.ChatOptions{
            Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
            UserID:   "user-123",
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

See the [examples](examples/) directory for more.

> **Using the OpenClaw plugin?** Your API key is stored in `openclaw.json` — no environment variables needed. See the [@sonzai-labs/openclaw-context](../sonzai-openclaw/) docs.

## Evaluation & Simulation

```go
import "github.com/sonz-ai/sonzai-go/eval"

// Synchronous evaluation
result, _ := client.Eval.Evaluate(ctx, "agent-id", eval.EvaluateOptions{
    Messages:   []eval.Message{{Role: "user", Content: "Hello"}},
    TemplateID: "template-uuid",
})
fmt.Printf("Score: %.2f\n", result.Score)

// Simulation (streaming — launches run, then streams events)
ref, _ := client.Eval.Simulate(ctx, "agent-id", eval.SimulateOptions{
    UserPersona: map[string]any{"name": "Alex", "background": "Student"},
    Config:      map[string]any{"max_sessions": 2, "max_turns_per_session": 5},
}, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})

// Fire-and-forget (returns RunRef immediately)
ref, _ := client.Eval.SimulateAsync(ctx, "agent-id", eval.SimulateOptions{
    Config: map[string]any{"max_sessions": 2},
})
fmt.Printf("Run started: %s\n", ref.RunID)

// Reconnect to stream later
client.Eval.Runs.StreamEvents(ctx, ref.RunID, 0, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})

// Run eval (simulation + evaluation combined)
ref, _ = client.Eval.Run(ctx, "agent-id", eval.RunEvalOptions{
    TemplateID: "template-uuid",
    SimulationConfig: map[string]any{"max_sessions": 2},
}, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})

// Re-evaluate an existing run
ref, _ = client.Eval.ReEval(ctx, "agent-id", eval.ReEvalOptions{
    TemplateID:  "new-template-uuid",
    SourceRunID: "existing-run-uuid",
}, func(event eval.SimulationEvent) error {
    fmt.Printf("[%s] %s\n", event.Type, event.Message)
    return nil
})
```

## Custom States

```go
// Create a custom state
state, _ := client.CustomStates.Create(ctx, "agent-id", sonzai.CustomStateCreateOptions{
    Key:         "player_level",
    Value:       map[string]any{"level": 15, "xp": 2400},
    Scope:       "user",
    ContentType: "json",
    UserID:      "user-123",
})

// Upsert by composite key (create or update)
state, _ = client.CustomStates.Upsert(ctx, "agent-id", sonzai.CustomStateUpsertOptions{
    Key:    "player_level",
    Value:  map[string]any{"level": 16, "xp": 3000},
    Scope:  "user",
    UserID: "user-123",
})

// Get by composite key
state, _ = client.CustomStates.GetByKey(ctx, "agent-id", sonzai.CustomStateGetByKeyOptions{
    Key:    "player_level",
    Scope:  "user",
    UserID: "user-123",
})

// Delete by composite key
client.CustomStates.DeleteByKey(ctx, "agent-id", sonzai.CustomStateDeleteByKeyOptions{
    Key:    "player_level",
    Scope:  "user",
    UserID: "user-123",
})
```

## Requirements

This library requires Go 1.22 or later.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT - see [LICENSE](LICENSE) for details.
