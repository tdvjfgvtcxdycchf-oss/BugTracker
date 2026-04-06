#!/bin/bash
set -e

DOMAIN="bugtracker.sytes.net"
APP_DIR="/home/zolbrain/BugTracker"
WWW_DIR="/var/www/bugtracker"

# Адрес бэкенда (сервер друга) — поменяй на реальный URL когда будет готово
API_URL="http://176.108.248.47:9191"

echo "=== Checking required tools ==="
command -v git >/dev/null
command -v npm >/dev/null
command -v nginx >/dev/null

echo "=== Pulling latest code ==="
cd "$APP_DIR"
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [ -z "$CURRENT_BRANCH" ]; then
  CURRENT_BRANCH="master"
fi
git fetch origin "$CURRENT_BRANCH"
git pull --ff-only origin "$CURRENT_BRANCH"

echo "=== Stopping and removing backend containers ==="
if docker compose -f docker-compose.prod.yml ps -q 2>/dev/null | grep -q .; then
  docker compose -f docker-compose.prod.yml down
  echo "Backend stopped."
else
  echo "No running containers, skipping."
fi

echo "=== Building frontend ==="
cd "$APP_DIR/TaskTrackerFrontend"
npm ci
VITE_API_URL="$API_URL" npm run build

echo "=== Deploying frontend static files ==="
rm -rf "$WWW_DIR"
mkdir -p "$WWW_DIR"
cp -r dist/* "$WWW_DIR/"

echo "=== Reloading nginx ==="
nginx -t && systemctl reload nginx

echo "=== Done! https://$DOMAIN ==="
