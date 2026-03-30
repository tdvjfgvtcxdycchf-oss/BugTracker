-- Organizations / Projects / Membership + tasks scoped by project

CREATE TABLE organizations (
    id_pk       SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE org_member (
    id_pk      SERIAL PRIMARY KEY,
    org_id_fk  INTEGER NOT NULL,
    user_id_fk INTEGER NOT NULL,
    role       VARCHAR(20) NOT NULL, -- owner | admin | member
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_org_member_org  FOREIGN KEY (org_id_fk)  REFERENCES organizations(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_org_member_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk)        ON DELETE CASCADE,
    CONSTRAINT uq_org_member UNIQUE (org_id_fk, user_id_fk)
);

CREATE TABLE projects (
    id_pk      SERIAL PRIMARY KEY,
    org_id_fk  INTEGER NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_projects_org FOREIGN KEY (org_id_fk) REFERENCES organizations(id_pk) ON DELETE CASCADE
);

CREATE TABLE project_member (
    id_pk       SERIAL PRIMARY KEY,
    project_id_fk INTEGER NOT NULL,
    user_id_fk    INTEGER NOT NULL,
    role          VARCHAR(20) NOT NULL, -- pm | dev | qa | viewer
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_project_member_project FOREIGN KEY (project_id_fk) REFERENCES projects(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_project_member_user    FOREIGN KEY (user_id_fk)    REFERENCES "User"(id_pk)   ON DELETE CASCADE,
    CONSTRAINT uq_project_member UNIQUE (project_id_fk, user_id_fk)
);

ALTER TABLE Task
    ADD COLUMN project_id_fk INTEGER;

-- Create default org/project and backfill existing tasks/users.
WITH
    created_org AS (
        INSERT INTO organizations (name) VALUES ('Default organization')
        RETURNING id_pk
    ),
    created_project AS (
        INSERT INTO projects (org_id_fk, name)
        SELECT id_pk, 'Default project' FROM created_org
        RETURNING id_pk, org_id_fk
    ),
    first_owner AS (
        SELECT id_pk AS user_id
        FROM "User"
        ORDER BY id_pk
        LIMIT 1
    )
-- add org members
INSERT INTO org_member (org_id_fk, user_id_fk, role)
SELECT
    (SELECT org_id_fk FROM created_project),
    u.id_pk,
    CASE WHEN u.id_pk = (SELECT user_id FROM first_owner) THEN 'owner' ELSE 'member' END
FROM "User" u;

-- add project members (everyone can at least view)
-- Note: CTE from the previous statement is out of scope here;
--       use a subquery against the projects table directly.
INSERT INTO project_member (project_id_fk, user_id_fk, role)
SELECT (SELECT id_pk FROM projects ORDER BY id_pk LIMIT 1), u.id_pk, 'viewer'
FROM "User" u;

-- backfill tasks to default project
UPDATE Task
SET project_id_fk = (SELECT id_pk FROM projects ORDER BY id_pk LIMIT 1)
WHERE project_id_fk IS NULL;

ALTER TABLE Task
    ALTER COLUMN project_id_fk SET NOT NULL;

ALTER TABLE Task
    ADD CONSTRAINT fk_task_project FOREIGN KEY (project_id_fk) REFERENCES projects(id_pk) ON DELETE CASCADE;

CREATE INDEX idx_task_project_id_fk ON Task(project_id_fk);
CREATE INDEX idx_project_member_project_user ON project_member(project_id_fk, user_id_fk);
CREATE INDEX idx_org_member_org_user ON org_member(org_id_fk, user_id_fk);
