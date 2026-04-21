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
	"fmt"
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

	// AccountConfig provides tenant-scoped ("account-level") configuration
	// management. Use it to set defaults that apply across every project
	// inside a tenant — for example, the post-processing model map.
	AccountConfig *AccountConfigResource

	// CustomLLM provides project-scoped custom LLM configuration.
	CustomLLM *CustomLLMResource

	// ProjectNotifications provides project-scoped notification polling.
	ProjectNotifications *ProjectNotificationsResource

	// Tenants provides tenant lookup operations.
	Tenants *TenantsResource

	// APIKeys provides project API key management.
	APIKeys *APIKeysResource

	// Analytics provides platform analytics and cost reporting.
	Analytics *AnalyticsResource

	// UserPersonas provides user persona CRUD operations.
	UserPersonas *UserPersonasResource

	// Storefront provides agent marketplace (storefront) management.
	Storefront *StorefrontResource

	// Org provides organization-level billing, usage, and contract operations.
	Org *OrgResource

	// Workbench provides internal simulation and debugging operations.
	Workbench *WorkbenchResource

	// SupportTickets provides support ticket operations for the authenticated
	// user within their active tenant (list, create, get, close, comment).
	SupportTickets *SupportTicketsResource

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
// Returns an error if no API key is available (instead of panicking).
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("SONZAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("sonzai: apiKey must be provided or set via SONZAI_API_KEY environment variable")
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
		AccountConfig:        &AccountConfigResource{http: hc},
		CustomLLM:            &CustomLLMResource{http: hc},
		ProjectNotifications: &ProjectNotificationsResource{http: hc},
		Tenants:              &TenantsResource{http: hc},
		APIKeys:              &APIKeysResource{http: hc},
		Analytics:            &AnalyticsResource{http: hc},
		UserPersonas:         &UserPersonasResource{http: hc},
		Storefront:           &StorefrontResource{http: hc},
		Org:                  &OrgResource{http: hc},
		Workbench:            &WorkbenchResource{http: hc},
		SupportTickets:       &SupportTicketsResource{http: hc},
		http:                 hc,
	}, nil
}

// MustNewClient is like NewClient but panics on error.
// Use in tests or CLI tools where error handling is unnecessary.
func MustNewClient(apiKey string, opts ...ClientOption) *Client {
	c, err := NewClient(apiKey, opts...)
	if err != nil {
		panic(err)
	}
	return c
}
