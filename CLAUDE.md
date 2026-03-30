# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project structure

```
BugTracker/
├── TaskTrackerBackend/   # Go HTTP server
│   ├── cmd/main.go       # Entrypoint: logging, DB connection, CORS, graceful shutdown
│   ├── internal/
│   │   ├── server/       # HTTP layer: router setup (server.go) and handlers (handlers.go)
│   │   ├── service/      # Business logic (taskTracker.go)
│   │   └── sql/          # DB structs and raw pgx queries (postgres.go)
│   ├── migrations/       # Plain SQL files (000001_init.up.sql, etc.)
│   ├── docker/           # Dev docker-compose + Dockerfile + .env
│   └── makefile
├── TaskTrackerFrontend/  # React + TypeScript + Vite + Tailwind CSS v4
│   └── src/
│       ├── config.ts     # Exports API_URL from VITE_API_URL env var
│       ├── MyApp.tsx     # Router root
│       ├── AuthPage.tsx  # Login/register
│       ├── MainPage.tsx  # Task list
│       ├── BugsModal.tsx # Bug list for a task
│       └── BugDetailEditor.tsx  # Bug create/edit form with full lifecycle
├── nginx/bugtracker.conf # Nginx: HTTPS + /api/* proxy to backend
├── docker-compose.prod.yml
└── deploy.sh             # Full deploy script (git pull → build frontend → docker up → nginx reload)
```

## Backend commands

From `TaskTrackerBackend/`:

```bash
make runapp          # go run cmd/main.go (requires .env in working dir)
make docker-build    # build dev image
make docker-up       # start dev stack (postgres + app) in background
make docker-down     # stop and remove dev containers
make docker-logs     # follow logs
```

Prod deploy (run on server from `BugTracker/`):
```bash
bash deploy.sh
```

## Frontend commands

From `TaskTrackerFrontend/`:

```bash
npm run dev      # dev server with HMR
npm run build    # tsc + vite build → dist/
npm run preview  # preview built dist/
```

## Environment

Backend reads `POSTGRES_URL` from environment (or `docker/.env`).
Frontend reads `VITE_API_URL` (set in `.env` or passed at build time).

In production, nginx strips `/api/` prefix and proxies to `http://127.0.0.1:8081/`, so `VITE_API_URL` must be set to `https://bugtracker.sytes.net/api`.

Required `.env` fields (in `TaskTrackerBackend/docker/.env`):
```
POSTGRES_DB=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_URL=postgres://<user>:<pass>@postgres_db:5432/<db>?sslmode=disable
```

## Architecture notes

**Request flow:** Browser → Nginx (`/api/*`) → Go backend `:8081` → PostgreSQL

**Backend layers:**
- `sql/postgres.go` — raw pgx queries, DB structs (`User`, `Task`, `Bug`)
- `service/taskTracker.go` — thin wrapper, business logic (e.g. duplicate user check, password comparison)
- `server/handlers.go` — HTTP handlers, all return JSON; errors via `writeJSONError`
- `server/server.go` — single `NewRouter()` wires all routes; CORS is applied in `main.go`

**Frontend:**
Single-page app, no state management library. `localStorage` stores `userId` and `userEmail` after login. All API calls go to `API_URL` from `config.ts`. Bug photos are uploaded as `multipart/form-data` to `POST /bugs/{id}/photo` and served back from `GET /bugs/{id}/photo` — stored as `BYTEA` in the `Bug` table.

**Migrations:**
Plain SQL files in `migrations/`. The dev docker-compose mounts them into `/docker-entrypoint-initdb.d/` (only runs on first DB init). For an existing prod DB, new migrations must be applied manually:
```bash
docker exec -it <postgres_container> psql -U <user> -d <db> -f /path/to/migration.up.sql
```

**No tests** are present in the project.
