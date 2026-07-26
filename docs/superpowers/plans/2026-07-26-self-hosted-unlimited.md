# Self-Hosted Unlimited Edition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make all former paid-edition functionality and quotas available in every self-hosted installation while retaining ordinary administrator, knowledge-base, group, and node authorization.

**Architecture:** Preserve license DTOs and database compatibility, but make the effective server policy unlimited and capability-complete. Normalize the legacy edition value for compatibility, neutralize UI upgrade gates, and keep authorization middleware and permission checks untouched. The existing `ADMIN_PASSWORD` bootstrap remains the deployment-only route to establish the `admin` account; it is corrected to always preserve the admin role.

**Tech Stack:** Go, Echo, GORM/PostgreSQL, React, TypeScript, MUI, pnpm, Go testing.

## Global Constraints

- Never bypass authentication, `UserRoleAdmin`, knowledge-base membership, user-group permissions, or node permissions.
- Do not add a web root/backdoor endpoint or a compiled-in credential.
- Do not delete license schema or API enum fields in this release.
- Do not use `0` as an unlimited quota while existing checks use `count >= limit`.
- Preserve the existing API Token requirement for `/share/v1/chat/completions`.
- No unrelated RAG or model-provider refactor.

---

### Task 1: Establish the self-hosted backend policy

**Files:**
- Modify: `backend/domain/license.go`
- Modify: `backend/consts/license.go`
- Create: `backend/domain/license_test.go`

**Interfaces:**
- Produces `SelfHostedEditionLimitation() BaseEditionLimitation` with every `Allow*` property true and safe maximum quota values.
- `GetBaseEditionLimitation(context.Context)` returns that policy for absent, malformed, or legacy context data.
- `GetLicenseEdition(echo.Context)` returns the compatibility edition `LicenseEditionEnterprise` without changing role checks.

- [ ] **Step 1: Write failing policy tests**

```go
func TestGetBaseEditionLimitationFallsBackToSelfHosted(t *testing.T) {
    got := GetBaseEditionLimitation(context.Background())
    require.True(t, got.AllowOpenAIBotSettings)
    require.True(t, got.AllowMCPServer)
    require.Greater(t, got.MaxKb, 1)
    require.Greater(t, got.MaxNode, 300)
}

func TestGetBaseEditionLimitationIgnoresMalformedContext(t *testing.T) {
    ctx := context.WithValue(context.Background(), ContextKeyEditionLimitation, []byte("{"))
    require.True(t, GetBaseEditionLimitation(ctx).AllowAdminPerm)
}
```

- [ ] **Step 2: Run the focused test and observe failure**

Run: `go test ./domain -run TestGetBaseEditionLimitation`

Expected: FAIL because the existing fallback is the free-edition policy.

- [ ] **Step 3: Implement the single effective capability policy**

Use `math.MaxInt`, `math.MaxInt64`, and all `Allow*` fields in
`backend/domain/license.go`; return it from `GetBaseEditionLimitation` when
context is absent or cannot unmarshal. In `backend/consts/license.go`, retain
the enum but make `GetLicenseEdition` return `LicenseEditionEnterprise` as the
self-hosted compatibility value.

- [ ] **Step 4: Run focused tests**

Run: `go test ./domain -run TestGetBaseEditionLimitation`

Expected: PASS.

- [ ] **Step 5: Commit**

```text
git add backend/domain/license.go backend/consts/license.go backend/domain/license_test.go
git commit -m "feat: enable unlimited self-hosted capabilities"
```

### Task 2: Remove backend edition-only rejections without removing authorization

**Files:**
- Modify: `backend/usecase/node.go`
- Modify: `backend/usecase/stat.go`
- Modify: `backend/repo/pg/user.go`
- Modify: `backend/repo/pg/auth.go`
- Modify: `backend/pkg/oauth/github.go` only if its branch is confirmed to be an edition gate
- Create or modify focused package tests beside each changed package

