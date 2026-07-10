# Changelog

All notable changes to `github.com/sonz-ai/sonzai-go` are documented here. The
project follows [Semantic Versioning](https://semver.org/). Dates are `YYYY-MM-DD`.

## 1.8.0 — 2026-07-10

### Added

- New `Crm` resource on the client (`client.Crm`) for adapter-token access to
  a deployed app-runtime's runtime-local CRM. Configure its target with
  `WithRuntimeBaseURL` or `SONZAI_RUNTIME_BASE_URL`; it uses that runtime base
  URL rather than `https://api.sonz.ai`.
- Runtime CRM adapter routes: bulk contact upsert through
  `POST /api/rt/crm/import` (idempotent by `external_ref`) and the
  cursor-paginated change feed at `GET /api/rt/crm/events`, including the
  `EventIterator` pull helper.
- Runtime CRM request and response types for contacts, companies, pipelines,
  stages, deals, activities, custom fields, imports, and change-feed events.
  Managed/shared runtimes can receive `X-Sonzai-Tenant-ID` through the CRM
  import and events options.
- The adapter-token surface intentionally excludes staff CRM CRUD routes,
  which require browser-session authentication in the runtime.

- New `BuiltinAgents` resource on the client (`client.BuiltinAgents`) for
  Sonzai Built-in Agents — platform-hosted vertical task agents
  (`lead_research`, `market_intel`, `lead_extract`, `lead_score`,
  `lead_qualifier`): `List`, `Invoke`, `InvokeStream`, `CreateSession`,
  `ListSessions`, `GetSession`, and `SendMessage`.
- Slug constants `BuiltinAgentLeadResearch`, `BuiltinAgentMarketIntel`,
  `BuiltinAgentLeadExtract`, `BuiltinAgentLeadScore`, and
  `BuiltinAgentLeadQualifier`.
- `Invoke` and `SendMessage` run without the client-level HTTP timeout —
  long invocations (15+ minutes) are capped only by the caller's context.
- `InvokeStream` / streaming `SendMessage` parse the named SSE envelope
  (`update` progress frames, terminal `result` or `error`).
- REST surface: `GET /api/v1/builtin-agents`,
  `POST /api/v1/builtin-agents/{slug}/invoke?stream=<bool>`, and
  `POST/GET /api/v1/builtin-agents/sessions[/{id}[/messages]]`.

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
