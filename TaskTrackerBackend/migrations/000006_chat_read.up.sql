CREATE TABLE chat_read_state (
    thread_id_fk          INTEGER NOT NULL,
    user_id_fk            INTEGER NOT NULL,
    last_read_message_id  INTEGER,
    last_read_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_chat_read_state PRIMARY KEY (thread_id_fk, user_id_fk),
    CONSTRAINT fk_chat_read_state_thread FOREIGN KEY (thread_id_fk) REFERENCES chat_thread(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_read_state_user FOREIGN KEY (user_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_chat_read_state_msg FOREIGN KEY (last_read_message_id) REFERENCES chat_message(id_pk) ON DELETE SET NULL
);

CREATE INDEX idx_chat_read_state_user ON chat_read_state(user_id_fk, thread_id_fk);
