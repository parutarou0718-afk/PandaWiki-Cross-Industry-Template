CREATE TABLE IF NOT EXISTS document_feedbacks (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NULL,
    kb_id TEXT NOT NULL,
    node_id TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '', 
    correction_suggestion TEXT NOT NULL DEFAULT '',
    info JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
