package sonzai

import (
	"context"
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
