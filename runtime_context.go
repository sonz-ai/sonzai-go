package sonzai

import (
	"context"
	"net/http"
)

// contextKey is an unexported type for this package's context keys, per the
// standard library convention (avoids collisions with other packages' keys).
type contextKey int

const (
	tenantHostContextKey contextKey = iota
	operatorIDContextKey
)

// WithTenantHost returns a context that causes every SDK request made with
// it to forward host as both the outgoing HTTP Host header and the
// X-Sonzai-Tenant-Host header (platform-api's managed-placement S2S auth
// path resolves the tenant from the latter). This is for multi-tenant
// BFF/runtime consumers that serve many tenant hosts from a single process
// and resolve the tenant per-request from their own inbound Host header — a
// single *Client can be shared across every tenant since the host travels on
// ctx, not on the Client.
func WithTenantHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, tenantHostContextKey, host)
}

// TenantHostFromContext returns the host set by WithTenantHost, if any.
func TenantHostFromContext(ctx context.Context) (string, bool) {
	host, ok := ctx.Value(tenantHostContextKey).(string)
	return host, ok && host != ""
}

// WithOperatorID returns a context that causes every SDK request made with
// it to carry the X-Sonzai-Operator-ID header. platform-api's managed S2S
// auth path (Authorization: Bearer <service key> + X-Sonzai-Tenant-Host)
// reads this header to attribute a conversation takeover/release/message-send
// to a validated end-operator session, since the S2S bearer itself carries no
// per-operator identity.
func WithOperatorID(ctx context.Context, operatorID string) context.Context {
	return context.WithValue(ctx, operatorIDContextKey, operatorID)
}

// OperatorIDFromContext returns the operator ID set by WithOperatorID, if any.
func OperatorIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(operatorIDContextKey).(string)
	return id, ok && id != ""
}

// applyRuntimeContextHeaders stamps req with the WithTenantHost /
// WithOperatorID overrides carried on ctx, if any. Every request-building
// call site in this package (typed calls, streaming, and the Raw passthrough
// in raw.go) applies this so a single *Client can serve many tenant hosts
// concurrently, keyed off ctx rather than Client state.
func applyRuntimeContextHeaders(ctx context.Context, req *http.Request) {
	if host, ok := TenantHostFromContext(ctx); ok {
		req.Host = host
		req.Header.Set("X-Sonzai-Tenant-Host", host)
	}
	if operatorID, ok := OperatorIDFromContext(ctx); ok {
		req.Header.Set("X-Sonzai-Operator-ID", operatorID)
	}
}
