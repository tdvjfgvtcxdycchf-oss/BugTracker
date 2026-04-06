# BugTracker

Go backend + React/TS/Vite/Tailwind frontend.

## Structure
```
TaskTrackerBackend/
  cmd/main.go, internal/{server,service,sql}/, migrations/, docker/, makefile
TaskTrackerFrontend/src/
  config.ts, MyApp.tsx, AuthPage, MainPage, AdminPage, ChatPage, ProfilePage, BugsModal, BugDetailEditor
nginx/bugtracker.conf, docker-compose.prod.yml, deploy.sh
```

## Commands
Backend (`TaskTrackerBackend/`): `make runapp|docker-build|docker-up|docker-down|docker-logs`
Frontend (`TaskTrackerFrontend/`): `npm run dev|build|preview`
Deploy (server, `BugTracker/`): `bash deploy.sh`

## Env
- Backend: `POSTGRES_URL`, `JWT_SECRET`, `CORS_ALLOW_ORIGIN` (docker/.env)
- Frontend: `VITE_API_URL` (prod: `https://bugtracker.sytes.net/api`)
- Nginx proxies `/api/*` → `:8081`, strips prefix

## Architecture
**Flow:** Browser → Nginx → Go `:8081` → PostgreSQL

**Backend layers:** `sql/postgres.go` (pgx) → `service/taskTracker.go` (business logic) → `server/handlers.go` (JSON API) → `server/server.go` (routes)

**Frontend:** SPA, no state lib. `localStorage`: `userId`, `userEmail`, `jwtToken`, `selectedOrgId`, `selectedProjectId`.

**Auth:** JWT with `ver` claim (version-revokable). bcrypt passwords. Rate limit: 15 req/min/IP on `/login`, `POST /users`.

**Access model:** orgs → org_member (owner|admin|member) → projects → project_member (pm|dev|qa|viewer)

**Bug photos:** `multipart/form-data` → stored as BYTEA. `GET /bugs/{id}/photo` is JWT-protected (fetch as blob).

## Key API
- `GET /tasks?project_id=` · `POST /tasks` (requires `project_id`)
- `GET /projects?org_id=`
- `GET/PATCH /orgs/{id}/members/{userId}` · `DELETE` (owner/admin only)
- `GET/PATCH /projects/{id}/members/{userId}` · `DELETE`
- `POST /orgs/{id}/members` / `POST /projects/{id}/members` by email; 404 if user absent
- `GET /me` · `PATCH /me/email` · `PATCH /me/password` · `POST /me/logout-all`
- Auth error JSON includes optional `code` field

## Chat
Tables: `chat_thread` (scope: org|project|dm), `chat_message`, `chat_read_state`, `chat_typing_state`
Endpoints: `/chat/threads` CRUD, `/chat/threads/{id}/messages` (pagination: `limit`, `before_id`), `/chat/messages/{id}` edit/delete (soft), `/chat/threads/{id}/read|typing`
Access: org→org members, project→project members+org admin, dm→2 participants only

## Migrations
Plain SQL in `migrations/`. Dev: mounted into postgres init. Prod: apply manually via `psql`.

## Prod notes
- `deploy.sh`: `git pull --ff-only`, builds frontend, docker up, nginx reload
- Healthcheck: `GET /healthz`
- HTTP timeouts set. Nginx: security headers, `client_max_body_size 15m`
