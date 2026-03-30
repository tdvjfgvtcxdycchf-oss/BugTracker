ALTER TABLE chat_message
    ADD COLUMN edited_at TIMESTAMPTZ,
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE TABLE chat_typing_state (
    thread_id_fk INTEGER NOT NULL,
    user_id_fk   INTEGER NOT NULL,
    is_typing    BOOLEAN NOT NULL DEFAULT false,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (thread_id_fk, user_id_fk),
    CONSTRAINT fk_chat_typing_thread FOREIGN KEY (thread_id_fk) REFERENCES chat_thread(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_typing_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);

CREATE INDEX idx_chat_typing_updated ON chat_typing_state(thread_id_fk, updated_at DESC);
