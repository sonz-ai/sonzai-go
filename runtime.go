package sonzai

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RuntimeUsageSchemaVersion is the current signed usage-report contract.
const RuntimeUsageSchemaVersion = 2

// RuntimeBillingMode declares who paid the provider for a runtime LLM call.
// Standard means Sonzai paid the provider; BYOK means the tenant paid the
// provider directly and Sonzai bills a 33% service fee.
type RuntimeBillingMode string

const (
	// RuntimeBillingModeStandard bills provider raw cost times the Sonzai
	// retail multiplier.
	RuntimeBillingModeStandard RuntimeBillingMode = "standard"
	// RuntimeBillingModeBYOK bills only the 33% Sonzai service fee because the
	// tenant paid the provider directly. BYOM runtimes use this mode too.
	RuntimeBillingModeBYOK RuntimeBillingMode = "byok"
)

// RuntimeResource is the public platform control-plane surface for building
// a Sonzai-compatible runtime. It intentionally contains no LLM completion
// method: provider inference belongs in the runtime, while Sonzai Cloud owns
// context/memory and signed usage ingestion.
type RuntimeResource struct {
	http *httpClient
}

// RuntimeBackendAgentArtifact is a versioned prompt/schema bundle a runtime
// executes with its own LLM provider. ModelHint is informational; the runtime's
// configured provider/model remains authoritative.
type RuntimeBackendAgentArtifact struct {
	Slug           string         `json:"slug"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	ModelHint      string         `json:"model_hint"`
	System         string         `json:"system"`
	FindingsSchema map[string]any `json:"findings_schema,omitempty"`
	Tools          []string       `json:"tools,omitempty"`
	DisableTools   bool           `json:"disable_tools"`
	MaxToolRounds  int            `json:"max_tool_rounds"`
	Version        string         `json:"version"`
}

type RuntimeBackendAgentArtifactList struct {
	Artifacts []RuntimeBackendAgentArtifact `json:"artifacts"`
}

// BackendAgentArtifacts downloads execution configuration only. Sonzai Cloud
// does not perform a completion for this operation.
func (r *RuntimeResource) BackendAgentArtifacts(ctx context.Context) (*RuntimeBackendAgentArtifactList, error) {
	if err := requireRuntimeResource(r); err != nil {
		return nil, err
	}
	var out RuntimeBackendAgentArtifactList
	if err := r.doJSON(ctx, http.MethodGet, "/api/v1/runtime/backend-agent-artifacts", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RuntimeAPIError preserves a runtime control-plane response's status, code,
// and body so a runtime can distinguish license gates from transient errors.
type RuntimeAPIError struct {
	StatusCode int
	Code       string
	Message    string
	Body       []byte
}

func (e *RuntimeAPIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("runtime control plane [%d] %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("runtime control plane [%d]: %s", e.StatusCode, e.Message)
}

// RuntimeContextBundleParams scopes one per-turn context build.
type RuntimeContextBundleParams struct {
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
	CurrentMessage string `json:"current_message,omitempty"`
}

// RuntimeContextBundle is the provider-ready context returned to a runtime.
type RuntimeContextBundle struct {
	SystemPromptParts []string        `json:"system_prompt_parts"`
	MemoryContext     json.RawMessage `json:"memory_context"`
	Persona           json.RawMessage `json:"persona"`
	ToolDefinitions   json.RawMessage `json:"tool_definitions"`
	TTL               int             `json:"ttl"`
}

// ContextBundle fetches memory and prompt context for local provider
// execution. It never performs an LLM completion in Sonzai Cloud.
func (r *RuntimeResource) ContextBundle(ctx context.Context, agentID string, params RuntimeContextBundleParams) (*RuntimeContextBundle, error) {
	if err := requireRuntimeResource(r); err != nil {
		return nil, err
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal runtime context bundle request: %w", err)
	}
	path := fmt.Sprintf("/api/v1/agents/%s/context-bundle", url.PathEscape(agentID))
	var out RuntimeContextBundle
	if err := r.doJSON(ctx, http.MethodPost, path, body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RuntimeConversationOptions selects a runtime session transcript page.
type RuntimeConversationOptions struct {
	UserID    string
	SessionID string
	Page      int
	PageSize  int
}

// RuntimeConversationMessage is one raw transcript message.
type RuntimeConversationMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content"`
	Timestamp  time.Time         `json:"timestamp"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []RuntimeToolCall `json:"tool_calls,omitempty"`
	Author     string            `json:"author,omitempty"`
}

// RuntimeConversation is one page from the platform session store.
type RuntimeConversation struct {
	AgentID   string                       `json:"agent_id"`
	UserID    string                       `json:"user_id"`
	SessionID string                       `json:"session_id"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"page_size"`
	Total     int                          `json:"total"`
	HasMore   bool                         `json:"has_more"`
	Messages  []RuntimeConversationMessage `json:"messages"`
}

// Conversation reads prior runtime turns from Sonzai's session store.
func (r *RuntimeResource) Conversation(ctx context.Context, agentID string, opts RuntimeConversationOptions) (*RuntimeConversation, error) {
	if err := requireRuntimeResource(r); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("user_id", opts.UserID)
	q.Set("session_id", opts.SessionID)
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PageSize > 0 {
		q.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	path := fmt.Sprintf("/api/v1/agents/%s/conversations?%s", url.PathEscape(agentID), q.Encode())
	var out RuntimeConversation
	if err := r.doJSON(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RuntimeToolCallFunction is the OpenAI-compatible function-call payload
// persisted with an externally executed turn.
type RuntimeToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// RuntimeToolCall is one tool invocation attached to an assistant message.
type RuntimeToolCall struct {
	ID       string                  `json:"id"`
	Type     string                  `json:"type"`
	Function RuntimeToolCallFunction `json:"function"`
}

// RuntimeTurnMessage is one exact transcript message executed by a runtime.
type RuntimeTurnMessage struct {
	Role       string            `json:"role"`
	Content    *string           `json:"content,omitempty"`
	Timestamp  time.Time         `json:"timestamp,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []RuntimeToolCall `json:"tool_calls,omitempty"`
	Author     string            `json:"author,omitempty"`
}

