// Example: create an agent, chat with it, then clean up.
//
// Usage:
//
//	export SONZAI_API_KEY=sk-...
//	go run ./examples/agent-lifecycle
package main

import (
	"context"
	"fmt"
	"os"

	sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
	client := sonzai.MustNewClient("")
	ctx := context.Background()

	// Create an agent
	agent, err := client.Agents.Create(ctx, sonzai.CreateAgentOptions{
		Name:   "Aria",
		Gender: "female",
		Bio:    "A curious explorer who loves stargazing and telling stories.",
		Big5: &sonzai.Big5Scores{
			Openness: 85, Conscientiousness: 60, Extraversion: 70,
			Agreeableness: 75, Neuroticism: 25,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created agent: %s (%s)\n\n", agent.Name, agent.AgentID)

	// Chat with the agent
	resp, err := client.Agents.Chat(ctx, sonzai.AgentChatParams{
		AgentID: agent.AgentID,
		ChatOptions: sonzai.ChatOptions{
			Messages: []sonzai.ChatMessage{{Role: "user", Content: "What's your favorite constellation?"}},
			UserID:   "demo-user",
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "chat error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Aria: %s\n\n", resp.Content)

	// Check personality
	personality, err := client.Agents.Personality.Get(ctx, agent.AgentID, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "personality error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Openness: %.0f\n", personality.Profile.Big5.Openness.Score)

	// Clean up
	if err := client.Agents.Delete(ctx, agent.AgentID); err != nil {
		fmt.Fprintf(os.Stderr, "delete error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deleted agent %s\n", agent.AgentID)
}
