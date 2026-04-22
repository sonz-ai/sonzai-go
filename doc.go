// Package sonzai provides the official Go SDK for the Sonzai Mind Layer API.
//
// The SDK enables you to build AI agents with persistent memory, evolving
// personality, and proactive behaviors. It requires zero external dependencies
// and uses only the Go standard library.
//
// # Getting Started
//
// Create a client with your API key:
//
//	client := sonzai.NewClient("sk-your-api-key")
//
// Or use the SONZAI_API_KEY environment variable:
//
//	client := sonzai.NewClient("")
//
// # Chat
//
// Send messages and receive streaming or aggregated responses:
//
//	// Non-streaming
//	resp, err := client.Agents.Chat(ctx, sonzai.AgentChatParams{
//	    AgentID: "agent-id",
//	    ChatOptions: sonzai.ChatOptions{
//	        Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
//	        UserID:   "user-123",
//	    },
//	})
//
//	// Streaming
//	err := client.Agents.ChatStream(ctx, sonzai.AgentChatParams{
//	    AgentID: "agent-id",
//	    ChatOptions: sonzai.ChatOptions{
//	        Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
//	    },
//	}, func(event sonzai.ChatStreamEvent) error {
//	    fmt.Print(event.Content())
//	    return nil
//	})
//
// # Resources
//
// The client exposes the following resource groups:
//
//   - client.Agents — Chat, context engine data, and agent-scoped operations
//   - client.Agents.Memory — Memory tree, search, and timeline
//   - client.Agents.Personality — Personality profile and evolution history
//   - client.Agents.Sessions — Session lifecycle and tool configuration
//   - client.Agents.Instances — Agent instance CRUD and reset
//   - client.Agents.Notifications — Proactive notification management
//   - client.Agents.CustomState — Arbitrary key-value state storage
//   - client.Agents.Image — Image generation
//   - client.Agents.Priming — User priming, metadata, and batch import
//   - client.Knowledge — Project-scoped knowledge base (documents, graph, schemas, search, analytics)
//   - client.Eval — Evaluation, simulation, and benchmarking (separate sub-package)
//   - client.Eval.Templates — Evaluation template CRUD
//   - client.Eval.Runs — Evaluation run management
//
// # Error Handling
//
// The SDK returns typed errors for different failure scenarios:
//
//	resp, err := client.Agents.Chat(ctx, opts)
//	if err != nil {
//	    var authErr *sonzai.AuthenticationError
//	    if errors.As(err, &authErr) {
//	        log.Fatal("Invalid API key")
//	    }
//	}
package sonzai

// Ptr returns a pointer to v. Convenience helper for option structs that use
// pointer fields to distinguish "not set" from zero value (e.g.
// UpdateCapabilitiesOptions.WebSearch, AgentToolCapabilities.KnowledgeBase).
//
//	opts := sonzai.UpdateCapabilitiesOptions{WebSearch: sonzai.Ptr(true)}
func Ptr[T any](v T) *T { return &v }
