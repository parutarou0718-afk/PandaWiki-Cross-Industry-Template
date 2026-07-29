# Configurable Knowledge Schema Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Let PandaWiki administrators define knowledge fields that drive server-side graph extraction and the LLM Wiki Knowledge navigation and graph UI.

**Architecture:** Store one versioned, validated Schema per knowledge base and resolve it with a built-in default. Graph extraction receives only enabled field definitions and stores validated entity attributes in JSONB. Read APIs return the effective schema together with a permission-filtered graph projection; LLM Wiki renders those DTOs without local analysis.

**Tech Stack:** Go, GORM, PostgreSQL JSONB, NATS JetStream, React, TypeScript, Vitest.

## Global Constraints

- Never send remote documents to local filesystem or local graph analysis.
- The server derives access groups; no graph endpoint accepts `group_ids`.
- Prompts, raw model output, and full source documents are not persisted in graph attributes or returned to clients.
- Schema keys are stable lower-case snake case; values are server-validated.
- Schema rebuilds queue graph extraction only and do not re-vectorize documents.

---

### Task 1: Define Schema and Attribute Contracts

**Files:**
- Create: `backend/domain/knowledge_schema.go`
- Create: `backend/domain/knowledge_schema_test.go`
- Modify: `backend/domain/graph.go`

**Produces:** `KnowledgeSchema`, `KnowledgeField`, `KnowledgeNavigationSection`, `Validate`, `ValidateGraphAttributes`.

- [ ] Write tests that accept a text and select field, reject duplicate keys, unknown field values, invalid enum values, and an attribute with the wrong primitive type.
- [ ] Run `go test ./domain` and confirm failure because Schema types do not exist.
- [ ] Implement explicit value types (`text`, `number`, `date`, `boolean`, `select`), target types, bounded field descriptions, stable keys, and JSONB attributes.
- [ ] Run `go test ./domain` and confirm success.
- [ ] Commit domain contract changes.

### Task 2: Persist and Resolve Knowledge-Base Schemas

**Files:**
- Create: `backend/repo/pg/knowledge_schema.go`
- Create: `backend/store/pg/migration/000040_create_knowledge_graph_schema.up.sql`
- Create: `backend/store/pg/migration/000040_create_knowledge_graph_schema.down.sql`
- Modify: `backend/repo/pg/provider.go`

**Produces:** `GetEffectiveSchema`, `SaveSchema`, and a common fallback schema.

- [ ] Write a repository test for no-row fallback and schema update round-trip.
- [ ] Run the relevant repository test and confirm missing repository behavior.
- [ ] Add a JSONB schema table keyed by `kb_id`, with schema version and timestamps; do not add unbounded per-field columns.
- [ ] Implement a transactional save that validates before writing and a fallback that returns the built-in common schema without creating a row.
- [ ] Run repository/domain tests and confirm success.
- [ ] Commit persistence changes.

### Task 3: Make Graph Extraction Schema-Driven

**Files:**
- Modify: `backend/usecase/llm.go`
- Modify: `backend/usecase/graph.go`
- Modify: `backend/repo/pg/graph.go`
- Test: `backend/domain/knowledge_schema_test.go`

**Produces:** schema-derived extraction prompt and validated `graph_entities.attributes` values.

- [ ] Write tests that show disabled fields are absent from the prompt projection and unknown model attributes fail validation.
- [ ] Run `go test ./domain ./usecase` and confirm failure before changing extraction.
- [ ] Add `attributes JSONB` to graph entities through the migration; resolve schema inside `RefreshNode`, include enabled fields in the JSON-only model contract, and validate before replacing graph facts.
- [ ] Ensure an invalid model attribute leaves prior facts untouched by preserving the existing transaction boundary.
- [ ] Run `go test ./domain ./usecase ./repo/pg` and confirm success.
- [ ] Commit extraction changes.

### Task 4: Expose Secure Schema and Graph DTOs

**Files:**
- Modify: `backend/handler/v1/graph.go`
- Modify: `backend/handler/v1/provider.go`
- Test: `backend/handler/v1/graph_test.go`

**Produces:** GET/PUT Schema APIs and graph output containing effective navigation and safe attributes.

- [ ] Write handler tests for normal read, document-manager-only schema write, unknown JSON fields, and no client `group_ids` support.
- [ ] Run handler tests and confirm the routes are absent.
- [ ] Implement `GET /api/v1/knowledge_base/graph/schema`, `PUT /api/v1/knowledge_base/graph/schema`, and add navigation/schema to the graph read response.
- [ ] Keep rebuild behind `UserKBPermissionDocManage`; read uses normal knowledge-base permission and existing node visitability filtering.
- [ ] Run handler and backend build tests and confirm success.
- [ ] Commit API changes.

### Task 5: Render Server-Controlled Knowledge Navigation in LLM Wiki

**Files:**
- Create: `src/services/providers/pandawiki/api/knowledge-schema-api.ts`
- Create: `src/services/providers/pandawiki/dto/KnowledgeSchemaDTO.ts`
- Create: `src/services/providers/pandawiki/mapper/KnowledgeSchemaMapper.ts`
- Create: `src/components/providers/panda-wiki-knowledge-panel.tsx`
- Modify: `src/services/providers/contracts/GraphProvider.ts`
- Modify: `src/services/providers/pandawiki/PandaWikiProvider.ts`
- Modify: `src/components/layout/sidebar-panel.tsx`

**Produces:** remote Knowledge sections driven by server Schema and safe graph entities.

- [ ] Write mapping/component tests showing a server section controls labels and entity grouping, while an unknown attribute renders escaped text.
- [ ] Run targeted Vitest tests and confirm failure before adapters/components exist.
- [ ] Implement schema API, DTO mapper, provider method, and a remote-only knowledge panel that opens server nodes through `KnowledgeProvider`.
- [ ] Do not modify the local Knowledge panel or use local project paths.
- [ ] Run targeted Vitest and TypeScript checks and confirm success.
- [ ] Commit client navigation changes.

### Task 6: Upgrade the Remote Graph Surface

**Files:**
- Modify: `src/components/providers/panda-wiki-graph-view.tsx`
- Test: `src/components/providers/panda-wiki-graph-view.test.tsx`

**Produces:** graph filtering by server navigation sections and an entity attribute panel.

- [ ] Write a component test that filters only server-returned entities and opens an evidence source using the remote node provider.
- [ ] Run the test and confirm failure before interactive filter support.
- [ ] Add section filters and a property panel using server Schema labels; retain plain-text rendering for unknown attributes.
- [ ] Run full mock tests, typecheck, and production build.
- [ ] Commit client graph changes.

### Task 7: Final Verification and Deployment Notes

**Files:**
- Modify: `docs/knowledge-graph-v0.1.md`

- [ ] Run `gofmt`, relevant Go tests, `go vet ./...`, and `go build -buildvcs=false ./...`.
- [ ] Run `npm run test:mocks`, `npm run typecheck`, and `npm run build` in `research-work`.
- [ ] Document migration `000040`, graph-schema rebuild behavior, and the required server deployment order.
- [ ] Verify no local graph fallback, source-content leakage, prompt logging, or permission-group request parameters were added.
