# EditionConfig V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在现有 `system_settings` JSONB 表上实现部署级 EditionConfig 预置配置、合并回退和管理员 API。

**Architecture:** 使用 `domain` 定义持久化 envelope、最终配置和请求 DTO；使用 `usecase` 负责预置配置、严格校验、合并和 common 回退；使用 `handler/v1` 复用现有 JWT 管理员中间件暴露 GET/PUT。数据库只存 `edition_config` envelope，不复制预置配置。

**Tech Stack:** Go 1.24、Echo、GORM、PostgreSQL JSONB、现有 `log.Logger` 和 wire 注入。

## Global Constraints

- 第一阶段只实现后端 EditionConfig。
- 不修改管理端和用户前端。
- 不修改 PromptRepo。
- 不实现 ReportProfile。
- 不修改知识库、文档、权限、RAG 和向量数据。
- 不新增 Edition 数据表。
- 不覆盖已有 system_settings 配置。
- 读取缺失、损坏或未知版本必须返回 common。
- PUT 只能接收明确的 `edition_id` 和 `overrides`，完整校验后原子写入。

### Task 1: Domain types and persisted envelope

**Files:**
- Create: `backend/domain/edition.go`
- Modify: `backend/consts/system_setting.go`

- [ ] Add typed constants and structs for `EditionID`, `EditionConfig`, `EditionOverrides`, `EditionStoredConfig`, and `UpdateEditionReq`.
- [ ] Add `SystemSettingEditionConfig = "edition_config"`.
- [ ] Add JSON tags matching the documented snake_case API contract.
- [ ] Keep persisted metadata (`schema_version`, `edition_id`, `edition_version`, `overrides`, `updated_at`, `updated_by`) separate from resolved public config.

### Task 2: Presets, validation, merge and fallback usecase

**Files:**
- Create: `backend/usecase/edition.go`
- Test: `backend/usecase/edition_test.go`

- [ ] Define `NewEditionUsecase(settingRepo, logger)` and methods `Get(ctx)` and `Update(ctx, userID, req)`.
- [ ] Add common/research/legal/finance presets with edition version `1.0.0`.
- [ ] Reject unknown edition IDs, unknown override keys, empty strings, invalid arrays, client metadata fields, and schema versions other than 1.
- [ ] Merge only allowed overrides onto the selected preset.
- [ ] Return common for missing/corrupt/unknown stored data without writing a replacement.
- [ ] On update, validate the resolved result before calling the repo; persist one envelope update only.
- [ ] Add tests for preset completeness, fallback, legal overrides, invalid requests, and database error propagation.

### Task 3: System-setting repository helpers

**Files:**
- Modify: `backend/repo/pg/system_setting.go`
- Test: `backend/repo/pg/system_setting_test.go` if repository test infrastructure exists

- [ ] Add typed `GetEditionStoredConfig` and `UpdateEditionStoredConfig` helpers using the existing `system_settings` row.
- [ ] Preserve all keys other than `edition_config`.
- [ ] Return `gorm.ErrRecordNotFound` for a missing edition row so the usecase can apply common fallback.
- [ ] Ensure update is one SQL update and does not create partial rows.

### Task 4: Administrator API and dependency wiring

**Files:**
- Create: `backend/handler/v1/edition.go`
- Modify: `backend/cmd/api/wire.go`
- Modify: generated wire file only if generation is available: `backend/cmd/api/wire_gen.go`
- Test: `backend/handler/v1/edition_test.go`

- [ ] Register `GET /api/v1/system/edition` and `PUT /api/v1/system/edition` under `Authorize` and `ValidateUserRole(UserRoleAdmin)`.
- [ ] Bind only `UpdateEditionReq`; reject malformed JSON and unknown fields according to existing Echo conventions.
- [ ] Return resolved `EditionConfig` from GET and PUT.
- [ ] Return 403 for non-admin requests.
- [ ] Log actor, before/after edition/version, and update failures without logging prompt secrets unnecessarily.

### Task 5: Verification and documentation handoff

**Files:**
- Modify: `docs/superpowers/specs/2026-07-23-edition-config-v1-design.md` only if implementation decisions diverge

- [ ] Run `gofmt` on changed Go files.
- [ ] Run `go test ./...` from `backend`.
- [ ] Run backend build/lint commands available in the repository.
- [ ] Confirm no frontend, PromptRepo, ReportProfile, KB, document, permission, RAG, or migration files changed.
- [ ] Report modified files, API behavior, tests, limitations, and next-stage integration points.
