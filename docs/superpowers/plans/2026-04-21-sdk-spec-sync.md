# SDK Spec Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add every endpoint in the OpenAPI spec that is currently missing from the sonzai-go SDK so the SDK has complete coverage.

**Architecture:** Each missing resource group gets its own `.go` file following the existing pattern (`type XResource struct { http *httpClient }` + methods). New resources are added as fields on `Client` and initialized in `NewClient`. Responses with no schema in the spec use `map[string]any`. Tests live in `client_test.go`.

**Tech Stack:** Go 1.21+, `net/http/httptest` for tests, no new dependencies.

---

## File Map

| File | Status | Responsibility |
|------|--------|----------------|
| `me.go` | **Create** | `Client.Me()` + `MeResponse` type |
| `tenants.go` | **Create** | `TenantsResource` (List, Get) + `Tenant` type |
| `knowledge_org.go` | **Modify** | Add `ListOrgNodes()` |
| `api_keys.go` | **Create** | `APIKeysResource` (List, Create, Revoke) + types |
| `analytics.go` | **Create** | `AnalyticsResource` (5 GET endpoints, untyped) |
| `user_personas.go` | **Create** | `UserPersonasResource` (CRUD) + types |
| `storefront.go` | **Create** | `StorefrontResource` (7 endpoints) + types |
| `org.go` | **Create** | `OrgResource` (billing, contract, ledger, etc.) + types |
| `workbench.go` | **Create** | `WorkbenchResource` (9 debugging endpoints) |
| `sessions.go` | **Modify** | Add `SessionToolDef` type alias note |
| `types.go` | **Modify** | Add `ChatSSEChunkError` type |
| `client.go` | **Modify** | Add 7 new resource fields + initialize in `NewClient` |
| `client_test.go` | **Modify** | Add tests for every new method |

---

## Task 1: ChatSSEChunkError and SessionToolDef types

**Files:**
- Modify: `types.go`

These are the two new schemas added in the latest spec sync. `SessionToolDef` maps exactly to the existing `ToolDefinition` — add a type alias. `ChatSSEChunkError` is a new error wrapper for SSE stream errors.

- [ ] **Step 1: Add types to types.go**

Open `types.go` and append at the end:

```go
// ChatSSEChunkError is an error payload that may appear inside an SSE stream chunk.
type ChatSSEChunkError struct {
	Message string `json:"message"`
}

// SessionToolDef is an alias for ToolDefinition for use in session tool configuration.
// The spec names it SessionToolDef; the SDK exposes it as ToolDefinition.
type SessionToolDef = ToolDefinition
```

- [ ] **Step 2: Build to verify**

```bash
go build ./...
```
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add types.go
git commit -m "feat: add ChatSSEChunkError and SessionToolDef types from spec"
```

---

## Task 2: Client.Me() + MeResponse

**Files:**
- Create: `me.go`
- Modify: `client_test.go`

`GET /me` returns the authenticated user's org memberships. It's a top-level method on `Client`, not a sub-resource (same pattern as `Client.ListModels()`).

- [ ] **Step 1: Write the failing test**

Add to `client_test.go`:

```go
func TestMe(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/me" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, MeResponse{
			UserID: "user-1",
			Email:  "user@example.com",
			Orgs:   []OrgMembership{{OrgID: "org-1", Role: "admin"}},
		})
	})
	defer srv.Close()

	result, err := client.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.UserID != "user-1" {
		t.Errorf("got UserID %q, want %q", result.UserID, "user-1")
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
go test ./... -run TestMe -v
```
Expected: FAIL — `MeResponse` undefined, `client.Me` undefined.

- [ ] **Step 3: Create me.go**

```go
package sonzai

import "context"

// OrgMembership represents a user's membership in an organization.
type OrgMembership struct {
	OrgID string `json:"org_id"`
	Role  string `json:"role"`
	Name  string `json:"name,omitempty"`
}

// MeResponse is the response from GET /me.
type MeResponse struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Orgs   []OrgMembership `json:"orgs"`
}