// RuntimeCompletedTurn is the paired form accepted by ReportTurns.
type RuntimeCompletedTurn struct {
	UserMessage      RuntimeTurnMessage   `json:"user_message"`
	AssistantMessage RuntimeTurnMessage   `json:"assistant_message"`
	ToolResults      []RuntimeTurnMessage `json:"tool_results,omitempty"`
}

// RuntimeTurnReport appends externally executed messages to Sonzai memory.
// Send either Messages or Turns, never both.
type RuntimeTurnReport struct {
	UserID          string                 `json:"user_id"`
	SessionID       string                 `json:"session_id"`
	InstanceID      string                 `json:"instance_id,omitempty"`
	UserDisplayName string                 `json:"user_display_name,omitempty"`
	Messages        []RuntimeTurnMessage   `json:"messages,omitempty"`
	Turns           []RuntimeCompletedTurn `json:"turns,omitempty"`
}

// RuntimeTurnReportResult acknowledges transcript persistence.
type RuntimeTurnReportResult struct {
	Accepted       bool   `json:"accepted"`
	AgentID        string `json:"agent_id"`
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
	MessagesStored int    `json:"messages_stored"`
}

// ReportTurns writes completed local-provider turns back to Sonzai memory.
func (r *RuntimeResource) ReportTurns(ctx context.Context, agentID string, report RuntimeTurnReport) (*RuntimeTurnReportResult, error) {
	if err := requireRuntimeResource(r); err != nil {
		return nil, err
	}
	body, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal runtime turn report: %w", err)
	}
	path := fmt.Sprintf("/api/v1/agents/%s/turns", url.PathEscape(agentID))
	var out RuntimeTurnReportResult
	if err := r.doJSON(ctx, http.MethodPost, path, body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RuntimeUsageCounter is one invoice-grade aggregate. TokensIn contains
// fresh input tokens only; cache hits and cache writes have separate fields.
type RuntimeUsageCounter struct {
	ProjectID           string             `json:"project_id"`
	AgentID             string             `json:"agent_id"`
	Provider            string             `json:"provider"`
	Model               string             `json:"model"`
	UseCase             string             `json:"use_case"`
	BillingMode         RuntimeBillingMode `json:"billing_mode"`
	TokensIn            int64              `json:"tokens_in"`
	TokensOut           int64              `json:"tokens_out"`
	CacheReadTokens     int64              `json:"cache_read_tokens"`
	CacheCreationTokens int64              `json:"cache_creation_tokens"`
	Turns               int64              `json:"turns"`
	UnreportedTurns     int64              `json:"unreported_turns"`
}

// RuntimeUsageReport is the signed, idempotent runtime billing envelope.
type RuntimeUsageReport struct {
	SchemaVersion int                   `json:"schema_version"`
	ReportID      string                `json:"report_id"`
	TenantID      string                `json:"tenant_id"`
	InstanceID    string                `json:"instance_id,omitempty"`
	PeriodStart   time.Time             `json:"period_start"`
	PeriodEnd     time.Time             `json:"period_end"`
	HeartbeatAt   time.Time             `json:"heartbeat_at"`
	Counters      []RuntimeUsageCounter `json:"counters"`
	Signature     string                `json:"signature"`
}

// RuntimeUsageReportResult acknowledges a signed usage report.
type RuntimeUsageReportResult struct {
	Accepted bool   `json:"accepted"`
	ReportID string `json:"report_id"`
}

// SubmitUsageReport sends a pre-signed report and mirrors the signature in
// X-Sonzai-Metering-Signature. Use SignRuntimeUsageReport before calling it.
func (r *RuntimeResource) SubmitUsageReport(ctx context.Context, report RuntimeUsageReport) (*RuntimeUsageReportResult, error) {
	if err := requireRuntimeResource(r); err != nil {
		return nil, err
	}
	body, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshal runtime usage report: %w", err)
	}
	var out RuntimeUsageReportResult
	headers := map[string]string{"X-Sonzai-Metering-Signature": report.Signature}
	if err := r.doJSON(ctx, http.MethodPost, "/api/v1/usage/reports", body, headers, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SignRuntimeUsageReport returns the lowercase hex HMAC-SHA256 signature for
// the canonical v2 report. Counter order does not affect the signature.
func SignRuntimeUsageReport(report RuntimeUsageReport, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(CanonicalRuntimeUsageReport(report)))
	return hex.EncodeToString(mac.Sum(nil))
}

// CanonicalRuntimeUsageReport returns the stable signing payload shared by
// runtimes and platform-api. It is public so non-Go custom runtimes can test
// their implementations against the same contract.
func CanonicalRuntimeUsageReport(report RuntimeUsageReport) string {
	version := report.SchemaVersion
	if version == 0 {
		version = RuntimeUsageSchemaVersion
	}
	counters := append([]RuntimeUsageCounter(nil), report.Counters...)
	sort.SliceStable(counters, func(i, j int) bool {
		return runtimeUsageCounterSortKey(counters[i]) < runtimeUsageCounterSortKey(counters[j])
	})
	var b strings.Builder
	b.WriteString(strconv.Itoa(version))
	b.WriteByte('\n')
	b.WriteString(report.ReportID)
	b.WriteByte('\n')
	b.WriteString(report.TenantID)
	b.WriteByte('\n')
	b.WriteString(report.InstanceID)
	b.WriteByte('\n')
	b.WriteString(report.PeriodStart.UTC().Format(time.RFC3339Nano))
	b.WriteByte('\n')
	b.WriteString(report.PeriodEnd.UTC().Format(time.RFC3339Nano))
	b.WriteByte('\n')
	b.WriteString(report.HeartbeatAt.UTC().Format(time.RFC3339Nano))
	for _, c := range counters {
		b.WriteByte('\n')
		b.WriteString(c.ProjectID)
		b.WriteByte('\t')
		b.WriteString(c.AgentID)
		b.WriteByte('\t')
		b.WriteString(c.Provider)
		b.WriteByte('\t')
		b.WriteString(c.Model)
		b.WriteByte('\t')
		b.WriteString(c.UseCase)
		b.WriteByte('\t')
		b.WriteString(string(c.BillingMode))
		b.WriteByte('\t')
		b.WriteString(fmt.Sprintf("%d\t%d\t%d\t%d\t%d\t%d", c.TokensIn, c.TokensOut, c.CacheReadTokens, c.CacheCreationTokens, c.Turns, c.UnreportedTurns))
	}
	return b.String()
}

func runtimeUsageCounterSortKey(c RuntimeUsageCounter) string {
	return strings.Join([]string{c.ProjectID, c.AgentID, c.Provider, c.Model, c.UseCase, string(c.BillingMode)}, "\x00")
}

func requireRuntimeResource(r *RuntimeResource) error {
	if r == nil || r.http == nil {
		return fmt.Errorf("sonzai: runtime control-plane client is not configured")
	}
	return nil
}

func (r *RuntimeResource) doJSON(ctx context.Context, method, path string, body []byte, headers map[string]string, out any) error {
	raw, err := r.http.doRaw(ctx, method, path, body, headers)
	if err != nil {
		return err
	}
	if raw.StatusCode >= http.StatusBadRequest {
		return decodeRuntimeAPIError(raw)
	}
	if out == nil || len(raw.Body) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw.Body, out); err != nil {
		return fmt.Errorf("decode runtime control-plane response: %w", err)
	}
	return nil
}

func decodeRuntimeAPIError(raw *RawResponse) error {
	parsed := struct {
		Detail string `json:"detail"`
		Title  string `json:"title"`
		Code   string `json:"code"`
		Error  struct {
			Code string `json:"code"`
		} `json:"error"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	_ = json.Unmarshal(raw.Body, &parsed)
	code := parsed.Code
	if code == "" {
		code = parsed.Error.Code
	}
	if code == "" && len(parsed.Errors) > 0 {
		code = parsed.Errors[0].Message
	}
	message := parsed.Detail
	if message == "" {
		message = parsed.Title
	}
	if message == "" {
		message = strings.TrimSpace(string(raw.Body))
	}
	return &RuntimeAPIError{StatusCode: raw.StatusCode, Code: code, Message: message, Body: append([]byte(nil), raw.Body...)}
}
