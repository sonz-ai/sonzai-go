package sonzai

import (
	"context"
	"fmt"
)

// MCPCatalogResource provides operations on the per-project MCP catalog —
// the registry of MCP (Model Context Protocol) server endpoints a project
// can expose to its agents.
//
// Each catalog entry pairs a remote MCP URL with auth config; agents opt
// into specific entries via the `mcpEnabled` capability (a list of catalog
// IDs). At chat time, the platform discovers the entry's tools and registers
// them as agent-callable tools for the turn.
//
// Org-admin only on the platform side: non-admin callers get 403 on
// create / update / delete; reads are open within the project.
type MCPCatalogResource struct {
	http *httpClient
}

// MCPAuthKind enumerates the auth schemes a catalog entry can use to talk
// to its remote MCP server.
type MCPAuthKind string

const (
	MCPAuthNone   MCPAuthKind = "none"
	MCPAuthBearer MCPAuthKind = "bearer"
	MCPAuthHeader MCPAuthKind = "header"
	MCPAuthOAuth  MCPAuthKind = "oauth"
)

// MCPTransport enumerates the MCP transport protocols.
type MCPTransport string

const (
	MCPTransportStreamableHTTP MCPTransport = "streamable-http"
	MCPTransportSSE            MCPTransport = "sse"
)

// MCPCatalogAuth carries the auth config for a catalog entry. Read-side
// fields (BearerSecretRef / HeaderSecretRef) are sm:// pointers; write-side
// fields (BearerToken / HeaderValue) are the plaintext values you submit on
// create / update — they are never returned.
type MCPCatalogAuth struct {
	Kind             MCPAuthKind `json:"kind"`
	BearerToken      string      `json:"bearer_token,omitempty"`
	BearerSecretRef  string      `json:"bearer_secret_ref,omitempty"`
	HeaderName       string      `json:"header_name,omitempty"`
	HeaderValue      string      `json:"header_value,omitempty"`
	HeaderSecretRef  string      `json:"header_secret_ref,omitempty"`
}

// MCPHealth summarises the platform's last contact with the MCP server.
// Surfaced on every catalog read so admins can see which entries are
// failing without probing them.
type MCPHealth struct {
	Status              string `json:"status"`
	LastOKAt            string `json:"last_ok_at,omitempty"`
	LastErrorAt         string `json:"last_error_at,omitempty"`
	LastError           string `json:"last_error,omitempty"`
	LastAuthRejectedAt  string `json:"last_auth_rejected_at,omitempty"`
	Failures1h          int    `json:"failures_1h"`
}

// MCPCatalogEntry is a project-scoped registration of an MCP server.
type MCPCatalogEntry struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	Transport   MCPTransport   `json:"transport"`
	Auth        MCPCatalogAuth `json:"auth"`
	Health      MCPHealth      `json:"health"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// MCPCatalogTool describes one tool advertised by an MCP server, as
// discovered the last time the platform probed the entry.
type MCPCatalogTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// MCPProbeResponse is the result of a probe call — either the synchronous
// /probe (testing a config before persisting it) or the per-entry probe
// that re-fetches the tool list.
type MCPProbeResponse struct {
	OK        bool   `json:"ok"`
	LatencyMs int    `json:"latency_ms"`
	ToolCount int    `json:"tool_count"`
	Error     string `json:"error,omitempty"`
}

// CreateMCPCatalogOptions configures a catalog create request. All four
// fields plus the URL must be HTTPS — the platform rejects http:// URLs.
type CreateMCPCatalogOptions struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url"`
	Transport   MCPTransport   `json:"transport"`
	Auth        MCPCatalogAuth `json:"auth"`
}

// UpdateMCPCatalogOptions configures a catalog update. All fields are
// optional — only non-empty / non-nil values are sent. Pass *MCPCatalogAuth
// (rather than embedded struct) so callers can leave auth untouched.
type UpdateMCPCatalogOptions struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Transport   MCPTransport    `json:"transport,omitempty"`
	Auth        *MCPCatalogAuth `json:"auth,omitempty"`
}

