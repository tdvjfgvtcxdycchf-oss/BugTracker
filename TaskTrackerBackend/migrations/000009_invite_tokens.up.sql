CREATE TABLE invite_token (
    id_pk         SERIAL PRIMARY KEY,
    token         VARCHAR(64) NOT NULL UNIQUE,
    org_id_fk     INTEGER REFERENCES organizations(id_pk) ON DELETE CASCADE,
    project_id_fk INTEGER REFERENCES projects(id_pk) ON DELETE CASCADE,
    role          VARCHAR(20) NOT NULL DEFAULT 'member',
    created_by_fk INTEGER NOT NULL REFERENCES "User"(id_pk) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    CONSTRAINT chk_invite_scope CHECK (
        (org_id_fk IS NOT NULL AND project_id_fk IS NULL) OR
        (org_id_fk IS NULL AND project_id_fk IS NOT NULL)
    )
);

CREATE INDEX idx_invite_token ON invite_token(token);
