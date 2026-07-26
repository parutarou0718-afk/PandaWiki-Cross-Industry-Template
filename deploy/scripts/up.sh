#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

if [ ! -f "$DEPLOY_DIR/.env" ]; then
  echo "Missing $DEPLOY_DIR/.env. Copy .env.example and set unique secrets first." >&2
  exit 1
fi

cd "$DEPLOY_DIR"
docker compose up -d --remove-orphans
docker compose ps
