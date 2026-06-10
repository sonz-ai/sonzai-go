# Changelog

All notable changes to `github.com/sonz-ai/sonzai-go` are documented here. The
project follows [Semantic Versioning](https://semver.org/). Dates are `YYYY-MM-DD`.

## Unreleased

### Added

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
