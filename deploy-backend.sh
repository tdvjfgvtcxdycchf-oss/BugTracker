#!/bin/bash
# Запускать на БЭКЕНД-сервере
set -euo pipefail

APP_DIR="/home/zolbrain/BugTracker"
COMPOSE_FILE="$APP_DIR/docker-compose.prod.yml"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

log "=== [1/3] Pulling latest code ==="
cd "$APP_DIR"
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
log "Branch: $CURRENT_BRANCH"
git fetch origin "$CURRENT_BRANCH"
git pull --ff-only origin "$CURRENT_BRANCH"

log "=== [2/3] Rebuilding backend image ==="
docker compose -f "$COMPOSE_FILE" build app

log "=== [3/3] Restarting containers ==="
docker compose -f "$COMPOSE_FILE" up -d

log "Waiting for backend to become healthy..."
TRIES=0
until curl -sf "http://127.0.0.1:9191/healthz" >/dev/null 2>&1; do
  TRIES=$((TRIES + 1))
  if [ "$TRIES" -ge 30 ]; then
    log "Last logs:"
    docker compose -f "$COMPOSE_FILE" logs --tail=50 app
    fail "Backend did not become healthy after 60s"
  fi
  sleep 2
done

log ""
log "✓ Backend deployed and healthy"