**Interfaces:**
- `ValidateNodePermissionsEdit` permits valid permission changes regardless of legacy edition.
- `ValidateStatDay` permits supported statistic ranges regardless of legacy edition.
- User and SSO creation retain duplicate/input checks but do not enforce edition counts.

- [ ] **Step 1: Write failing tests for former edition gates**

Add one test per affected use case with `LicenseEditionFree` that asserts the
same valid request accepted under Enterprise is accepted under Free. Add tests
that invalid permission payloads and invalid statistic values remain rejected.

- [ ] **Step 2: Run the focused package tests and observe the edition gate failures**

Run: `go test ./usecase ./repo/pg -run 'Test(ValidateNodePermissionsEdit|ValidateStatDay|CreateUser)'`

Expected: existing free-edition cases fail before the code changes.

- [ ] **Step 3: Remove only edition comparisons**

Keep validation of request structure, database uniqueness, and all existing
role/KB middleware. Remove the explicit edition comparison in node/stat
validators and quota rejections in user/auth repositories. Do not alter
`ValidateUserRole`, `ValidateKBUserPerm`, group lookups, or node access code.

- [ ] **Step 4: Re-run focused tests**

Run: `go test ./usecase ./repo/pg -run 'Test(ValidateNodePermissionsEdit|ValidateStatDay|CreateUser)'`

Expected: PASS, including invalid input/permission regression cases.

- [ ] **Step 5: Commit**

```text
git add backend/usecase backend/repo/pg backend/pkg/oauth
git commit -m "feat: remove self-hosted edition limits"
```

### Task 3: Guarantee the deployment `admin` account remains an administrator

**Files:**
- Modify: `backend/repo/pg/user.go`
- Modify: `backend/repo/pg/user_test.go` or the repository's existing user test file
- Modify: deployment documentation only if an existing environment-variable guide exists

**Interfaces:**
- With `ADMIN_PASSWORD` configured, `UserUsecase.NewUserUsecase` creates the
  `admin` account when missing and updates the existing `admin` account's
  password and `role=admin` when present.
- No password is logged or returned.

- [ ] **Step 1: Write failing repository tests**

```go
func TestUpsertDefaultUserRestoresAdminRole(t *testing.T) {
    // Seed account "admin" with UserRoleUser, invoke UpsertDefaultUser with
    // UserRoleAdmin, reload it, and assert role is UserRoleAdmin.
}
```

- [ ] **Step 2: Run test and observe failure**

Run: `go test ./repo/pg -run TestUpsertDefaultUserRestoresAdminRole`

Expected: FAIL because the existing update changes only password.

- [ ] **Step 3: Update the existing transaction**

In `UpsertDefaultUser`, update both the bcrypt password and `role` from the
provided default user when the account already exists. Do not broaden any
HTTP handler or add an account-creation endpoint.

- [ ] **Step 4: Verify**

Run: `go test ./repo/pg -run TestUpsertDefaultUserRestoresAdminRole`

Expected: PASS.

- [ ] **Step 5: Commit**

```text
git add backend/repo/pg/user.go backend/repo/pg/*user*_test.go README.md
git commit -m "fix: preserve admin bootstrap role"
```

### Task 4: Make the admin interface edition-neutral

**Files:**
- Modify: `web/admin/src/components/VersionMask/index.tsx`
- Modify: `web/admin/src/hooks/useVersionFeature.ts`
- Modify direct edition-gated files found by `git grep`: `AuthTypeModal.tsx`, `MemberAdd.tsx`, `contribution/index.tsx`, document property/editor files, `Comments.tsx`, `AddRole.tsx`, `CardAI.tsx`, `CardAuth.tsx`, `CardKB.tsx`, `CardMCP.tsx`, `CardRobotApi.tsx`, `CardSecurity.tsx`, `UserGroup/index.tsx`, and `Statistic/index.tsx`
- Add/update frontend tests only where the repository has an existing test pattern

