# Changelog

All notable changes to `github.com/sonz-ai/sonzai-go` are documented here. The
project follows [Semantic Versioning](https://semver.org/). Dates are `YYYY-MM-DD`.

## v1.5.2 — 2026-05-07

### Added

- New `BYOK` resource on the client (`client.BYOK`) exposing project-scoped
  bring-your-own-key management: `List`, `Set`, `Delete`, `SetActive`, and `Test`.
- `BYOKProvider` typed string enum with constants `BYOKProviderOpenAI`,
  `BYOKProviderGemini`, `BYOKProviderXAI`, and `BYOKProviderOpenRouter`.
- `BYOKKeyResponse` struct carrying `Provider`, `APIKeyPrefix`, `IsActive`,
  `HealthStatus`, and optional last-check / last-used timestamps. Key material
  is never returned by the API.
- Keys are validated against the provider's `/v1/models` endpoint before storage;
  upstream LLM billing for the project routes through the customer's key.
- REST surface: `GET/PUT/PATCH/DELETE /api/v1/projects/{project_id}/byok-keys[/{provider}]`
  and `POST /api/v1/projects/{project_id}/byok-keys/{provider}/test`.
