## Deploy notes: Organizations / Projects / Access control

### What changed

- Added orgs/projects tables and membership:
  - `organizations`, `org_member`
  - `projects`, `project_member`
- Added chat subsystem:
  - `chat_thread`
  - `chat_message`
- Tasks are now scoped by project:
  - `Task.project_id_fk` added and backfilled
  - API now requires `project_id` query/body
- All `/bugs/*` routes are now protected by project membership.
- `GET /bugs/{id}/photo` now requires JWT (no more public image URLs).

### Required env

- Backend requires `JWT_SECRET` in `TaskTrackerBackend/docker/.env` (or environment).
- Backend also supports `CORS_ALLOW_ORIGIN` (recommended set to your prod frontend origin).
- Frontend should be built with `VITE_API_URL=https://<domain>/api`.

### Production DB migration (existing database)

If your prod database already exists, **apply migrations manually** inside the postgres container.

1) Find container name:

```bash
docker compose -f docker-compose.prod.yml ps
```

2) Apply migrations in order (skip those already applied):

```bash
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000002_add_bug_photo.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000003_roles_workflow.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000004_orgs_projects.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000005_chat.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000006_chat_read.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000007_chat_edit_typing.up.sql
docker exec -i bugtracker-postgres_db-1 psql -U postgres -d postgres < TaskTrackerBackend/migrations/000008_user_jwt_version.up.sql
```

### Frontend deploy

Because the frontend is static, make sure it is built with the correct `VITE_API_URL`.

```bash
cd TaskTrackerFrontend
echo "VITE_API_URL=https://bugtracker.sytes.net/api" > .env
npm ci
npm run build

sudo rm -rf /var/www/bugtracker
sudo mkdir -p /var/www/bugtracker
sudo cp -r dist/* /var/www/bugtracker/
sudo nginx -t && sudo systemctl reload nginx
```

### Backend deploy

```bash
docker compose -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.prod.yml logs -n 80 app
```

### Smoke checks

- `GET https://<domain>/` returns 200
- `POST https://<domain>/api/login` returns `{ token }`
- `GET https://<domain>/api/orgs` returns a list of orgs (JWT required)

