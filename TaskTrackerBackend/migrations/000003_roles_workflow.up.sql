-- Добавляем роль пользователя
ALTER TABLE "User" ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'qa';

-- Таблица комментариев к багу (обязательны при Reject / Can't Reproduce)
CREATE TABLE bug_comment (
    id_pk        SERIAL PRIMARY KEY,
    bug_id_fk    INTEGER NOT NULL,
    user_id_fk   INTEGER NOT NULL,
    body         TEXT    NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_comment_bug  FOREIGN KEY (bug_id_fk)  REFERENCES Bug(id_pk)    ON DELETE CASCADE,
    CONSTRAINT fk_comment_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);

-- Таблица истории изменений (audit log)
CREATE TABLE audit_log (
    id_pk        SERIAL PRIMARY KEY,
    bug_id_fk    INTEGER      NOT NULL,
    user_id_fk   INTEGER      NOT NULL,
    field        VARCHAR(100) NOT NULL,
    old_value    TEXT,
    new_value    TEXT,
    changed_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_audit_bug  FOREIGN KEY (bug_id_fk)  REFERENCES Bug(id_pk)    ON DELETE CASCADE,
    CONSTRAINT fk_audit_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);

-- Таблица связей между багами
CREATE TABLE bug_relation (
    id_pk           SERIAL PRIMARY KEY,
    from_bug_id_fk  INTEGER     NOT NULL,
    to_bug_id_fk    INTEGER     NOT NULL,
    relation_type   VARCHAR(50) NOT NULL, -- 'duplicate' | 'blocks' | 'related'

    CONSTRAINT fk_rel_from FOREIGN KEY (from_bug_id_fk) REFERENCES Bug(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_rel_to   FOREIGN KEY (to_bug_id_fk)   REFERENCES Bug(id_pk) ON DELETE CASCADE,
    CONSTRAINT uq_bug_relation UNIQUE (from_bug_id_fk, to_bug_id_fk, relation_type)
);

-- Таблица тегов
CREATE TABLE bug_tag (
    id_pk       SERIAL PRIMARY KEY,
    bug_id_fk   INTEGER      NOT NULL,
    tag         VARCHAR(100) NOT NULL,

    CONSTRAINT fk_tag_bug FOREIGN KEY (bug_id_fk) REFERENCES Bug(id_pk) ON DELETE CASCADE,
    CONSTRAINT uq_bug_tag UNIQUE (bug_id_fk, tag)
);

-- Шаблоны баг-репортов
CREATE TABLE bug_template (
    id_pk                SERIAL PRIMARY KEY,
    name                 VARCHAR(255) NOT NULL,
    severity             VARCHAR(50),
    priority             VARCHAR(50),
    os                   VARCHAR(200),
    version_product      VARCHAR(50),
    description          TEXT,
    playback_description TEXT,
    expected_result      TEXT,
    actual_result        TEXT,
    created_by_fk        INTEGER NOT NULL,

    CONSTRAINT fk_template_user FOREIGN KEY (created_by_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);