// Me returns the authenticated user's profile and org memberships.
func (c *Client) Me(ctx context.Context) (*MeResponse, error) {
	var result MeResponse
	if err := c.http.Get(ctx, "/api/v1/me", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
```

- [ ] **Step 4: Run test to confirm it passes**

```bash
go test ./... -run TestMe -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add me.go client_test.go
git commit -m "feat: add Client.Me() for GET /me"
```

---

## Task 3: TenantsResource + ListOrgNodes

**Files:**
- Create: `tenants.go`
- Modify: `knowledge_org.go` (add `ListOrgNodes`)
- Modify: `client.go` (add `Tenants` field)
- Modify: `client_test.go`

Tenants are the top-level org objects. `GET /tenants` lists them; `GET /tenants/{id}` gets one. `GET /tenants/{id}/knowledge/org-nodes` lists org-global KB nodes — it belongs on `KnowledgeResource` alongside the existing `CreateOrgNode`.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestTenantsList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/tenants" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"tenants": []Tenant{{TenantID: "t-1", Name: "Acme"}}})
	})
	defer srv.Close()
	result, err := client.Tenants.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tenants) != 1 || result.Tenants[0].TenantID != "t-1" {
		t.Error("unexpected result")
	}
}

func TestTenantsGet(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tenants/t-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, Tenant{TenantID: "t-1", Name: "Acme", IsActive: true})
	})
	defer srv.Close()
	result, err := client.Tenants.Get(context.Background(), "t-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.TenantID != "t-1" {
		t.Error("unexpected result")
	}
}

func TestKnowledgeListOrgNodes(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tenants/t-1/knowledge/org-nodes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"nodes": []KBNode{{NodeID: "n-1"}}, "total": 1})
	})
	defer srv.Close()
	result, err := client.Knowledge.ListOrgNodes(context.Background(), "t-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Errorf("got total %d, want 1", result.Total)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestTenants|TestKnowledgeListOrgNodes" -v
```
Expected: FAIL — `client.Tenants` undefined, types undefined.

- [ ] **Step 3: Create tenants.go**

```go
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
```

- [ ] **Step 4: Add ListOrgNodes to knowledge_org.go**

Append to `knowledge_org.go`:

```go
// OrgNodeListResponse is the response from listing org-global KB nodes.
type OrgNodeListResponse struct {
	Nodes []*KBNode `json:"nodes"`
	Total int       `json:"total"`
}

// ListOrgNodes returns all nodes in the organization-global knowledge base scope.
func (k *KnowledgeResource) ListOrgNodes(ctx context.Context, tenantID string, limit int) (*OrgNodeListResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result OrgNodeListResponse
	path := fmt.Sprintf("/api/v1/tenants/%s/knowledge/org-nodes", tenantID)
	if err := k.http.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
```

- [ ] **Step 5: Add Tenants to client.go**

In `client.go`, add `Tenants *TenantsResource` to the `Client` struct after `ProjectNotifications`:

```go
// Tenants provides tenant lookup operations.
Tenants *TenantsResource
```

In `NewClient`, add to the return struct:

```go
Tenants: &TenantsResource{http: hc},
```

Also add `Tenants` to `TestNewClientCreatesResources` in `client_test.go`:

```go
if c.Tenants == nil {
    t.Fatal("Tenants is nil")
}
```

- [ ] **Step 6: Build and run tests**

```bash
go build ./... && go test ./... -run "TestTenants|TestKnowledgeListOrgNodes|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add tenants.go knowledge_org.go client.go client_test.go
git commit -m "feat: add TenantsResource and KnowledgeResource.ListOrgNodes"
```

---

## Task 4: APIKeysResource

**Files:**
- Create: `api_keys.go`
- Modify: `client.go`
- Modify: `client_test.go`

Project API keys are created/listed/revoked under `/projects/{projectId}/keys`. A new resource `APIKeysResource` handles this.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestAPIKeysCreate(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/projects/proj-1/keys" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, APIKey{
			KeyID:     "key-1",
			Key:       "sk-live-abc123",
			ProjectID: "proj-1",
			IsActive:  true,
		})
	})
	defer srv.Close()
	result, err := client.APIKeys.Create(context.Background(), "proj-1", CreateAPIKeyOptions{Name: "my-key"})
	if err != nil {
		t.Fatal(err)
	}
	if result.KeyID != "key-1" {
		t.Errorf("got KeyID %q, want key-1", result.KeyID)
	}
}

