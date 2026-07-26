# Ubuntu Self-Hosted Deployment Design

## Goal

Provide a clean Ubuntu deployment package that starts this fork, including its
unlimited self-hosted policy, with empty persistent volumes and an operator
configured `admin` account.

## Deployment Boundary

The package builds fork-owned `api`, `consumer`, `app`, and `nginx` images.
It retains the upstream service topology for PostgreSQL, Redis, MinIO, NATS,
Qdrant, crawler, raglite, and Caddy. It does not include backup data,
certificates, image archives, or an actual `.env` file.

## Layout

```text
deploy/
  docker-compose.yml
  .env.example
  nginx/
    Dockerfile
    nginx.conf
    server.conf
  scripts/
    build-images.sh
    up.sh
    healthcheck.sh
  README.md
```

`docker-compose.yml` uses relative source build contexts so an Ubuntu operator
can clone the branch and run the package from the repository root. Persistent
data is created under `deploy/data/` and ignored by Git.

## Security and Administration

The operator copies `.env.example` to `.env` and supplies unique secrets for
PostgreSQL, Redis, NATS, MinIO, Qdrant, JWT, and `ADMIN_PASSWORD`.
`ADMIN_PASSWORD` initializes or repairs the normal `admin` account. No
credential is committed, generated into logs, or exposed by the deployment
scripts.

## First Boot

1. Build the web app production output before its container image is built.
2. Build the four fork-owned images.
3. Start Compose with empty volumes.
4. The API container applies its normal migration process before serving.
5. Verify container health and open the configured admin port.

## Verification

The package must pass Docker Compose configuration validation with placeholder
values. Scripts must use `set -eu`, resolve their repository root from their
own location, and never read an operator's `.env` into output. Documentation
must include install, start, status, logs, upgrade, backup, and reset
instructions.
