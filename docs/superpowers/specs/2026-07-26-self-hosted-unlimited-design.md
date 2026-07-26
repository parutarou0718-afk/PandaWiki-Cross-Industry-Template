# Self-Hosted Unlimited Edition Design

## Goal

Turn this PandaWiki fork into a single self-hosted edition: commercial
edition gates, upgrade prompts, and edition-based quotas are removed, while
normal authentication, administrator roles, knowledge-base membership,
user-group access, node permissions, token validation, and input validation
continue to control access.

## Non-goals

- Do not bypass login, system roles, knowledge-base permissions, group-based
  node permissions, or document permissions.
- Do not add a web-accessible superuser endpoint, a default root account, or
  a fixed password.
- Do not delete existing license tables, response fields, or enum values in
  V1. They remain for database and client compatibility but no longer decide
  available features.
- Do not change unrelated RAG retrieval, model-provider, storage, or billing
  infrastructure.

## Architecture

### 1. One effective capability policy

`domain.GetBaseEditionLimitation` becomes the single compatibility boundary
for edition limits. Its effective self-hosted policy enables every capability
in `BaseEditionLimitation` and uses safe unbounded values for knowledge bases,
nodes, administrators, and SSO users. Existing callers keep their ordinary
validation behavior but no longer reject an operation because of the
installation edition.

The implementation must use values compatible with existing `count >= limit`
checks. It must not use `0` as an unlimited value unless every affected check
is changed to understand that sentinel. The preferred V1 implementation uses
the largest safe integer for each existing field type.

Malformed or absent `edition_limitation` context values must resolve to this
self-hosted policy rather than the former free-edition policy.

### 2. Remove direct edition gates

Some backend paths compare `LicenseEdition` directly instead of consulting
`BaseEditionLimitation`. These paths must be changed to allow the feature
without weakening normal authorization:

- node permission editing;
- statistics time-range validation;
- administrator and SSO-user quota checks;
- enterprise-auth and OAuth branches where the condition is only an edition
  gate;
- any middleware route registered with `ValidateLicenseEdition`.

Each direct comparison must be audited before removal. A comparison that
implements an identity, ownership, or data-permission rule is not an edition
gate and remains unchanged.

### 3. Admin frontend is edition-neutral

The admin UI must no longer mask controls, hide settings, disable actions, or
show upgrade copy because a deployment reports Free, Professional, Business,
or Enterprise. `VersionMask` and `VersionCanUse` become neutral compatibility
components; direct `license.edition` feature branches are removed or replaced
with ordinary availability/error handling.

The version enum and API response field remain available so older clients and
stored data still parse. Product marketing/version-comparison presentation is
removed from normal administration flows rather than falsely displaying the
installation as a paid upstream tier.

### 4. Deployment administrator bootstrap

The repository already supplies a deployment-only administrator bootstrap:
when `ADMIN_PASSWORD` is configured, startup creates or updates the `admin`
account. V1 keeps this existing mechanism rather than adding another command
surface. It is corrected so an existing `admin` account is always restored to
the administrator role when the deployment operator explicitly configures
`ADMIN_PASSWORD`.

It must:

- require an operator-supplied deployment secret;
- never create a default root account;
- never contain a compiled-in credential;
- avoid emitting passwords in output or logs;
- make its change through the existing user/role persistence path; and
- have no HTTP equivalent or hidden UI control.

### 5. OpenAI-compatible chatbot API remains secured

Removing its edition gate must not remove its existing knowledge-base
FullControl requirement for configuration or its Bearer-token requirement for
calls. The implementation also fixes the discovered multi-knowledge-base
scope defect before enabling it broadly: the OpenAI bot Auth record used to
derive RAG groups must be loaded by both `kb_id` and `source_type`, not merely
the first record matching `source_type`.

The API's browser CORS headers must include `Authorization` and `X-KB-ID` if
browser clients are supported. It must not return internal error strings to
callers. Token hashing/rotation, rate limiting, standard OpenAI `[DONE]`, and
citations are documented follow-up hardening work, not prerequisites for the
edition-gate removal unless the existing project supplies them already.

## Preserved authorization model

```text
Server-local bootstrap command
  -> system administrator account
  -> normal administrator login
  -> knowledge-base membership and role checks
  -> user groups
  -> node visibility / visitability / edit permissions
  -> RAG group filtering
```

At no point does the self-hosted policy grant one ordinary user access to
another user's knowledge base or protected node.

## Error handling and compatibility

- Existing license fields may still be returned by APIs, but are no longer an
  authorization decision point for self-hosted features.
- Existing databases require no destructive migration for the edition change.
- Startup fails clearly if the configured bootstrap password cannot be hashed
  or the `admin` user cannot be created/updated.
- Enabling a feature does not silently create or overwrite user accounts,
  knowledge bases, tokens, or permissions.

## Verification

Automated tests must cover:

1. every `BaseEditionLimitation` feature and quota is effective under missing,
   malformed, and formerly free-edition context data;
2. normal role, KB membership, and node/group permission denials still occur;
3. formerly edition-gated API/settings flows succeed for an authorized
   administrator and fail for an unauthorized user;
4. the OpenAI bot uses the current KB's Auth/group scope;
5. the `ADMIN_PASSWORD` bootstrap creates/elevates the `admin` account and
   does not log its password;
6. the admin build contains no functional VersionMask gate or paid-upgrade
   message in affected controls.

Run focused Go tests, backend build/vet, admin TypeScript/type checks and
production build. Existing unrelated test failures, if any, are recorded
separately and never masked.

## Delivery sequence

1. Establish backend self-hosted policy and tests.
2. Remove direct backend edition gates and test preserved authorization.
3. Correct and test the deployment administrator bootstrap.
4. Remove admin UI gates and paid-version messaging; build the admin app.
5. Enable and harden the existing OpenAI-compatible chatbot API scope/CORS
   behavior; perform end-to-end verification.