func TestAPIKeysList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/proj-1/keys" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"keys": []APIKey{{KeyID: "key-1"}}})
	})
	defer srv.Close()
	result, err := client.APIKeys.List(context.Background(), "proj-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Keys) != 1 {
		t.Errorf("got %d keys, want 1", len(result.Keys))
	}
}

func TestAPIKeysRevoke(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/projects/proj-1/keys/key-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"success": true})
	})
	defer srv.Close()
	err := client.APIKeys.Revoke(context.Background(), "proj-1", "key-1")
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestAPIKeys" -v
```
Expected: FAIL — `client.APIKeys` undefined.

- [ ] **Step 3: Create api_keys.go**

```go
package sonzai

import (
	"context"
	"fmt"
)

// APIKeysResource provides project API key management operations.
type APIKeysResource struct {
	http *httpClient
}

// APIKey represents a project API key.
type APIKey struct {
	KeyID     string   `json:"key_id"`
	Key       string   `json:"key,omitempty"`
	KeyPrefix string   `json:"key_prefix,omitempty"`
	Name      string   `json:"name,omitempty"`
	ProjectID string   `json:"project_id"`
	TenantID  string   `json:"tenant_id,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
	IsActive  bool     `json:"is_active"`
	CreatedBy string   `json:"created_by,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

// APIKeyListResponse is the response from listing API keys.
type APIKeyListResponse struct {
	Keys []APIKey `json:"keys"`
}

// CreateAPIKeyOptions configures an API key creation request.
type CreateAPIKeyOptions struct {
	Name        string   `json:"name,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	ExpiresDays int      `json:"expires_days,omitempty"`
}

// List returns all API keys for a project.
func (a *APIKeysResource) List(ctx context.Context, projectID string) (*APIKeyListResponse, error) {
	var result APIKeyListResponse
	if err := a.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/keys", projectID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new API key for a project.
// The full key value is only returned once in the response — store it securely.
func (a *APIKeysResource) Create(ctx context.Context, projectID string, opts CreateAPIKeyOptions) (*APIKey, error) {
	var result APIKey
	if err := a.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/keys", projectID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Revoke permanently revokes an API key.
func (a *APIKeysResource) Revoke(ctx context.Context, projectID, keyID string) error {
	return a.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/keys/%s", projectID, keyID), nil)
}
```

- [ ] **Step 4: Add APIKeys to client.go**

Add to `Client` struct:

```go
// APIKeys provides project API key management.
APIKeys *APIKeysResource
```

Add to `NewClient` return:

```go
APIKeys: &APIKeysResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.APIKeys == nil {
    t.Fatal("APIKeys is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestAPIKeys|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add api_keys.go client.go client_test.go
git commit -m "feat: add APIKeysResource for project API key management"
```

---

## Task 5: AnalyticsResource

**Files:**
- Create: `analytics.go`
- Modify: `client.go`
- Modify: `client_test.go`

Five `GET /analytics/*` endpoints. The spec defines no response schemas for these, so responses use `map[string]any`. Accept optional `days` and `month`/`start`/`end` query params.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestAnalyticsOverview(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/overview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"total_agents": 5})
	})
	defer srv.Close()
	result, err := client.Analytics.Overview(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAnalyticsCostBreakdown(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/cost/breakdown" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("days") != "7" {
			t.Errorf("expected days=7, got %q", r.URL.Query().Get("days"))
		}
		jsonResponse(w, 200, map[string]any{"total_usd": 1.23})
	})
	defer srv.Close()
	result, err := client.Analytics.CostBreakdown(context.Background(), &AnalyticsOptions{Days: 7})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestAnalytics" -v
