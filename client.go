// Package sonzai provides a Go SDK for the Sonzai Mind Layer API.
//
// Usage:
//
//	client := sonzai.NewClient("your-api-key")
//
//	// Chat with an agent
//	resp, err := client.Agents.Chat(ctx, sonzai.AgentChatParams{
//	    AgentID:     "agent-id",
//	    ChatOptions: sonzai.ChatOptions{Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}}},
//	})
//
//	// Stream chat
//	err := client.Agents.ChatStream(ctx, sonzai.AgentChatParams{
//	    AgentID:     "agent-id",
//	    ChatOptions: sonzai.ChatOptions{Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}}},
//	}, func(event sonzai.ChatStreamEvent) error {
//	    fmt.Print(event.Content())
//	    return nil
//	})
//
//	// Evaluate agent quality
//	result, err := client.Eval.Evaluate(ctx, "agent-id", eval.EvaluateOptions{...})
package sonzai

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/sonz-ai/sonzai-go/eval"
)

const (
	defaultBaseURL = "https://api.sonz.ai"
	defaultTimeout = 30 * time.Second
)

// ClientOption configures the Sonzai client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// WithBaseURL sets the API base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) { c.baseURL = url }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.timeout = d }
}

// WithHTTPClient sets a custom HTTP client. When provided, the SDK uses this
// client instead of creating a new one. The caller is responsible for setting
// timeouts and transport configuration.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) { c.httpClient = client }
}

// Client is the Sonzai Mind Layer API client.
type Client struct {
	// Agents provides chat, memory, personality, and other agent-scoped operations.
	Agents *AgentsResource

	// Knowledge provides project-scoped knowledge base operations
	// (documents, graph nodes, schemas, search, analytics).
	Knowledge *KnowledgeResource

	// Eval provides evaluation, simulation, and benchmarking operations.
	Eval *eval.Client

	// Voices provides the global voice catalog.
	Voices *VoicesResource

	// Webhooks provides webhook registration and management.
	Webhooks *WebhooksResource

	// ProjectConfig provides project-scoped configuration management.
	ProjectConfig *ProjectConfigResource

	// CustomLLM provides project-scoped custom LLM configuration.
	CustomLLM *CustomLLMResource

	// ProjectNotifications provides project-scoped notification polling.
	ProjectNotifications *ProjectNotificationsResource

	http *httpClient
}

// ListModels returns all LLM providers and model variants enabled on this
// deployment. This is a platform-level call — it does not require an agent ID.
// Use it to populate model picker UIs or validate model IDs before a chat
// request.
//
//	result, err := client.ListModels(ctx)
//	for _, p := range result.Providers {
//	    fmt.Println(p.ProviderName, p.Models)
//	}
func (c *Client) ListModels(ctx context.Context) (*PlatformModelsResponse, error) {
	var result PlatformModelsResponse
	if err := c.http.Get(ctx, "/api/v1/models", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewClient creates a new Sonzai client with the given API key.
// If apiKey is empty, it falls back to the SONZAI_API_KEY environment variable.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("SONZAI_API_KEY")
	}
	if apiKey == "" {
		panic("sonzai: apiKey must be provided or set via SONZAI_API_KEY environment variable")
	}

	cfg := &clientConfig{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}

	if envURL := os.Getenv("SONZAI_BASE_URL"); envURL != "" {
		cfg.baseURL = envURL
	}

	for _, opt := range opts {
		opt(cfg)
	}

	hc := newHTTPClient(cfg.baseURL, apiKey, cfg.timeout, cfg.httpClient)

	return &Client{
		Agents:               newAgentsResource(hc),
		Knowledge:            &KnowledgeResource{http: hc},
		Eval:                 eval.New(hc),
		Voices:               &VoicesResource{http: hc},
		Webhooks:             &WebhooksResource{http: hc},
		ProjectConfig:        &ProjectConfigResource{http: hc},
		CustomLLM:            &CustomLLMResource{http: hc},
		ProjectNotifications: &ProjectNotificationsResource{http: hc},
		http:                 hc,
	}
}

