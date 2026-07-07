package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// TenantsResource provides tenant lookup operations.
type TenantsResource struct {
	http *httpClient
}

// Tenant represents an organization tenant.
type Tenant struct {
	TenantID     string `json:"tenant_id"`
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	ClerkOrgID   string `json:"clerk_org_id,omitempty"`
	LicenseKeyID string `json:"license_key_id,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// TenantListResponse is the response from listing tenants.
type TenantListResponse struct {
	Tenants []Tenant `json:"tenants"`
}

// List returns all tenants accessible to the authenticated user.
func (t *TenantsResource) List(ctx context.Context) (*TenantListResponse, error) {
	var result TenantListResponse
	if err := t.http.Get(ctx, "/api/v1/tenants", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single tenant by ID.
func (t *TenantsResource) Get(ctx context.Context, tenantID string) (*Tenant, error) {
	var result Tenant
	if err := t.http.Get(ctx, fmt.Sprintf("/api/v1/tenants/%s", tenantID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListOrgKnowledgeNodesOptions filters a ListOrgKnowledgeNodes request. Both
// fields are optional.
type ListOrgKnowledgeNodesOptions struct {
	// NodeType, when set, restricts results to nodes of this type.
	NodeType string
	// Limit caps the number of nodes returned.
	Limit int
}

// ListOrgKnowledgeNodes returns nodes in the tenant's organization-global
// knowledge base scope. It is the same endpoint as
// Knowledge.ListOrgNodes, exposed here on Tenants for parity with the Python
// (tenants.list_org_knowledge_nodes) and TypeScript
// (tenants.listOrgKnowledgeNodes) SDKs.
func (t *TenantsResource) ListOrgKnowledgeNodes(ctx context.Context, tenantID string, opts *ListOrgKnowledgeNodesOptions) (*OrgNodeListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.NodeType != "" {
			params["node_type"] = opts.NodeType
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
	}
	var result OrgNodeListResponse
	path := fmt.Sprintf("/api/v1/tenants/%s/knowledge/org-nodes", tenantID)
	if err := t.http.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// HostManifestOptions configures a GetHostManifest request.
type HostManifestOptions struct {
	// IfNoneMatch, when set, is sent as the If-None-Match request header for
	// a conditional fetch; a matching ETag yields NotModified=true with no
	// Manifest body, instead of re-transferring an unchanged manifest.
	IfNoneMatch string
}

// HostManifestResult is the response from GetHostManifest.
type HostManifestResult struct {
	// Manifest is the raw manifest JSON body (nil when NotModified is true).
	Manifest json.RawMessage
	// ETag is the upstream's ETag response header, if present.
	ETag string
	// CacheControl is the upstream's Cache-Control response header, if present.
	CacheControl string
	// NotModified is true on a 304 response to a conditional request
	// (opts.IfNoneMatch matched the current ETag).
	NotModified bool
}

// GetHostManifest fetches the tenant manifest resolved from host — the
// unauthenticated, host-routed GET /v1/host/manifest endpoint a managed-
// placement multi-tenant runtime uses to pull branding/module/terminology
// configuration for the tenant whose domain it's currently serving (see
// docs/design/platform-deployment-tiers.md §3). Pair this with
// sonzai.WithTenantHost(ctx, host) so the request forwards the right Host /
// X-Sonzai-Tenant-Host — this endpoint has no {id} path parameter; the
// tenant is resolved entirely from the request's Host.
//
// A 404 (unknown host, or no manifest set for the tenant yet) surfaces as
// *NotFoundError, same as every other typed method in this SDK.
func (t *TenantsResource) GetHostManifest(ctx context.Context, opts HostManifestOptions) (*HostManifestResult, error) {
	headers := map[string]string{}
	if opts.IfNoneMatch != "" {
		headers["If-None-Match"] = opts.IfNoneMatch
	}
	raw, err := t.http.doRaw(ctx, "GET", "/v1/host/manifest", nil, headers)
	if err != nil {
		return nil, err
	}

	result := &HostManifestResult{
		ETag:         raw.Header.Get("ETag"),
		CacheControl: raw.Header.Get("Cache-Control"),
	}
	switch {
	case raw.StatusCode == 304:
		result.NotModified = true
		return result, nil
	case raw.StatusCode >= 400:
		msg := string(raw.Body)
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw.Body, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return nil, newErrorForStatus(raw.StatusCode, msg, nil)
	default:
		result.Manifest = raw.Body
		return result, nil
	}
}
