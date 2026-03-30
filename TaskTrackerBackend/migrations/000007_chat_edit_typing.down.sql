DROP TABLE IF EXISTS chat_typing_state;

ALTER TABLE chat_message
    DROP COLUMN IF EXISTS edited_at,
    DROP COLUMN IF EXISTS deleted_at;
