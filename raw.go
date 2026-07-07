package sonzai

import (
	"context"
	"net/http"
)

// RawResponse is a raw, unretried, un-decoded HTTP response: status code,
// body bytes, and response headers. It exists for BFF/proxy consumers (e.g.
// a tenant-facing runtime fronting platform-api) that must relay an
// upstream's exact status/body/headers to their own caller — conditional
// requests (ETag/304), a documented "not ready yet" status that must stay
// distinct from a genuine outage (e.g. 503 vs 502), or an upstream error body
// whose exact shape the caller inspects itself. Prefer the typed resource
// methods (client.Agents, client.Conversations, client.ML, ...) for normal
// use; reach for Raw/RawStream only when byte-for-byte passthrough is the
// actual contract.
type RawResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// Raw issues a single, unretried HTTP request to path and returns the raw
// response. Unlike the typed Get/Post/... helpers, a non-2xx status is NOT
// translated into a Go error — err is only non-nil for a transport-level
// failure (DNS, connection, timeout). The request carries the same
// Authorization bearer and User-Agent as every other SDK call, plus the
// ctx-scoped WithTenantHost / WithOperatorID headers (runtime_context.go).
// body is the already-marshaled request body, or nil.
func (c *Client) Raw(ctx context.Context, method, path string, body []byte, extraHeaders map[string]string) (*RawResponse, error) {
	return c.http.doRaw(ctx, method, path, body, extraHeaders)
}

// RawStream issues a single, unretried, streaming HTTP request and returns
// the live *http.Response for the caller to read incrementally (e.g. relay
// an SSE stream byte-for-byte to its own caller) — the caller owns closing
// resp.Body. Like StreamSSE, it runs without the Client's overall request
// timeout (see longRunningClient) so a long-lived stream is capped only by
// ctx, not by WithTimeout.
func (c *Client) RawStream(ctx context.Context, method, path string, body []byte, extraHeaders map[string]string) (*http.Response, error) {
	return c.http.doRawStream(ctx, method, path, body, extraHeaders)
}