// List returns every MCP catalog entry registered for the project.
func (m *MCPCatalogResource) List(ctx context.Context, projectID string) ([]MCPCatalogEntry, error) {
	var result struct {
		Entries []MCPCatalogEntry `json:"entries"`
	}
	if err := m.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog", projectID), nil, &result); err != nil {
		return nil, err
	}
	return result.Entries, nil
}

// Get returns a single catalog entry by ID.
func (m *MCPCatalogResource) Get(ctx context.Context, projectID, id string) (*MCPCatalogEntry, error) {
	var entry MCPCatalogEntry
	if err := m.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s", projectID, id), nil, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Create registers a new MCP server with the project's catalog. Org-admin only.
func (m *MCPCatalogResource) Create(ctx context.Context, projectID string, opts CreateMCPCatalogOptions) (*MCPCatalogEntry, error) {
	var entry MCPCatalogEntry
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog", projectID), opts, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Update patches an existing catalog entry. Only non-empty fields are sent.
// Auth is updated atomically — pass a fresh MCPCatalogAuth to rotate creds.
// Org-admin only.
func (m *MCPCatalogResource) Update(ctx context.Context, projectID, id string, opts UpdateMCPCatalogOptions) (*MCPCatalogEntry, error) {
	var entry MCPCatalogEntry
	if err := m.http.Patch(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s", projectID, id), opts, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Delete removes a catalog entry. Agents that had this ID in their
// `mcpEnabled` capability lose access to the server's tools immediately;
// the agent's mcpEnabled list is left unchanged (the platform tolerates
// dangling IDs and skips them at tool-decl time). Org-admin only.
func (m *MCPCatalogResource) Delete(ctx context.Context, projectID, id string) error {
	return m.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s", projectID, id), nil)
}

// ProbeConfig synchronously contacts the MCP server described by `opts`
// without persisting anything. Use this from a "Test connection" button
// before saving — the response includes latency, tool count, and any
// auth/transport error the platform observed.
func (m *MCPCatalogResource) ProbeConfig(ctx context.Context, projectID string, opts CreateMCPCatalogOptions) (*MCPProbeResponse, error) {
	var probe MCPProbeResponse
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/probe", projectID), opts, &probe); err != nil {
		return nil, err
	}
	return &probe, nil
}

// Probe re-contacts an existing catalog entry's MCP server and refreshes
// its tool list + health summary. Call this after the upstream MCP
// deployment has changed (added/removed tools, rotated auth) without
// having to wait for the platform's background refresh.
func (m *MCPCatalogResource) Probe(ctx context.Context, projectID, id string) (*MCPProbeResponse, error) {
	var probe MCPProbeResponse
	if err := m.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s/probe", projectID, id), nil, &probe); err != nil {
		return nil, err
	}
	return &probe, nil
}

// ListTools returns the tools advertised by this catalog entry's MCP
// server, as captured the last time the platform probed it. Useful for
// dashboard pickers that show "which tools does this MCP expose?".
func (m *MCPCatalogResource) ListTools(ctx context.Context, projectID, id string) ([]MCPCatalogTool, error) {
	var result struct {
		Tools []MCPCatalogTool `json:"tools"`
	}
	if err := m.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s/tools", projectID, id), nil, &result); err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// ListUsages returns every agent ID currently opted into this catalog
// entry via its `mcpEnabled` capability. Useful for "who would I break
// if I delete this entry?" diligence before removing it.
func (m *MCPCatalogResource) ListUsages(ctx context.Context, projectID, id string) ([]string, error) {
	var result struct {
		AgentIDs []string `json:"agent_ids"`
	}
	if err := m.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/mcp/catalog/%s/usages", projectID, id), nil, &result); err != nil {
		return nil, err
	}
	return result.AgentIDs, nil
}
