CREATE TABLE chat_thread (
    id_pk         SERIAL PRIMARY KEY,
    scope         VARCHAR(20) NOT NULL, -- org | project | dm
    org_id_fk     INTEGER,
    project_id_fk INTEGER,
    dm_user_a_fk  INTEGER,
    dm_user_b_fk  INTEGER,
    created_by_fk INTEGER NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_chat_thread_org FOREIGN KEY (org_id_fk) REFERENCES organizations(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_thread_project FOREIGN KEY (project_id_fk) REFERENCES projects(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_thread_user_a FOREIGN KEY (dm_user_a_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_thread_user_b FOREIGN KEY (dm_user_b_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_thread_creator FOREIGN KEY (created_by_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_chat_thread_org ON chat_thread(scope, org_id_fk) WHERE scope = 'org';
CREATE UNIQUE INDEX uq_chat_thread_project ON chat_thread(scope, project_id_fk) WHERE scope = 'project';
CREATE UNIQUE INDEX uq_chat_thread_dm_pair ON chat_thread(scope, dm_user_a_fk, dm_user_b_fk) WHERE scope = 'dm';

CREATE TABLE chat_message (
    id_pk         SERIAL PRIMARY KEY,
    thread_id_fk  INTEGER NOT NULL,
    user_id_fk    INTEGER NOT NULL,
    body          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_chat_message_thread FOREIGN KEY (thread_id_fk) REFERENCES chat_thread(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_message_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);

CREATE INDEX idx_chat_message_thread_created ON chat_message(thread_id_fk, created_at DESC);
