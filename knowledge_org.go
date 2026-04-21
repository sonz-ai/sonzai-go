package sonzai

import (
	"context"
	"fmt"
	"strconv"
)

// =============================================================================
// ORG-GLOBAL KNOWLEDGE BASE (docs/ORGANIZATION_GLOBAL_KB.md)
//
// Wraps the Phase 5 admin endpoints on services/platform/api. Lets tenant
// admins populate the organization-global KB scope — the scope every
// project under the tenant reads when its agents opt into
// KBScopeMode = cascade / union / org_only.
//
// These helpers live on the existing KnowledgeResource rather than a new
// resource type so callers can `client.Knowledge.CreateOrgNode(...)`
// without another field on Client.
// =============================================================================

// KBScopeMode controls how an agent's knowledge_search reads across the
// project and organization-global scopes. See
// sonzai-ai-monolith-ts/docs/ORGANIZATION_GLOBAL_KB.md §6.2.
type KBScopeMode string

const (
	// KBScopeProjectOnly reads only the caller's project scope (default).
	KBScopeProjectOnly KBScopeMode = "project_only"
	// KBScopeOrgOnly reads only the organization-global scope.
	KBScopeOrgOnly KBScopeMode = "org_only"
	// KBScopeCascade reads both scopes; project wins on ID collisions.
	KBScopeCascade KBScopeMode = "cascade"
	// KBScopeUnion reads both scopes with equal weight.
	KBScopeUnion KBScopeMode = "union"
)

// CreateOrgNodeOptions is the request body for CreateOrgNode.
type CreateOrgNodeOptions struct {
	NodeType   string         `json:"node_type"`
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties,omitempty"`
	// Confidence defaults to 1.0 server-side for hand-authored org knowledge.
	Confidence float64 `json:"confidence,omitempty"`
}

// KBNodeWithScope is a KBNode with scope provenance attached, returned by
// the promote endpoint and by cascade reads.
type KBNodeWithScope struct {
	*KBNode
	ScopeType string  `json:"scope_type"` // "project" | "organization"
	Relevance float64 `json:"relevance"`
}

// CreateOrgNode creates a knowledge-base node directly in the
// organization-global scope (scope_id == ""). Every project under the
// tenant can read it via cascade / union / org_only scope modes.
//
// Idempotency is the caller's responsibility — look up by label before
// calling this if duplicates are a concern.
func (k *KnowledgeResource) CreateOrgNode(ctx context.Context, tenantID string, opts CreateOrgNodeOptions) (*KBNode, error) {
	var result KBNode
	path := fmt.Sprintf("/api/v1/tenants/%s/knowledge/org-nodes", tenantID)
	if err := k.http.Post(ctx, path, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PromoteNodeToOrg copies a project-scoped node into the organization-global
// scope. The project copy is preserved — promotion is additive. If an org
// node with the same (NodeType, NormLabel) already exists, the server
// returns it instead of writing a duplicate.
func (k *KnowledgeResource) PromoteNodeToOrg(ctx context.Context, projectID, nodeID, tenantID string) (*KBNodeWithScope, error) {
	var result KBNodeWithScope
	path := fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes/%s/promote-to-org", projectID, nodeID)
	body := struct {
		TenantID string `json:"tenant_id"`
	}{TenantID: tenantID}
	if err := k.http.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

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
