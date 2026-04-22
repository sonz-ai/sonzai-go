package sonzai

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

// KnowledgeResource provides knowledge base operations for a project.
// Knowledge endpoints are project-scoped (not agent-scoped).
type KnowledgeResource struct {
	http *httpClient
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// KBDocument represents a document uploaded to the knowledge base.
type KBDocument struct {
	ProjectID        string `json:"project_id"`
	DocumentID       string `json:"document_id"`
	FileName         string `json:"file_name"`
	ContentType      string `json:"content_type"`
	FileSize         int64  `json:"file_size"`
	GCSPath          string `json:"gcs_path"`
	Checksum         string `json:"checksum"`
	Status           string `json:"status"` // "pending", "parsing", "extracting", "indexed", "failed"
	UploadedBy       string `json:"uploaded_by,omitempty"`
	ExtractionTokens int    `json:"extraction_tokens,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

// KBDocumentListResponse is the response from listing documents.
type KBDocumentListResponse struct {
	Documents []*KBDocument `json:"documents"`
	Total     int           `json:"total"`
}

// KBNode represents a node in the knowledge graph.
type KBNode struct {
	ProjectID  string         `json:"project_id"`
	NodeID     string         `json:"node_id"`
	NodeType   string         `json:"node_type"`
	Label      string         `json:"label"`
	NormLabel  string         `json:"norm_label,omitempty"`
	Properties map[string]any `json:"properties"`
	SourceType string         `json:"source_type"`
	Version    int            `json:"version"`
	IsActive   bool           `json:"is_active"`
	Confidence float64        `json:"confidence"`
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
}

// KBNodeListResponse is the response from listing nodes.
type KBNodeListResponse struct {
	Nodes      []*KBNode `json:"nodes"`
	Total      int       `json:"total"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

// KBEdge represents an edge in the knowledge graph.
type KBEdge struct {
	ProjectID  string  `json:"project_id"`
	EdgeID     string  `json:"edge_id"`
	FromNodeID string  `json:"from_node_id"`
	ToNodeID   string  `json:"to_node_id"`
	EdgeType   string  `json:"edge_type"`
	Confidence float64 `json:"confidence"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
}

// KBNodeHistory represents a version history entry for a node.
type KBNodeHistory struct {
	ProjectID  string         `json:"project_id"`
	NodeID     string         `json:"node_id"`
	Version    int            `json:"version"`
	Properties map[string]any `json:"properties"`
	ChangedBy  string         `json:"changed_by"`
	ChangeType string         `json:"change_type"`
	ChangedAt  string         `json:"changed_at,omitempty"`
}

// KBNodeDetailResponse is the response from getting a single node.
type KBNodeDetailResponse struct {
	Node     *KBNode          `json:"node"`
	Outgoing []*KBEdge        `json:"outgoing"`
	Incoming []*KBEdge        `json:"incoming"`
	History  []*KBNodeHistory `json:"history"`
}

// KBNodeHistoryResponse is the response from getting node history.
type KBNodeHistoryResponse struct {
	History []*KBNodeHistory `json:"history"`
	Total   int              `json:"total"`
}

// KBRelatedNode represents a related node in a search result.
type KBRelatedNode struct {
	NodeID     string         `json:"node_id"`
	Label      string         `json:"label"`
	NodeType   string         `json:"node_type"`
	EdgeType   string         `json:"edge_type"`
	Properties map[string]any `json:"properties,omitempty"`
}

// KBSearchResult represents a single search result.
type KBSearchResult struct {
	NodeID     string          `json:"node_id"`
	NodeType   string          `json:"node_type"`
	Label      string          `json:"label"`
	Properties map[string]any  `json:"properties"`
	Source     string          `json:"source"`
	UpdatedAt  string          `json:"updated_at"`
	Score      float64         `json:"score"`
	Related    []KBRelatedNode `json:"related,omitempty"`
	History    []KBNodeHistory `json:"history,omitempty"`
}

// KBSearchResponse is the response from the knowledge search endpoint.
type KBSearchResponse struct {
	Query   string           `json:"query"`
	Results []KBSearchResult `json:"results"`
	Total   int              `json:"total"`
}

// KBSchemaField represents a field in an entity schema.
type KBSchemaField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	EnumValues  []string `json:"enum_values,omitempty"`
}

// KBSimilarityConfig configures similarity matching for a schema.
type KBSimilarityConfig struct {
	MatchFields []string `json:"match_fields,omitempty"`
	Threshold   float64  `json:"threshold,omitempty"`
}

// KBEntitySchema represents an entity type schema.
type KBEntitySchema struct {
	ProjectID        string              `json:"project_id"`
	SchemaID         string              `json:"schema_id"`
	EntityType       string              `json:"entity_type"`
	Fields           []KBSchemaField     `json:"fields"`
	Description      string              `json:"description,omitempty"`
	SimilarityConfig *KBSimilarityConfig `json:"similarity_config,omitempty"`
	CreatedAt        string              `json:"created_at,omitempty"`
	UpdatedAt        string              `json:"updated_at,omitempty"`
}

// KBSchemaListResponse is the response from listing schemas.
type KBSchemaListResponse struct {
	Schemas []*KBEntitySchema `json:"schemas"`
	Total   int               `json:"total"`
}

// KBStats represents knowledge base statistics.
type KBStats struct {
	Documents struct {
		Total   int `json:"total"`
		Indexed int `json:"indexed"`
		Pending int `json:"pending"`
		Failed  int `json:"failed"`
	} `json:"documents"`
	Nodes struct {
		Total  int `json:"total"`
		Active int `json:"active"`
	} `json:"nodes"`
	Edges            int `json:"edges"`
	ExtractionTokens int `json:"extraction_tokens"`
}

// KBAnalyticsRule represents an analytics rule.
type KBAnalyticsRule struct {
	ProjectID string `json:"project_id"`
	RuleID    string `json:"rule_id"`
	RuleType  string `json:"rule_type"` // "recommendation" or "trend"
	Name      string `json:"name"`
	Config    any    `json:"config"`
	Enabled   bool   `json:"enabled"`
	Schedule  string `json:"schedule,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// KBAnalyticsRuleListResponse is the response from listing analytics rules.
type KBAnalyticsRuleListResponse struct {
	Rules []*KBAnalyticsRule `json:"rules"`
	Total int                `json:"total"`
}

// KBRecommendationScore represents a recommendation score.
type KBRecommendationScore struct {
	ProjectID  string  `json:"project_id"`
	RuleID     string  `json:"rule_id"`
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	TargetType string  `json:"target_type"`
	Score      float64 `json:"score"`
}

// KBRecommendationsResponse is the response from getting recommendations.
type KBRecommendationsResponse struct {
	Recommendations []*KBRecommendationScore `json:"recommendations"`
	Total           int                      `json:"total"`
}

// KBTrendAggregation represents a trend aggregation for a node.
type KBTrendAggregation struct {
	ProjectID string  `json:"project_id"`
	NodeID    string  `json:"node_id"`
	RuleID    string  `json:"rule_id"`
	Window    string  `json:"window"`
	Value     float64 `json:"value"`
	Direction string  `json:"direction"`
}

// KBTrendsResponse is the response from getting trends.
type KBTrendsResponse struct {
	Trends []*KBTrendAggregation `json:"trends"`
	Total  int                   `json:"total"`
}

// KBTrendRanking represents a trend ranking entry.
type KBTrendRanking struct {
	ProjectID string  `json:"project_id"`
	NodeID    string  `json:"node_id"`
	RuleID    string  `json:"rule_id"`
	Type      string  `json:"type"`
	Window    string  `json:"window"`
	Rank      int     `json:"rank"`
	Score     float64 `json:"score"`
}

// KBTrendRankingsResponse is the response from getting trend rankings.
type KBTrendRankingsResponse struct {
	Rankings []*KBTrendRanking `json:"rankings"`
	Total    int               `json:"total"`
}

// KBConversionStats represents conversion statistics.
type KBConversionStats struct {
	ProjectID       string  `json:"project_id"`
	RuleID          string  `json:"rule_id"`
	SegmentKey      string  `json:"segment_key"`
	TargetType      string  `json:"target_type"`
	ShownCount      int     `json:"shown_count"`
	ConversionCount int     `json:"conversion_count"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// KBConversionsResponse is the response from getting conversion stats.
type KBConversionsResponse struct {
	Conversions []*KBConversionStats `json:"conversions"`
	Total       int                  `json:"total"`
}

// ---------------------------------------------------------------------------
// Request Options
// ---------------------------------------------------------------------------

// InsertFactsOptions configures a fact insertion request.
type InsertFactsOptions struct {
	Source        string            `json:"source,omitempty"`
	Facts         []InsertFactEntry `json:"facts"`
	Relationships []InsertRelEntry  `json:"relationships,omitempty"`
}

// InsertFactEntry represents a single entity to insert.
type InsertFactEntry struct {
	EntityType string         `json:"entity_type"`
	Label      string         `json:"label"`
	Properties map[string]any `json:"properties,omitempty"`
}

// InsertRelEntry represents a relationship to insert.
type InsertRelEntry struct {
	FromLabel string `json:"from_label"`
	ToLabel   string `json:"to_label"`
	EdgeType  string `json:"edge_type"`
}

// InsertFactEdgeDetail describes a created relationship edge.
type InsertFactEdgeDetail struct {
	EdgeID   string `json:"edge_id"`
	FromNode string `json:"from_node"`
	ToNode   string `json:"to_node"`
	Relation string `json:"relation"`
}

// InsertFactsResponse is the response from inserting facts.
type InsertFactsResponse struct {
	Processed int                    `json:"processed"`
	Created   int                    `json:"created"`
	Updated   int                    `json:"updated"`
	Details   []InsertFactDetail     `json:"details"`
	Edges     []InsertFactEdgeDetail `json:"edges,omitempty"`
}

// InsertFactDetail describes an individual inserted/updated fact.
type InsertFactDetail struct {
	Label   string `json:"label"`
	Type    string `json:"type"`
	Action  string `json:"action"` // "created" or "updated"
	NodeID  string `json:"node_id"`
	Version int    `json:"version"`
}

// ListNodesOptions configures a list nodes request with filtering, pagination, and sorting.
type ListNodesOptions struct {
	NodeType   string            `json:"node_type,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
	SortBy     string            `json:"sort_by,omitempty"`
	SortOrder  string            `json:"sort_order,omitempty"`
}

// KBSearchOptions configures a knowledge search request.
type KBSearchOptions struct {
	Query          string
	Limit          int
	IncludeHistory bool
	EntityTypes    string // comma-separated
	Filters        string // JSON string
	Depth          int    // graph traversal depth (query param: depth)
}

// CreateSchemaOptions configures a schema creation request.
type CreateSchemaOptions struct {
	EntityType       string              `json:"entity_type"`
	Fields           []KBSchemaField     `json:"fields"`
	Description      string              `json:"description,omitempty"`
	SimilarityConfig *KBSimilarityConfig `json:"similarity_config,omitempty"`
}

// CreateAnalyticsRuleOptions configures an analytics rule creation request.
type CreateAnalyticsRuleOptions struct {
	RuleType string `json:"rule_type"` // "recommendation" or "trend"
	Name     string `json:"name"`
	Config   any    `json:"config"`
	Enabled  bool   `json:"enabled"`
	Schedule string `json:"schedule,omitempty"`
}

// UpdateAnalyticsRuleOptions configures an analytics rule update request.
type UpdateAnalyticsRuleOptions struct {
	Name     string `json:"name,omitempty"`
	Config   any    `json:"config,omitempty"`
	Enabled  bool   `json:"enabled"`
	Schedule string `json:"schedule,omitempty"`
}

// RecordFeedbackOptions configures a feedback recording request.
type RecordFeedbackOptions struct {
	SourceNodeID string  `json:"source_node_id"`
	TargetNodeID string  `json:"target_node_id"`
	RuleID       string  `json:"rule_id"`
	Converted    bool    `json:"converted"`
	ScoreAtTime  float64 `json:"score_at_time"`
}

// ---------------------------------------------------------------------------
// Document Methods
// ---------------------------------------------------------------------------

// ListDocuments returns documents for a project.
func (k *KnowledgeResource) ListDocuments(ctx context.Context, projectID string, limit int) (*KBDocumentListResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result KBDocumentListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/documents", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDocument returns a single document.
func (k *KnowledgeResource) GetDocument(ctx context.Context, projectID, docID string) (*KBDocument, error) {
	var result KBDocument
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/documents/%s", projectID, docID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadDocumentOptions configures a document upload request.
type UploadDocumentOptions struct {
	FileName string
	FileData []byte
}

// UploadDocumentBytes uploads a document to the knowledge base from an in-memory byte slice.
// For streaming uploads from an io.Reader, use UploadDocument instead.
func (k *KnowledgeResource) UploadDocumentBytes(ctx context.Context, projectID string, opts UploadDocumentOptions) (*KBDocument, error) {
	var result KBDocument
	err := k.http.PostMultipartFile(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/documents", projectID), "file", opts.FileName, opts.FileData, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDocument deletes a document.
func (k *KnowledgeResource) DeleteDocument(ctx context.Context, projectID, docID string) error {
	return k.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/documents/%s", projectID, docID), nil)
}

// UploadDocument uploads a file to the knowledge base, triggering the AI extraction pipeline.
func (k *KnowledgeResource) UploadDocument(ctx context.Context, projectID string, fileName string, fileContent io.Reader, contentType string) (*KBDocument, error) {
	var result KBDocument
	err := k.http.PostMultipart(ctx,
		fmt.Sprintf("/api/v1/projects/%s/knowledge/documents", projectID),
		nil,
		fileName,
		fileContent,
		contentType,
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Facts / Graph Methods
// ---------------------------------------------------------------------------

// InsertFacts inserts entities and relationships into the knowledge graph.
func (k *KnowledgeResource) InsertFacts(ctx context.Context, projectID string, opts InsertFactsOptions) (*InsertFactsResponse, error) {
	var result InsertFactsResponse
	err := k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/facts", projectID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListNodes returns knowledge graph nodes for a project.
func (k *KnowledgeResource) ListNodes(ctx context.Context, projectID string, opts *ListNodesOptions) (*KBNodeListResponse, error) {
	params := map[string]string{}
	if opts != nil {
		if opts.NodeType != "" {
			params["type"] = opts.NodeType
		}
		if opts.Limit > 0 {
			params["limit"] = strconv.Itoa(opts.Limit)
		}
		if opts.Offset > 0 {
			params["offset"] = strconv.Itoa(opts.Offset)
		}
		if opts.SortBy != "" {
			params["sort_by"] = opts.SortBy
		}
		if opts.SortOrder != "" {
			params["sort_order"] = opts.SortOrder
		}
		for k, v := range opts.Properties {
			params["properties."+k] = v
		}
	}
	var result KBNodeListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListNodesWithOptions returns knowledge graph nodes with filtering, pagination, and sorting.
func (k *KnowledgeResource) ListNodesWithOptions(ctx context.Context, projectID string, opts ListNodesOptions) (*KBNodeListResponse, error) {
	params := map[string]string{}
	if opts.NodeType != "" {
		params["type"] = opts.NodeType
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.Offset > 0 {
		params["offset"] = strconv.Itoa(opts.Offset)
	}
	if opts.SortBy != "" {
		params["sort_by"] = opts.SortBy
	}
	if opts.SortOrder != "" {
		params["sort_order"] = opts.SortOrder
	}
	for k, v := range opts.Properties {
		params["properties."+k] = v
	}
	var result KBNodeListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetNode returns a single node with its connected edges.
func (k *KnowledgeResource) GetNode(ctx context.Context, projectID, nodeID string, includeHistory bool) (*KBNodeDetailResponse, error) {
	params := map[string]string{}
	if includeHistory {
		params["history"] = "true"
	}
	var result KBNodeDetailResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes/%s", projectID, nodeID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteNode soft-deletes a node.
func (k *KnowledgeResource) DeleteNode(ctx context.Context, projectID, nodeID string) error {
	return k.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes/%s", projectID, nodeID), nil)
}

// GetNodeHistory returns version history for a node.
func (k *KnowledgeResource) GetNodeHistory(ctx context.Context, projectID, nodeID string, limit int) (*KBNodeHistoryResponse, error) {
	params := map[string]string{}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result KBNodeHistoryResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/nodes/%s/history", projectID, nodeID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// Search performs a BM25 search with 1-hop graph traversal.
func (k *KnowledgeResource) Search(ctx context.Context, projectID string, opts KBSearchOptions) (*KBSearchResponse, error) {
	params := map[string]string{
		"q": opts.Query,
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}
	if opts.IncludeHistory {
		params["history"] = "true"
	}
	if opts.EntityTypes != "" {
		params["type"] = opts.EntityTypes
	}
	if opts.Filters != "" {
		params["filters"] = opts.Filters
	}
	if opts.Depth > 0 {
		params["depth"] = strconv.Itoa(opts.Depth)
	}
	var result KBSearchResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/search", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Schema Methods
// ---------------------------------------------------------------------------

// CreateSchema creates an entity type schema.
func (k *KnowledgeResource) CreateSchema(ctx context.Context, projectID string, opts CreateSchemaOptions) (*KBEntitySchema, error) {
	var result KBEntitySchema
	err := k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/schemas", projectID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSchemas returns entity schemas for a project.
func (k *KnowledgeResource) ListSchemas(ctx context.Context, projectID string) (*KBSchemaListResponse, error) {
	var result KBSchemaListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/schemas", projectID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSchema updates an entity type schema.
func (k *KnowledgeResource) UpdateSchema(ctx context.Context, projectID, schemaID string, opts CreateSchemaOptions) (*KBEntitySchema, error) {
	var result KBEntitySchema
	err := k.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/schemas/%s", projectID, schemaID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSchema deletes an entity type schema.
func (k *KnowledgeResource) DeleteSchema(ctx context.Context, projectID, schemaID string) error {
	return k.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/schemas/%s", projectID, schemaID), nil)
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

// GetStats returns knowledge base statistics.
func (k *KnowledgeResource) GetStats(ctx context.Context, projectID string) (*KBStats, error) {
	var result KBStats
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/stats", projectID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Analytics Rules
// ---------------------------------------------------------------------------

// CreateAnalyticsRule creates an analytics rule.
func (k *KnowledgeResource) CreateAnalyticsRule(ctx context.Context, projectID string, opts CreateAnalyticsRuleOptions) (*KBAnalyticsRule, error) {
	var result KBAnalyticsRule
	err := k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules", projectID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAnalyticsRules returns analytics rules for a project.
func (k *KnowledgeResource) ListAnalyticsRules(ctx context.Context, projectID string) (*KBAnalyticsRuleListResponse, error) {
	var result KBAnalyticsRuleListResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules", projectID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAnalyticsRule returns a single analytics rule.
func (k *KnowledgeResource) GetAnalyticsRule(ctx context.Context, projectID, ruleID string) (*KBAnalyticsRule, error) {
	var result KBAnalyticsRule
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules/%s", projectID, ruleID), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateAnalyticsRule updates an analytics rule.
func (k *KnowledgeResource) UpdateAnalyticsRule(ctx context.Context, projectID, ruleID string, opts UpdateAnalyticsRuleOptions) (*KBAnalyticsRule, error) {
	var result KBAnalyticsRule
	err := k.http.Put(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules/%s", projectID, ruleID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAnalyticsRule deletes an analytics rule.
func (k *KnowledgeResource) DeleteAnalyticsRule(ctx context.Context, projectID, ruleID string) error {
	return k.http.Delete(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules/%s", projectID, ruleID), nil)
}

// RunAnalyticsRule triggers a manual run of an analytics rule.
func (k *KnowledgeResource) RunAnalyticsRule(ctx context.Context, projectID, ruleID string) error {
	var result map[string]any
	return k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rules/%s/run", projectID, ruleID), nil, &result)
}

// ---------------------------------------------------------------------------
// Analytics Queries
// ---------------------------------------------------------------------------

// GetRecommendations returns recommendations for a source node.
func (k *KnowledgeResource) GetRecommendations(ctx context.Context, projectID, ruleID, sourceID string, limit int) (*KBRecommendationsResponse, error) {
	params := map[string]string{
		"rule_id":   ruleID,
		"source_id": sourceID,
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result KBRecommendationsResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/recommendations", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTrends returns trend aggregations for a node.
func (k *KnowledgeResource) GetTrends(ctx context.Context, projectID, nodeID string) (*KBTrendsResponse, error) {
	params := map[string]string{
		"node_id": nodeID,
	}
	var result KBTrendsResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/trends", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTrendRankings returns trend rankings.
func (k *KnowledgeResource) GetTrendRankings(ctx context.Context, projectID, ruleID, rankingType, window string, limit int) (*KBTrendRankingsResponse, error) {
	params := map[string]string{
		"rule_id": ruleID,
		"type":    rankingType,
		"window":  window,
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var result KBTrendRankingsResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/rankings", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConversions returns conversion statistics.
func (k *KnowledgeResource) GetConversions(ctx context.Context, projectID, ruleID, segment string) (*KBConversionsResponse, error) {
	params := map[string]string{
		"rule_id": ruleID,
	}
	if segment != "" {
		params["segment"] = segment
	}
	var result KBConversionsResponse
	err := k.http.Get(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/conversions", projectID), params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RecordFeedback records recommendation feedback (shown/converted).
func (k *KnowledgeResource) RecordFeedback(ctx context.Context, projectID string, opts RecordFeedbackOptions) error {
	var result map[string]any
	return k.http.Post(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/analytics/feedback", projectID), opts, &result)
}

// KBBulkUpdateEntry represents a single entry in a bulk update request.
type KBBulkUpdateEntry struct {
	EntityType string                 `json:"entity_type"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties"`
}

// KBBulkUpdateOptions configures a bulk update request.
type KBBulkUpdateOptions struct {
	Source  string              `json:"source,omitempty"`
	Updates []KBBulkUpdateEntry `json:"updates"`
}

// KBBulkUpdateResponse is the response from a bulk update.
type KBBulkUpdateResponse struct {
	Processed int    `json:"processed,omitempty"`
	Updated   int    `json:"updated,omitempty"`
	NotFound  int    `json:"not_found,omitempty"`
	Created   int    `json:"created,omitempty"`
	Status    string `json:"status,omitempty"`
	Count     int    `json:"count,omitempty"`
}

// BulkUpdate batch-updates KB node properties. Sync for <=100 items; async for larger.
func (k *KnowledgeResource) BulkUpdate(ctx context.Context, projectID string, opts KBBulkUpdateOptions) (*KBBulkUpdateResponse, error) {
	var result KBBulkUpdateResponse
	err := k.http.Patch(ctx, fmt.Sprintf("/api/v1/projects/%s/knowledge/bulk-update", projectID), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