```
Expected: FAIL — `client.Analytics` undefined.

- [ ] **Step 3: Create analytics.go**

```go
package sonzai

import (
	"context"
	"strconv"
)

// AnalyticsResource provides platform analytics endpoints.
type AnalyticsResource struct {
	http *httpClient
}

// AnalyticsOptions configures analytics query time range.
// Use Days for a lookback window (1-365). Use Month ("YYYY-MM") or
// Start+End ("YYYY-MM-DD") for a specific range.
type AnalyticsOptions struct {
	Days      int
	Month     string
	Start     string
	End       string
	ProjectID string
}

func (o *AnalyticsOptions) toParams() map[string]string {
	if o == nil {
		return nil
	}
	p := map[string]string{}
	if o.Days > 0 {
		p["days"] = strconv.Itoa(o.Days)
	}
	if o.Month != "" {
		p["month"] = o.Month
	}
	if o.Start != "" {
		p["start"] = o.Start
	}
	if o.End != "" {
		p["end"] = o.End
	}
	if o.ProjectID != "" {
		p["project_id"] = o.ProjectID
	}
	return p
}

// Overview returns high-level platform analytics.
func (a *AnalyticsResource) Overview(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/overview", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Usage returns token and request usage metrics.
func (a *AnalyticsResource) Usage(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/usage", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Cost returns cost summary metrics.
func (a *AnalyticsResource) Cost(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/cost", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CostBreakdown returns cost broken down by operation, model, and agent.
func (a *AnalyticsResource) CostBreakdown(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/cost/breakdown", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Realtime returns real-time platform metrics.
func (a *AnalyticsResource) Realtime(ctx context.Context, opts *AnalyticsOptions) (map[string]any, error) {
	var result map[string]any
	if err := a.http.Get(ctx, "/api/v1/analytics/realtime", opts.toParams(), &result); err != nil {
		return nil, err
	}
	return result, nil
}
```

- [ ] **Step 4: Add Analytics to client.go**

Add to `Client` struct:

```go
// Analytics provides platform analytics and cost reporting.
Analytics *AnalyticsResource
```

Add to `NewClient` return:

```go
Analytics: &AnalyticsResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.Analytics == nil {
    t.Fatal("Analytics is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestAnalytics|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add analytics.go client.go client_test.go
git commit -m "feat: add AnalyticsResource for GET /analytics/* endpoints"
```

---

## Task 6: UserPersonasResource

**Files:**
- Create: `user_personas.go`
- Modify: `client.go`
- Modify: `client_test.go`

User personas are tenant-scoped prompt templates for customizing how an agent addresses users. Full CRUD on `/user-personas/*`.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestUserPersonasCreate(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/user-personas" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, UserPersona{
			PersonaID:   "p-1",
			Name:        "Friendly",
			Description: "A friendly persona",
			Style:       "casual",
			IsDefault:   false,
		})
	})
	defer srv.Close()
	result, err := client.UserPersonas.Create(context.Background(), CreateUserPersonaOptions{
		Name:        "Friendly",
		Description: "A friendly persona",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PersonaID != "p-1" {
		t.Errorf("got PersonaID %q, want p-1", result.PersonaID)
	}
}

func TestUserPersonasList(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user-personas" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"personas": []UserPersona{{PersonaID: "p-1"}}})
	})
	defer srv.Close()
	result, err := client.UserPersonas.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Personas) != 1 {
		t.Errorf("got %d personas, want 1", len(result.Personas))
	}
}

func TestUserPersonasDelete(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/user-personas/p-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"success": true})
	})
	defer srv.Close()
	if err := client.UserPersonas.Delete(context.Background(), "p-1"); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestUserPersonas" -v
```
Expected: FAIL.

- [ ] **Step 3: Create user_personas.go**

```go
package sonzai

import (
	"context"
	"fmt"
)

// UserPersonasResource provides user persona CRUD operations.
type UserPersonasResource struct {
	http *httpClient
}

// UserPersona represents a user persona template.
type UserPersona struct {
	PersonaID   string `json:"persona_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Style       string `json:"style,omitempty"`
	IsDefault   bool   `json:"is_default"`
	TenantID    string `json:"tenant_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// UserPersonaListResponse is the response from listing user personas.
type UserPersonaListResponse struct {
	Personas []UserPersona `json:"personas"`
}

// CreateUserPersonaOptions configures a user persona creation request.
type CreateUserPersonaOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Style       string `json:"style,omitempty"`
}

// UpdateUserPersonaOptions configures a user persona update request.
type UpdateUserPersonaOptions struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Style       string `json:"style,omitempty"`
}

// List returns all user personas for the tenant.
func (u *UserPersonasResource) List(ctx context.Context) (*UserPersonaListResponse, error) {
	var result UserPersonaListResponse
	if err := u.http.Get(ctx, "/api/v1/user-personas", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a single user persona by ID.
func (u *UserPersonasResource) Get(ctx context.Context, personaID string) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Get(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new user persona.
func (u *UserPersonasResource) Create(ctx context.Context, opts CreateUserPersonaOptions) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Post(ctx, "/api/v1/user-personas", opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update replaces a user persona's fields.
func (u *UserPersonasResource) Update(ctx context.Context, personaID string, opts UpdateUserPersonaOptions) (*UserPersona, error) {
	var result UserPersona
	if err := u.http.Put(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete permanently deletes a user persona.
func (u *UserPersonasResource) Delete(ctx context.Context, personaID string) error {
	return u.http.Delete(ctx, fmt.Sprintf("/api/v1/user-personas/%s", personaID), nil)
}
```

- [ ] **Step 4: Add UserPersonas to client.go**

Add to `Client` struct:

```go
// UserPersonas provides user persona CRUD operations.
UserPersonas *UserPersonasResource
```

Add to `NewClient` return:

```go
UserPersonas: &UserPersonasResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.UserPersonas == nil {
    t.Fatal("UserPersonas is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestUserPersonas|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add user_personas.go client.go client_test.go
git commit -m "feat: add UserPersonasResource for CRUD on /user-personas"
```

---

## Task 7: StorefrontResource

**Files:**
- Create: `storefront.go`
- Modify: `client.go`
- Modify: `client_test.go`

The storefront is a public agent marketplace page for the tenant. Responses for `GET /storefront` and `GET /storefront/agents` have no schema — use `map[string]any`.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestStorefrontGet(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/storefront" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"slug": "my-store"})
	})
	defer srv.Close()
	result, err := client.Storefront.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["slug"] != "my-store" {
		t.Error("unexpected result")
	}
}

func TestStorefrontUpsertAgent(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/storefront/agents/agent-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"agent_id": "agent-1"})
	})
	defer srv.Close()
	err := client.Storefront.UpsertAgent(context.Background(), "agent-1", StorefrontAgentOptions{
		DisplayName: "My Agent",
	})
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestStorefront" -v
```
Expected: FAIL.

- [ ] **Step 3: Create storefront.go**

```go
package sonzai

import (
	"context"
	"fmt"
)

// StorefrontResource provides storefront (agent marketplace) management operations.
type StorefrontResource struct {
	http *httpClient
}

// StorefrontUpdateOptions configures a storefront update request.
type StorefrontUpdateOptions struct {
	DisplayName       string `json:"display_name,omitempty"`
	Description       string `json:"description,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Theme             string `json:"theme,omitempty"`
	HeroImageURL      string `json:"hero_image_url,omitempty"`
	AccessType        string `json:"access_type,omitempty"`
	InviteCode        string `json:"invite_code,omitempty"`
	ContactEmail      string `json:"contact_email,omitempty"`
	MaxVisitsPerUser  int    `json:"max_visits_per_user,omitempty"`
}

// StorefrontAgentOptions configures an agent listing on the storefront.
type StorefrontAgentOptions struct {
	DisplayName      string `json:"display_name,omitempty"`
	Description      string `json:"description,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	Slug             string `json:"slug,omitempty"`
	Featured         bool   `json:"featured,omitempty"`
	MaxTurnsPerVisit int    `json:"max_turns_per_visit,omitempty"`
}

// Get returns the tenant's storefront configuration.
func (s *StorefrontResource) Get(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := s.http.Get(ctx, "/api/v1/storefront", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Update updates the tenant's storefront configuration.
func (s *StorefrontResource) Update(ctx context.Context, opts StorefrontUpdateOptions) error {
	return s.http.Put(ctx, "/api/v1/storefront", opts, nil)
}

// ListAgents returns agents listed on the storefront.
func (s *StorefrontResource) ListAgents(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := s.http.Get(ctx, "/api/v1/storefront/agents", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpsertAgent adds or updates an agent listing on the storefront.
func (s *StorefrontResource) UpsertAgent(ctx context.Context, agentID string, opts StorefrontAgentOptions) error {
	return s.http.Put(ctx, fmt.Sprintf("/api/v1/storefront/agents/%s", agentID), opts, nil)
}

// RemoveAgent removes an agent from the storefront.
func (s *StorefrontResource) RemoveAgent(ctx context.Context, agentID string) error {
	return s.http.Delete(ctx, fmt.Sprintf("/api/v1/storefront/agents/%s", agentID), nil)
}

// Publish makes the storefront publicly visible.
func (s *StorefrontResource) Publish(ctx context.Context) error {
	return s.http.Post(ctx, "/api/v1/storefront/publish", nil, nil)
}

// Unpublish hides the storefront from public access.
func (s *StorefrontResource) Unpublish(ctx context.Context) error {
	return s.http.Post(ctx, "/api/v1/storefront/unpublish", nil, nil)
}
```

- [ ] **Step 4: Add Storefront to client.go**

Add to `Client` struct:

```go
// Storefront provides agent marketplace (storefront) management.
Storefront *StorefrontResource
```

Add to `NewClient` return:

```go
Storefront: &StorefrontResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.Storefront == nil {
    t.Fatal("Storefront is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestStorefront|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add storefront.go client.go client_test.go
git commit -m "feat: add StorefrontResource for agent marketplace management"
```

---

## Task 8: OrgResource

**Files:**
- Create: `org.go`
- Modify: `client.go`
- Modify: `client_test.go`

`/org/*` endpoints cover billing portal, subscription management, usage reporting, ledger, model pricing, and voucher redemption. Most responses have no schema — use `map[string]any`.

- [ ] **Step 1: Write failing tests**

Add to `client_test.go`:

```go
func TestOrgGetBilling(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/org/billing" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"status": "active"})
	})
	defer srv.Close()
	result, err := client.Org.GetBilling(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "active" {
		t.Error("unexpected result")
	}
}

func TestOrgRedeemVoucher(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/org/vouchers/redeem" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"redeemed": true})
	})
	defer srv.Close()
	result, err := client.Org.RedeemVoucher(context.Background(), "VOUCHER123")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected result")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./... -run "TestOrg" -v
```
Expected: FAIL.

- [ ] **Step 3: Create org.go**

```go
package sonzai

import "context"

// OrgResource provides organization-level billing, contract, and usage operations.
type OrgResource struct {
	http *httpClient
}

// OrgBillingCheckoutOptions configures a billing checkout session.
type OrgBillingCheckoutOptions struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

// OrgSubscribeOptions configures a contract subscription.
type OrgSubscribeOptions struct {
	ContractID string `json:"contractId"`
}

// GetBilling returns the org's current billing status.
func (o *OrgResource) GetBilling(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/billing", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateBillingPortal creates a billing management portal session. Returns the portal URL.
func (o *OrgResource) CreateBillingPortal(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/billing/portal", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateBillingCheckout creates a checkout session for purchasing credits.
func (o *OrgResource) CreateBillingCheckout(ctx context.Context, opts OrgBillingCheckoutOptions) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/billing/checkout", opts, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContract returns the org's current contract details.
func (o *OrgResource) GetContract(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/contract", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Subscribe subscribes the org to a contract plan.
func (o *OrgResource) Subscribe(ctx context.Context, opts OrgSubscribeOptions) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/contract/subscribe", opts, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetLedger returns the org's credit ledger.
func (o *OrgResource) GetLedger(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/ledger", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetModelPricing returns current model pricing for the org.
func (o *OrgResource) GetModelPricing(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/model-pricing", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetServiceUsage returns current service usage metrics.
func (o *OrgResource) GetServiceUsage(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/service-usage", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUsageSummary returns an aggregated usage summary.
func (o *OrgResource) GetUsageSummary(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/usage-summary", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListServiceAgreements returns the org's service agreements.
func (o *OrgResource) ListServiceAgreements(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/service-agreements", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListActiveCharacters returns currently active characters for the org.
func (o *OrgResource) ListActiveCharacters(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/characters", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetContextEngineEvents returns context engine events for the org.
func (o *OrgResource) GetContextEngineEvents(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Get(ctx, "/api/v1/org/events", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// RedeemVoucher redeems a voucher code for credits.
func (o *OrgResource) RedeemVoucher(ctx context.Context, code string) (map[string]any, error) {
	var result map[string]any
	if err := o.http.Post(ctx, "/api/v1/org/vouchers/redeem", map[string]string{"code": code}, &result); err != nil {
		return nil, err
	}
	return result, nil
}
```

- [ ] **Step 4: Add Org to client.go**

Add to `Client` struct:

```go
// Org provides organization-level billing, usage, and contract operations.
Org *OrgResource
```

Add to `NewClient` return:

```go
Org: &OrgResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.Org == nil {
    t.Fatal("Org is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestOrg|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add org.go client.go client_test.go
git commit -m "feat: add OrgResource for billing, contract, and usage operations"
```

---

## Task 9: WorkbenchResource

**Files:**
- Create: `workbench.go`
- Modify: `client.go`
- Modify: `client_test.go`

The workbench endpoints are internal simulation/debugging tools under `POST /workbench/*`. All request and response bodies are untyped in the spec — use `map[string]any` for both.

- [ ] **Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestWorkbenchChat(t *testing.T) {
	srv, client := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/workbench/chat" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		jsonResponse(w, 200, map[string]any{"response": "hello"})
	})
	defer srv.Close()
	result, err := client.Workbench.Chat(context.Background(), map[string]any{"message": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if result["response"] != "hello" {
		t.Error("unexpected result")
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
go test ./... -run "TestWorkbench" -v
```
Expected: FAIL.

- [ ] **Step 3: Create workbench.go**

```go
package sonzai

import "context"

// WorkbenchResource provides internal simulation and debugging operations.
// These endpoints are intended for development and testing workflows.
type WorkbenchResource struct {
	http *httpClient
}

// Chat runs a workbench chat request.
func (w *WorkbenchResource) Chat(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/chat", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Prepare prepares a workbench session.
func (w *WorkbenchResource) Prepare(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/prepare", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetState returns the current workbench state.
func (w *WorkbenchResource) GetState(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/state", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// AdvanceTime advances the workbench simulation clock.
func (w *WorkbenchResource) AdvanceTime(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/advance-time", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ResetAgent resets an agent within a workbench session.
func (w *WorkbenchResource) ResetAgent(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/reset-agent", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SessionEnd ends the workbench session.
func (w *WorkbenchResource) SessionEnd(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/session-end", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SimulateUser sends a simulated user message in a workbench session.
func (w *WorkbenchResource) SimulateUser(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/simulate-user", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateBio runs bio generation in the workbench.
func (w *WorkbenchResource) GenerateBio(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-bio", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateCharacter runs character generation in the workbench.
func (w *WorkbenchResource) GenerateCharacter(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-character", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateSeedMemories runs seed memory generation in the workbench.
func (w *WorkbenchResource) GenerateSeedMemories(ctx context.Context, body map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := w.http.Post(ctx, "/api/v1/workbench/generate-seed-memories", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
```

- [ ] **Step 4: Add Workbench to client.go**

Add to `Client` struct:

```go
// Workbench provides internal simulation and debugging operations.
Workbench *WorkbenchResource
```

Add to `NewClient` return:

```go
Workbench: &WorkbenchResource{http: hc},
```

Add to `TestNewClientCreatesResources`:

```go
if c.Workbench == nil {
    t.Fatal("Workbench is nil")
}
```

- [ ] **Step 5: Build and run tests**

```bash
go build ./... && go test ./... -run "TestWorkbench|TestNewClientCreatesResources" -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add workbench.go client.go client_test.go
git commit -m "feat: add WorkbenchResource for internal simulation endpoints"
```

---

## Task 10: Final verification and release commit

- [ ] **Step 1: Run full test suite**

```bash
go test ./... -timeout 2m -v 2>&1 | tail -20
```
Expected: all tests PASS, no failures.

- [ ] **Step 2: Run vet and format**

```bash
go vet ./... && gofmt -w .
```
Expected: no output.

- [ ] **Step 3: Verify spec coverage**

```bash
python3 - <<'EOF'
import json, subprocess, re
with open('openapi.json') as f:
    spec = json.load(f)
result = subprocess.run(['grep', '-rn', 'api/v1/', '--include=*.go', '.'], capture_output=True, text=True)
sdk_paths = set()
for line in result.stdout.splitlines():
    if '_test.go' in line:
        continue
    for m in re.findall(r'"(/api/v1/[^"]*)"', line):
        sdk_paths.add(re.sub(r'%s', '{id}', m))

missing = []
for path, methods in spec['paths'].items():
    norm = re.sub(r'\{[^}]+\}', '{id}', path)
    norm_clean = re.sub(r'\{id\}', '', norm).replace('//', '/')
    for method, op in methods.items():
        if method not in ('get','post','put','patch','delete'):
            continue
        found = any(norm_clean.rstrip('/') in sp for sp in sdk_paths)
        if not found:
            missing.append(f"{method.upper()} {path}")
print(f"Still missing: {len(missing)}")
for m in missing[:20]:
    print(" ", m)
EOF
```
Expected: 0 or only endpoints confirmed as intentionally excluded.

- [ ] **Step 4: Commit version bump**

```bash
git add .
git commit -m "release: v1.2.5"
```

- [ ] **Step 5: Push**

```bash
git push origin main
```

---

## Self-Review Checklist

- [x] **Spec coverage:** Tasks 1-9 cover all 8 missing resource groups identified in analysis.
- [x] **No placeholders:** All steps contain actual code.
- [x] **Type consistency:** `UserPersona` used in both list response and CRUD methods. `APIKey` used in both list and create responses. `ToolDefinition` aliased as `SessionToolDef`.
- [x] **Pattern consistency:** All new resources follow existing `type XResource struct { http *httpClient }` pattern.
- [x] **Untyped responses:** Analytics, Org, Storefront, Workbench use `map[string]any` where spec has no schema — consistent with how untyped responses are handled elsewhere.
- [x] **Build verification:** Every task ends with `go build ./...` before commit.
