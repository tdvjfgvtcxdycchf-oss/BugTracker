#!/bin/bash
set -e

DOMAIN="bugtracker.sytes.net"
APP_DIR="/home/zolbrain/BugTracker"
WWW_DIR="/var/www/bugtracker"
API_URL="https://bugtracker.sytes.net/api"

echo "=== Checking required tools ==="
command -v git >/dev/null
command -v npm >/dev/null
command -v docker >/dev/null
command -v nginx >/dev/null

echo "=== Pulling latest code ==="
cd "$APP_DIR"
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [ -z "$CURRENT_BRANCH" ]; then
  CURRENT_BRANCH="master"
fi
git fetch origin "$CURRENT_BRANCH"
git pull --ff-only origin "$CURRENT_BRANCH"

# Проверяем наличие .env для бэкенда
ENV_FILE="$APP_DIR/TaskTrackerBackend/docker/.env"
if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: $ENV_FILE не найден!"
  echo "Создайте его вручную:"
  echo "  POSTGRES_DB=postgres"
  echo "  POSTGRES_USER=postgres"
  echo "  POSTGRES_PASSWORD=<ваш_пароль>"
  echo "  POSTGRES_URL=postgres://postgres:<ваш_пароль>@postgres_db:5432/postgres?sslmode=disable"
  exit 1
fi

echo "=== Building frontend ==="
cd "$APP_DIR/TaskTrackerFrontend"
npm ci
VITE_API_URL="$API_URL" npm run build

echo "=== Deploying frontend static files ==="
rm -rf "$WWW_DIR"
mkdir -p "$WWW_DIR"
cp -r dist/* "$WWW_DIR/"

echo "=== Restarting backend ==="
cd "$APP_DIR"
docker compose -f docker-compose.prod.yml up -d --build

echo "=== Reloading nginx ==="
nginx -t && systemctl reload nginx

echo "=== Done! https://$DOMAIN ==="
