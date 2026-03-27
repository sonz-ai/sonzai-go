package sonzai

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PrimingResource provides user priming operations for an agent.
type PrimingResource struct {
	http *httpClient
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// UserPrimingMetadata represents stored metadata for a primed user.
type UserPrimingMetadata struct {
	AgentID      string            `json:"agent_id"`
	UserID       string            `json:"user_id"`
	DisplayName  string            `json:"display_name,omitempty"`
	Company      string            `json:"company,omitempty"`
	Title        string            `json:"title,omitempty"`
	Email        string            `json:"email,omitempty"`
	Phone        string            `json:"phone,omitempty"`
	SourceType   string            `json:"source_type,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
	PrimedAt     string            `json:"primed_at,omitempty"`
}

// PrimeUserMetadata represents metadata to include when priming a user.
type PrimeUserMetadata struct {
	Company string            `json:"company,omitempty"`
	Title   string            `json:"title,omitempty"`
	Email   string            `json:"email,omitempty"`
	Phone   string            `json:"phone,omitempty"`
	Custom  map[string]string `json:"custom,omitempty"`
}

// PrimeContentBlock represents a content block for priming.
type PrimeContentBlock struct {
	Type string `json:"type"` // "text", "chat_transcript", etc.
	Body string `json:"body"`
}

// ImportJob represents an import/priming job.
type ImportJob struct {
	JobID          string `json:"job_id"`
	TenantID       string `json:"tenant_id,omitempty"`
	AgentID        string `json:"agent_id,omitempty"`
	JobType        string `json:"job_type,omitempty"` // "prime" or "import"
	UserID         string `json:"user_id,omitempty"`
	Source         string `json:"source,omitempty"`
	Status         string `json:"status"`
	TotalUsers     int    `json:"total_users,omitempty"`
	ProcessedUsers int    `json:"processed_users,omitempty"`
	FactsCreated   int    `json:"facts_created,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// ImportJobListResponse is the response from listing import jobs.
type ImportJobListResponse struct {
	Jobs  []*ImportJob `json:"jobs"`
	Count int          `json:"count"`
}

// ---------------------------------------------------------------------------
// Request Options
// ---------------------------------------------------------------------------

// StructuredColumnMapping defines how a CSV column maps to fact metadata.
type StructuredColumnMapping struct {
	Property string `json:"property"`
	IsLabel  bool   `json:"is_label,omitempty"`
	Type     string `json:"type,omitempty"` // "number", "boolean", or default string
}

// StructuredImportSpec defines a CSV-to-facts structured import.
type StructuredImportSpec struct {
	EntityType    string                             `json:"entity_type"`
	ContentCSV    string                             `json:"content_csv"`
	ColumnMapping map[string]StructuredColumnMapping `json:"column_mapping"`
	ProjectID     string                             `json:"project_id,omitempty"`
}

// PrimeUserOptions configures a user priming request.
type PrimeUserOptions struct {
	DisplayName      string               `json:"display_name,omitempty"`
	Metadata         *PrimeUserMetadata   `json:"metadata,omitempty"`
	Content          []PrimeContentBlock  `json:"content,omitempty"`
	Source           string               `json:"source,omitempty"`
	StructuredImport *StructuredImportSpec `json:"structured_import,omitempty"`
}

// PrimeUserResponse is the response from priming a user.
type PrimeUserResponse struct {
	JobID        string `json:"job_id"`
	Status       string `json:"status"`
	FactsCreated int    `json:"facts_created"`
	RowsParsed   int    `json:"rows_parsed,omitempty"`
	KBResolved   int    `json:"kb_resolved,omitempty"`
	Unresolved   int    `json:"unresolved,omitempty"`
}

// AddContentOptions configures an add-content request.
type AddContentOptions struct {
	Content []PrimeContentBlock `json:"content"`
	Source  string              `json:"source,omitempty"`
}

// AddContentResponse is the response from adding content.
type AddContentResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// UpdateMetadataOptions configures a metadata update request.
type UpdateMetadataOptions struct {
	DisplayName *string           `json:"display_name,omitempty"`
	Company     *string           `json:"company,omitempty"`
	Title       *string           `json:"title,omitempty"`
	Email       *string           `json:"email,omitempty"`
	Phone       *string           `json:"phone,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// UpdateMetadataResponse is the response from updating metadata.
type UpdateMetadataResponse struct {
	Metadata     *UserPrimingMetadata `json:"metadata"`
	FactsCreated int                  `json:"facts_created"`
}

// BatchImportUser represents a single user in a batch import.
type BatchImportUser struct {
	UserID      string              `json:"user_id"`
	DisplayName string              `json:"display_name,omitempty"`
	Metadata    *PrimeUserMetadata  `json:"metadata,omitempty"`
	Content     []PrimeContentBlock `json:"content,omitempty"`
}

// BatchImportOptions configures a batch import request.
type BatchImportOptions struct {
	Users  []BatchImportUser `json:"users"`
	Source string            `json:"source,omitempty"`
}

// BatchImportResponse is the response from a batch import.
type BatchImportResponse struct {
	JobID        string `json:"job_id"`
	Status       string `json:"status"`
	TotalUsers   int    `json:"total_users"`
	FactsCreated int    `json:"facts_created"`
}

// ---------------------------------------------------------------------------
// Methods
// ---------------------------------------------------------------------------

// PrimeUser primes a user with metadata and content for an agent.
// Returns immediately with a job ID; LLM extraction runs asynchronously.
func (p *PrimingResource) PrimeUser(ctx context.Context, agentID, userID string, opts PrimeUserOptions) (*PrimeUserResponse, error) {
	var result PrimeUserResponse
	err := p.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/prime", agentID, url.PathEscape(userID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPrimeStatus returns the status of a priming job.
func (p *PrimingResource) GetPrimeStatus(ctx context.Context, agentID, userID, jobID string) (*ImportJob, error) {
	var result ImportJob
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/prime/%s", agentID, url.PathEscape(userID), jobID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddContent adds content blocks for async LLM extraction.
func (p *PrimingResource) AddContent(ctx context.Context, agentID, userID string, opts AddContentOptions) (*AddContentResponse, error) {
	var result AddContentResponse
	err := p.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/content", agentID, url.PathEscape(userID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMetadata returns the priming metadata for a user.
func (p *PrimingResource) GetMetadata(ctx context.Context, agentID, userID string) (*UserPrimingMetadata, error) {
	var result UserPrimingMetadata
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/metadata", agentID, url.PathEscape(userID)), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateMetadata partially updates priming metadata for a user.
func (p *PrimingResource) UpdateMetadata(ctx context.Context, agentID, userID string, opts UpdateMetadataOptions) (*UpdateMetadataResponse, error) {
	var result UpdateMetadataResponse
	err := p.http.Patch(ctx, fmt.Sprintf("/api/v1/agents/%s/users/%s/metadata", agentID, url.PathEscape(userID)), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchImport imports multiple users with metadata and content.
func (p *PrimingResource) BatchImport(ctx context.Context, agentID string, opts BatchImportOptions) (*BatchImportResponse, error) {
	var result BatchImportResponse
	err := p.http.Post(ctx, fmt.Sprintf("/api/v1/agents/%s/users/import", agentID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetImportStatus returns the status of a batch import job.
func (p *PrimingResource) GetImportStatus(ctx context.Context, agentID, jobID string) (*ImportJob, error) {
	var result ImportJob
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/import/%s", agentID, jobID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListImportJobs returns recent import jobs for an agent.
func (p *PrimingResource) ListImportJobs(ctx context.Context, agentID string, limit int) (*ImportJobListResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result ImportJobListResponse
	err := p.http.Get(ctx, fmt.Sprintf("/api/v1/agents/%s/users/imports", agentID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
