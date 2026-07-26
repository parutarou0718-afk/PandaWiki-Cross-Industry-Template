#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

cd "$DEPLOY_DIR"
docker compose ps

for service in postgres nats qdrant raglite api consumer nginx app; do
  container="panda-wiki-$service"
  state=$(docker inspect --format '{{.State.Status}}' "$container" 2>/dev/null || true)
  if [ "$state" != "running" ]; then
    echo "$container is not running (state: ${state:-missing})" >&2
    docker compose logs --tail=80 "$service" >&2 || true
    exit 1
  fi
done

echo "All PandaWiki containers are running."
