// Example: run a simulation and evaluate agent quality.
//
// Usage:
//
//	export SONZAI_API_KEY=sk-...
//	go run ./examples/evaluation -agent <agent-id> [-template <template-id>]
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	sonzai "github.com/sonz-ai/sonzai-go"
	"github.com/sonz-ai/sonzai-go/eval"
)

func main() {
	agentID := flag.String("agent", "", "Agent ID to evaluate")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/evaluation -agent <agent-id>")
		os.Exit(1)
	}

	client := sonzai.MustNewClient("")
	ctx := context.Background()

	// Run simulation with live progress
	fmt.Println("Running simulation...")
	ref, err := client.Eval.Simulate(ctx, *agentID, eval.SimulateOptions{
		UserPersona: map[string]interface{}{
			"name":                "Alex",
			"background":          "College student who loves sci-fi",
			"personality_traits":  []string{"curious", "friendly"},
			"communication_style": "casual",
		},
		Config: map[string]interface{}{
			"max_sessions":          2,
			"max_turns_per_session": 5,
		},
	}, func(event eval.SimulationEvent) error {
		switch event.Type {
		case "turn_complete":
			fmt.Printf("  [Turn] %s\n", event.Message)
		case "session_complete":
			fmt.Printf("  [Session complete]\n")
		default:
			fmt.Printf("  [%s] %s\n", event.Type, event.Message)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone. Run ID: %s\n", ref.RunID)
}
