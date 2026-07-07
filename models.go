package sonzai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ModelsResource provides access to trained model artifacts exported for
// local (on-runtime) inference — e.g. a tenant BFF/runtime that scores leads
// without a per-request upstream call, refreshing its cached artifact on an
// interval (design docs/design/platform-deployment-tiers.md §7.1).
type ModelsResource struct {
	http *httpClient
}

// ModelExportOptions configures a GET /api/v1/models/{useCase}/export request.
type ModelExportOptions struct {
	// Format optionally requests a specific export format (e.g. "onnx").
	// Empty requests the platform's default export for the use case.
	Format string
}

// ModelExport is a trained model artifact exported for local inference.
// Artifact/ArtifactB64/FeatureSchema/Calibration/Version are left as
// json.RawMessage / raw types deliberately: the artifact's internal shape is
// format-specific (Format selects e.g. "sonzai-gbdt-v1" vs "onnx-v1") and
// this resource has no opinion on it — callers parse Artifact/ArtifactB64
// per their own format handling, and Version may be a number or a string
// depending on the exporter, so callers unmarshal it into whatever type
// their format handling expects.
type ModelExport struct {
	UseCase           string          `json:"use_case"`
	Version           json.RawMessage `json:"version"`
	Format            string          `json:"format"`
	Artifact          json.RawMessage `json:"artifact,omitempty"`
	ArtifactB64       string          `json:"artifact_b64,omitempty"`
	FeatureSchema     json.RawMessage `json:"feature_schema,omitempty"`
	Calibration       json.RawMessage `json:"calibration,omitempty"`
	ActionCatalogHash string          `json:"action_catalog_hash,omitempty"`
	SHA256            string          `json:"sha256,omitempty"`
	ExportedAt        string          `json:"exported_at,omitempty"`
}

// Export fetches the current model artifact for useCase. A 404 (no artifact
// exported for this use case, or this specific format when opts.Format is
// set) surfaces as *NotFoundError.
func (m *ModelsResource) Export(ctx context.Context, useCase string, opts ModelExportOptions) (*ModelExport, error) {
	params := map[string]string{}
	if opts.Format != "" {
		params["format"] = opts.Format
	}
	var result ModelExport
	if err := m.http.Get(ctx, fmt.Sprintf("/api/v1/models/%s/export", url.PathEscape(useCase)), params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
