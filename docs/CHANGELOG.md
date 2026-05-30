# IBEX Harness — Changelog

All notable changes to IBEX Harness will be documented in this file.

We follow:

- **Semantic Versioning** for the platform release tags: `vMAJOR.MINOR.PATCH`
- **URL-based API versioning** for REST: `/v1`, `/v2`
- **Additive evolution** for protobuf contracts (breaking changes require new package versions)

This changelog is:

- human-readable,
- operationally useful,
- and focused on user-visible changes, security fixes, and migrations.

---

## Changelog Rules

### 1) What goes in the changelog

Include:

- new features and behavior changes
- bug fixes that affect runtime behavior
- security fixes and mitigations
- performance improvements that affect SLOs
- migrations and breaking changes
- deprecations and removals
- operational changes (new env vars, new infra requirements)

Exclude:

- internal refactors with no behavior change (unless they reduce risk)
- trivial formatting changes
- dependency bumps unless security-critical or breaking

### 2) Format and discipline

- Every release must update the changelog as part of the release PR.
- Entries must be written in a way that users can understand.
- If change is security-sensitive, do not disclose exploit details before patch adoption.

### 3) Breaking change policy

Breaking changes require:

- a new MAJOR version OR a new REST API version
- a migration guide link
- explicit deprecation window (typically 12 months for API versions)

---

## [Unreleased]

### Added

- _TBD_

### Changed

- _TBD_

### Fixed

- _TBD_

### Security

- _TBD_

### Performance

- _TBD_

### Deprecated

- _TBD_

### Removed

- _TBD_

### Migration Notes

- _TBD_

---

## [0.1.0] — YYYY-MM-DD

Initial internal prototype release.

### Added

- Initial monorepo structure (`services/`, `packages/`, `infra/`, `docs/`)
- LLM Proxy service skeleton (Go)
- Auth service skeleton (Go)
- Memory service skeleton (Python/FastAPI)
- Context assembly service skeleton (Python/gRPC)
- PostgreSQL schema with RLS policies
- Redis cache and Streams for async jobs
- ClickHouse analytics schema (traces + billing events)
- MinIO object storage for session archives
- Basic CI pipeline (lint + typecheck + unit tests)
- Core documentation set in `docs/`

### Security

- Tenant isolation model defined (RLS + defense-in-depth)
- Secret scanning enabled in CI

### Migration Notes

- Apply initial database migrations
- Seed tier limits and dev org/user/agent fixtures

---

## [0.1.1] — YYYY-MM-DD

### Fixed

- (example) Fixed RLS context initialization bug that could cause empty query results

### Security

- (example) Prevented token-like strings from being logged in error paths

---

## [0.2.0] — YYYY-MM-DD

### Added

- (example) Memory semantic search endpoint with pgvector
- (example) Context injection formatting with nonce-delimited memory blocks
- (example) Worker pipeline for memory extraction (Celery)

### Changed

- (example) Proxy context assembly deadline reduced from 60ms → 40ms

### Performance

- (example) Reduced proxy overhead p95 from 28ms → 16ms by caching directive content

### Migration Notes

- Added `content_tokens` to `memories` for budget packing efficiency

---

## Release Entry Template (copy/paste)

```markdown
## [X.Y.Z] — YYYY-MM-DD

### Added
- ...

### Changed
- ...

### Fixed
- ...

### Security
- ...

### Performance
- ...

### Deprecated
- ...

### Removed
- ...

### Migration Notes
- ...
- New env vars:
  - `VAR_NAME` (service: ..., required: yes/no, default: ...)
- Backfill required:
  - ...
- Rollout strategy:
  - ...
- Rollback notes:
  - (DB migrations forward-only; app rollback must remain compatible)
```
