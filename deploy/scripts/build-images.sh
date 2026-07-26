#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
REPO_DIR=$(CDPATH= cd -- "$DEPLOY_DIR/.." && pwd)

cd "$REPO_DIR/web"
pnpm install --frozen-lockfile
pnpm --filter panda-wiki-admin build
pnpm --filter panda-wiki-app build

cd "$DEPLOY_DIR"
docker compose build --pull api consumer app nginx
