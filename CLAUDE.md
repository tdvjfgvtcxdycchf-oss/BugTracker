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
│       ├── AdminPage.tsx # Org/project management (create + add members)
│       ├── ChatPage.tsx  # Org/project/DM chat UI
│       ├── ProfilePage.tsx # Self-service account/security page
│       ├── BugsModal.tsx # Bug list for a task
│       └── BugDetailEditor.tsx  # Bug create/edit form with full lifecycle
├── nginx/bugtracker.conf # Nginx: HTTPS + /api/* proxy to backend
├── docker-compose.prod.yml
├── DEPLOY_ORGS_PROJECTS.md # Prod notes for org/project migration + deploy
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
Backend requires `JWT_SECRET` for protected routes (JWT auth).

In production, nginx strips `/api/` prefix and proxies to `http://127.0.0.1:8081/`, so `VITE_API_URL` must be set to `https://bugtracker.sytes.net/api`.

Required `.env` fields (in `TaskTrackerBackend/docker/.env`):
```
POSTGRES_DB=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_URL=postgres://<user>:<pass>@postgres_db:5432/<db>?sslmode=disable
JWT_SECRET=
CORS_ALLOW_ORIGIN=https://bugtracker.sytes.net
```

## Architecture notes

**Request flow:** Browser → Nginx (`/api/*`) → Go backend `:8081` → PostgreSQL

**Backend layers:**
- `sql/postgres.go` — raw pgx queries, DB structs (`User`, `Task`, `Bug`)
- `service/taskTracker.go` — thin wrapper, business logic (e.g. duplicate user check, password comparison)
- `server/handlers.go` — HTTP handlers, all return JSON; errors via `writeJSONError`
- `server/server.go` — single `NewRouter()` wires all routes; CORS is applied in `main.go`

**Frontend:**
Single-page app, no state management library. `localStorage` stores `userId`, `userEmail`, `jwtToken`, and selected `selectedOrgId/selectedProjectId` after login. All API calls go to `API_URL` from `config.ts`.

Bug photos are uploaded as `multipart/form-data` to `POST /bugs/{id}/photo` and served back from `GET /bugs/{id}/photo` — stored as `BYTEA` in the `Bug` table. `GET /bugs/{id}/photo` is JWT-protected, so the frontend fetches the image as a blob with `Authorization` header.

**Access model (orgs/projects):**
- Users belong to organizations (`org_member`) and may have org role `owner|admin|member`.
- Projects belong to an organization and have per-project roles (`project_member`: `pm|dev|qa|viewer`).
- Tasks belong to a project (`Task.project_id_fk`), and the API scopes reads/writes by project membership (or org owner/admin).

**Key API conventions:**
- Protected routes require `Authorization: Bearer <jwt>`.
- Passwords are stored as `bcrypt` hashes. Login keeps backward compatibility with legacy plaintext and auto-upgrades to hash after successful legacy login.
- Password policy for new/changed passwords: min 8 chars, at least one upper-case, one lower-case, one digit.
- Password change rejects reusing current password (`code: password_reused`).
- Auth errors now include optional machine code field in JSON (`{ "error": "...", "code": "..." }`) for login/register/password flows.
- Public auth endpoints have in-memory IP rate limiting:
  - `POST /login` and `POST /users`
  - current limit: 15 requests per minute per IP (best-effort, app-instance local).
- Health endpoint: `GET /healthz` (used by docker healthcheck).
- JWT contains `ver` claim (`jwt_version` from DB). Middleware checks version; incrementing `jwt_version` revokes old sessions.
- Profile/self-service endpoints:
  - `GET /me`
  - `PATCH /me/email` (`new_email`, `current_password`)
  - `PATCH /me/password` (`current_password`, `new_password`) — rotates password and revokes sessions
  - `POST /me/logout-all` — increments `jwt_version` and invalidates all active tokens
- `/tasks`:
  - `GET /tasks?project_id=...` (required) — returns tasks only for that project if user has access.
  - `POST /tasks` requires `project_id` in JSON body.
- `/projects`:
  - `GET /projects?org_id=...` (required).
- Membership management is done by email:
  - `POST /orgs/{id}/members` and `POST /projects/{id}/members`
  - If email user does not exist: returns `404 {"error":"user_not_found"}` (UI suggests user registration).
- Member management endpoints (owner/admin only):
  - `GET /orgs/{id}/members`
  - `PATCH /orgs/{id}/members/{userId}` (change role)
  - `DELETE /orgs/{id}/members/{userId}`
  - `GET /projects/{id}/members`
  - `PATCH /projects/{id}/members/{userId}` (change role)
  - `DELETE /projects/{id}/members/{userId}`

**Chat system:**
- DB tables:
  - `chat_thread` (scope: `org|project|dm`)
  - `chat_message`
  - `chat_read_state` (per-user read cursor)
  - `chat_typing_state` (ephemeral typing indicator state)
- Endpoints:
  - `POST /chat/threads` — ensure/create thread
  - `GET /chat/threads?scope=org&org_id=...`
  - `GET /chat/threads?scope=project&project_id=...`
  - `GET /chat/threads?scope=dm`
  - `GET /chat/threads/{id}/messages?limit=...&before_id=...` (pagination)
  - `POST /chat/threads/{id}/messages`
  - `PATCH /chat/messages/{id}` (edit own message)
  - `DELETE /chat/messages/{id}` (soft-delete own message)
  - `POST /chat/threads/{id}/read`
  - `POST /chat/threads/{id}/typing`
  - `GET /chat/threads/{id}/typing`
- Access:
  - org thread → org members only
  - project thread → project members (or org owner/admin)
  - dm thread → only the two participants

**Migrations:**
Plain SQL files in `migrations/`. The dev docker-compose mounts them into `/docker-entrypoint-initdb.d/` (only runs on first DB init). For an existing prod DB, new migrations must be applied manually:
```bash
docker exec -it <postgres_container> psql -U <user> -d <db> -f /path/to/migration.up.sql
```

**No tests** are present in the project.

## Production hardening notes

- `deploy.sh` now:
  - uses current checked-out branch (not hardcoded `main`)
  - uses `git pull --ff-only` to avoid accidental merge commits in prod deploy
  - forces frontend build with `VITE_API_URL=https://bugtracker.sytes.net/api`
- `docker-compose.prod.yml` has healthchecks for `postgres_db` and `app` (`/healthz`).
- Backend runtime now enforces:
  - non-empty `JWT_SECRET` at startup
  - restrictive CORS origin via `CORS_ALLOW_ORIGIN`
  - HTTP server timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`)
- Nginx config adds:
  - security headers (`X-Content-Type-Options`, `X-Frame-Options`, etc.)
  - `client_max_body_size 15m`
  - proxy timeouts for `/api/`
