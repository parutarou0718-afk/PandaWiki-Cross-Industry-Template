# Graph Entity Summary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Precompute permission-safe server entity summaries during graph extraction and show them in the desktop remote graph inspector.

**Architecture:** Graph extraction returns an optional concise summary for every extracted entity in the context of its source node. PandaWiki stores those source-scoped summaries. The permission-filtered graph query combines only summaries from source nodes visible to the requesting user; research-work maps and renders the returned value without a local model call.

**Tech Stack:** Go, GORM, PostgreSQL migrations, NATS graph consumer, React, TypeScript, Vitest.

## Global Constraints

- Never aggregate restricted and public source facts into one globally visible summary.
- No client LLM call, local graph cache, raw source body, new credentials, or new API endpoint.
- Summary generation stays in the existing asynchronous graph extraction flow.
- A missing or failed summary must not remove graph facts or make graph extraction fail.
- Log only IDs and counts; never full summary text, prompts, or source content.
- Local projects keep their current graph behavior.

---

### Task 1: Source-scoped summary storage

**Files:**
- Create: `backend/store/pg/migration/000041_create_graph_entity_summaries.up.sql`
- Create: `backend/store/pg/migration/000041_create_graph_entity_summaries.down.sql`
- Modify: `backend/domain/graph.go`
- Modify: `backend/domain/graph_test.go`
- Modify: `backend/repo/pg/graph.go`

**Produces:** `GraphEntitySummary` keyed by entity, knowledge base, source node, and release; `GraphExtractedEntity.Summary` capped at 600 runes.

- [ ] Write a failing domain test for retaining a valid short summary, allowing empty summary, and dropping an over-limit summary while retaining valid facts.
- [ ] Run `go test ./domain -run 'Graph.*Summary|SanitizeGraphExtraction' -count=1`; expect failure.
- [ ] Add migration 000041: `graph_entity_summaries`, unique `(entity_id,node_id)`, indexes `(kb_id,node_id)` and `entity_id`, entity foreign key with `ON DELETE CASCADE`.
- [ ] Add domain validation and sanitizer behavior for optional plain-text source summaries.
- [ ] In the existing `ReplaceNodeExtraction` transaction, delete summaries for `(kb_id,node_id)` and upsert summaries only for extracted entities with non-empty summary after their stable entity IDs are resolved.
- [ ] Re-run the targeted test; expect PASS.
- [ ] Commit only Task 1 files with `feat: persist graph entity source summaries`.

### Task 2: Generate and expose only visible summaries

**Files:**
- Modify: `backend/usecase/llm.go`
- Modify: `backend/usecase/graph.go`
- Modify: `backend/repo/pg/graph.go`
- Modify: `backend/domain/graph_test.go`
- Test: `backend/repo/pg/graph_test.go` or existing PostgreSQL integration suite

**Consumes:** Task 1 source-scoped storage.

**Produces:** `GET /api/v1/knowledge_base/graph` returns optional `entities[].summary` created exclusively from nodes passing its current visit-permission SQL predicate.

- [ ] Write a failing extraction test accepting `{"name":"…","type":"…","summary":"…","attributes":{}}` and a visible-graph test with one public plus one restricted source summary for the same entity.
- [ ] Run `go test ./domain ./usecase ./repo/pg -run 'Graph.*Summary|VisibleGraph' -count=1`; expect failure.
- [ ] Extend the existing graph extraction prompt: optional `summary`, plain text, at most 600 characters, describes only the entity in this document; omission remains valid.
- [ ] Query `graph_entity_summaries` through the exact visible-node predicate already used for evidence; group deterministically, de-duplicate, join at most three visible source summaries, and bound the assembled entity summary.
- [ ] Preserve previous graph facts if summary generation is empty or fails; log only `kb_id`, `node_id`, and counts.
- [ ] Re-run focused tests; expect PASS.
- [ ] Commit only Task 2 files with `feat: expose permission-safe graph entity summaries`.

### Task 3: Desktop summary presentation

**Files:**
- Modify: `src/types/wiki.ts`
- Modify: `src/services/providers/pandawiki/dto/GraphDTO.ts`
- Modify: `src/services/providers/pandawiki/mapper/GraphMapper.ts`
- Modify: `src/components/providers/panda-wiki-graph-view.tsx`
- Test: `src/services/providers/pandawiki/mapper/GraphMapper.test.ts`
- Test: `src/components/providers/panda-wiki-graph-view.test.tsx`

**Consumes:** optional `summary` supplied by the normal graph response.

**Produces:** a text-only `服务器分析` block between Schema attributes and evidence cards; an empty state says `暂无服务器摘要` and makes no network or LLM request.

- [ ] Write failing mapper/inspector tests for a server summary and an empty summary.
- [ ] Run `npm test -- --run src/services/providers/pandawiki/mapper/GraphMapper.test.ts src/components/providers/panda-wiki-graph-view.test.tsx`; expect failure.
- [ ] Add optional summary to DTO, domain model, mapper, and inspector. Do not add a request or cache.
- [ ] Re-run focused tests, then `npm run typecheck` and `npm run build`; expect PASS.
- [ ] Commit only Task 3 files with `feat: show server graph entity summaries`.

### Task 4: Verification and deployment guidance

**Files:**
- Modify: `deploy/README.md` or the existing canonical graph deployment guide

- [ ] Run `gofmt` for modified Go files.
- [ ] Run `go test ./domain ./usecase ./repo/pg ./handler/mq -count=1`, `go vet ./domain ./usecase ./repo/pg ./handler/mq`, and `go build -buildvcs=false ./cmd/api ./cmd/consumer`.
- [ ] Document that normal migration applies 000041; API and Consumer must be rebuilt/recreated; an explicit graph rebuild generates summaries for existing facts.
- [ ] Push backend `feature/pandawiki-graph-v0.1` and desktop `feature/pandawiki-graph-workspace-ui`; never push credentials, build artifacts, token stores, or `.env` files.
