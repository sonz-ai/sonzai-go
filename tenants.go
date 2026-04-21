package sonzai

import (
	"context"
	"fmt"
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
