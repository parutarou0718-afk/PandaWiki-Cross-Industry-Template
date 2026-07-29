# Knowledge Schema Admin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (or subagent-driven-development) task-by-task.

**Goal:** Add a KB Settings UI for full-control users to manage server-owned knowledge graph schema fields and navigation.

**Architecture:** Extend the existing settings tab and request client. The UI maps explicit TypeScript DTOs to the graph schema API; the backend remains responsible for validation and permissions.

**Tech Stack:** React, Material UI, existing generated request helper, Go graph schema API.

## Global Constraints

- Do not add local storage or client-side extraction.
- Preserve server validation and `full_control` write permission.
- Do not add relation/document schema targets in this iteration.

### Task 1: Add schema request client and typed editor helpers

**Files:**
- Create: `web/admin/src/api/KnowledgeSchema.ts`
- Create: `web/admin/src/pages/setting/component/knowledge-schema.ts`
- Test: `web/admin/src/pages/setting/component/knowledge-schema.test.ts`

- [ ] Write tests for default schema reset, duplicate-key validation and select-option validation.
- [ ] Implement explicit request types and pure validation helpers.
- [ ] Run the focused test.

### Task 2: Add Knowledge model settings tab

**Files:**
- Create: `web/admin/src/pages/setting/component/CardKnowledgeSchema.tsx`
- Modify: `web/admin/src/pages/setting/index.tsx`

- [ ] Build field and navigation editing controls that use the typed client.
- [ ] Add tab only within the existing KB settings route.
- [ ] Run admin TypeScript build.

### Task 3: Verify and commit

- [ ] Run backend graph schema tests and admin build.
- [ ] Commit UI and documentation with the related feature.