**Interfaces:**
- `VersionMask` returns its children; `VersionCanUse` returns `null`.
- Feature-status hooks report supported for the self-hosted deployment.
- UI controls no longer use `license.edition` to suppress requests, hide token
  fields, or show paid upgrade text.

- [ ] **Step 1: Add or identify rendering tests**

Test one representative formerly Business-only setting (OpenAI API Token) and
one formerly Professional-only setting (AI configuration) with a Free license
response. Assert the control is rendered and the normal authorization/API
error is displayed if the server rejects it for a non-edition reason.

- [ ] **Step 2: Run the test/type check to observe the current masking behavior**

Run: `pnpm --dir web/admin build`

Expected before implementation: code compiles but representative UI tests show
the version mask/hidden content.

- [ ] **Step 3: Implement neutral compatibility components and remove direct branches**

Do not fabricate a paid license in UI state. Keep received license DTOs for
compatibility, but remove only license-edition conditional rendering and
edition-specific request suppression. Preserve ordinary API errors, loading
state, and role-based controls.

- [ ] **Step 4: Verify frontend**

Run: `pnpm --dir web/admin build`

Expected: `tsc -b` and Vite build PASS.

- [ ] **Step 5: Commit**

```text
git add web/admin/src
git commit -m "feat: remove admin edition feature gates"
```

### Task 5: Enable the existing OpenAI-compatible chatbot API safely

**Files:**
- Modify: `backend/handler/share/chat.go`
- Modify: `backend/usecase/chat.go`
- Modify: `backend/repo/pg/auth.go`
- Add focused tests under the existing handler/usecase/repository test conventions

**Interfaces:**
- `POST /share/v1/chat/completions` remains Bearer-token protected per KB.
- The Auth record used to derive RAG groups is fetched with `(kb_id, source_type)`.
- Browser CORS permits `Authorization` and `X-KB-ID`.
- Client responses do not expose raw internal errors.

- [ ] **Step 1: Write failing tests**

Create a two-KB fixture with distinct OpenAI bot Auth groups. Assert a chat
for KB B loads KB B's Auth record, not the first `source_type=openai_api`
record. Add a CORS preflight assertion for both required headers.

- [ ] **Step 2: Run focused tests and observe failure**

Run: `go test ./usecase ./repo/pg ./handler/share -run 'TestOpenAI'`

Expected: KB-scoped Auth test fails against the global source-type lookup.

- [ ] **Step 3: Implement smallest safe fix**

Add/use a repository method accepting `kbID` and `sourceType` in the OpenAI
chat path. Add the two headers to the route's CORS allow list. Replace handler
responses that forward `err.Error()` with stable OpenAI error messages while
logging diagnostic details server-side. Do not remove token validation or
FullControl configuration authorization.

- [ ] **Step 4: Verify focused tests**

Run: `go test ./usecase ./repo/pg ./handler/share -run 'TestOpenAI'`

Expected: PASS.

- [ ] **Step 5: Commit**

```text
git add backend/handler/share/chat.go backend/usecase/chat.go backend/repo/pg/auth.go
git commit -m "fix: scope OpenAI bot access by knowledge base"
```

### Task 6: Full verification and documentation

**Files:**
- Modify: relevant deployment/readme documentation with `ADMIN_PASSWORD` setup
- Modify: generated Swagger only through the repository's documented generator, if source changes require it

- [ ] **Step 1: Format**

Run: `gofmt -w` only on changed Go files.

- [ ] **Step 2: Run backend verification**

Run: `go test ./domain ./usecase ./repo/pg ./handler/share ./handler/v1`

Run: `go vet ./...`

Run: `go build -buildvcs=false ./...`

- [ ] **Step 3: Run frontend verification**

Run: `pnpm --dir web/admin build`

- [ ] **Step 4: Review the diff and commit documentation**

Confirm no secrets, volumes, build output, or unrelated formatting changes are
staged. Document the required deployment setting:

```text
ADMIN_PASSWORD=<operator-chosen-secret>
```

Commit:

```text
git add README.md docs
git commit -m "docs: describe self-hosted administrator bootstrap"
```
