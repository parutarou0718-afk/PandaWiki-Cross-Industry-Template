# ReportProfile V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` task-by-task.

**Goal:** Add template-driven, single-model reports using the authenticated user's existing RAG group filtering.

**Architecture:** Store reusable profiles and per-user report jobs in PostgreSQL. The report usecase obtains group IDs from `AuthRepo.GetAuthGroupIdsWithParentsByAuthId`, calls `LLMUsecase.GetRankNodes`, constructs a single model prompt from a profile and inputs, and persists only ranked-node-derived citations.

**Global constraints:** No multi-agent, network access, MCP, RAG permission changes, or document export. Do not log complete prompts or document chunks.

### Task 1: Persistence and presets

- [ ] Add `report_profiles`, `reports`, and `report_citations` migrations with JSONB profile inputs/sections and report input payloads.
- [ ] Add domain types and repository CRUD, including edition filtering and owner-scoped report queries.
- [ ] Seed common, research, legal, and finance profiles as code presets; store only admin overrides.
- [ ] Add failing repository/usecase tests for edition filtering, enabled state, profile version, and default restoration.

### Task 2: Permission-safe report generation

- [ ] Add a ReportUsecase that loads the current user group IDs via `GetAuthGroupIdsWithParentsByAuthId`.
- [ ] Pass those exact IDs to `LLMUsecase.GetRankNodes`; never query a dataset directly.
- [ ] Create pending/failed/succeeded report records, generate once with the selected model, and save citations from returned ranked nodes only.
- [ ] Add tests for owner scope, disabled profiles, no-result citation-required reports, failed generation, and group ID propagation.

### Task 3: HTTP API and Wire

- [ ] Add authenticated read/generate/history endpoints and admin-only profile mutation endpoints.
- [ ] Validate input fields against the profile schema and reject unknown fields.
- [ ] Generate Wire sources and add handler/usecase tests.

### Task 4: Admin configuration UI

- [ ] Add an administrator-only ReportProfile section under system settings.
- [ ] Support list/edit/enable/edition selection/default restore through the new API.

### Task 5: Public Wiki report UI

- [ ] Add a minimal authenticated report screen for profile selection, inputs, submit, status, result, citations, and own-history only.
- [ ] Use the existing app/web/info edition value only for profile display; do not add a separate Edition request.

### Task 6: Verification

- [ ] Run Go formatting, Wire generation, focused and all non-Discord tests, vet, and build.
- [ ] Run admin/app TypeScript checks, Prettier, admin Vite build, and app build; record the known Windows standalone symlink limitation if it recurs.
