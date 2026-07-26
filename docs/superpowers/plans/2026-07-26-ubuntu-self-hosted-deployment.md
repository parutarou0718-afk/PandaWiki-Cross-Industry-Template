# Ubuntu Self-Hosted Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a safe, clean Ubuntu Docker Compose deployment package for this fork.

**Architecture:** The package composes official data services with four images built from this checkout. Runtime secrets are supplied only in `deploy/.env`; all persisted data is in ignored `deploy/data` directories.

**Tech Stack:** Docker Compose v2, Bash, Nginx, PandaWiki Dockerfiles.

## Global Constraints

- Never commit `.env`, data volumes, certificates, databases, tokens, or image archives.
- Never replace normal authentication or role checks.
- Deploy from a clean clone on Ubuntu; do not rely on Windows paths.

### Task 1: Create the Compose package

**Files:** Create `deploy/docker-compose.yml`, `deploy/.env.example`,
`deploy/.gitignore`, and the Nginx image/configuration files.

- [ ] Define all upstream data services with named relative volume paths under
  `deploy/data`.
- [ ] Define `api`, `consumer`, `app`, and `nginx` build contexts that point at
  the fork source and preserve the original container hostnames.
- [ ] Add an example environment file with placeholders only.
- [ ] Validate using `docker compose --env-file .env.example config`.

### Task 2: Add Ubuntu operational scripts and instructions

**Files:** Create `deploy/scripts/build-images.sh`, `deploy/scripts/up.sh`,
`deploy/scripts/healthcheck.sh`, and `deploy/README.md`.

- [ ] Build the app artifacts then all Compose images from the repository root.
- [ ] Start the stack only after `.env` exists.
- [ ] Check all required containers and expose useful failure logs.
- [ ] Document fresh deployment, login, backup, update, and destructive reset
  procedures.

### Task 3: Verify and commit

- [ ] Validate Compose with placeholder values.
- [ ] Shell syntax-check every script when Bash is available.
- [ ] Confirm `deploy/.gitignore` excludes data, `.env`, and TLS material.
- [ ] Commit only deployment configuration, scripts, and documentation.
