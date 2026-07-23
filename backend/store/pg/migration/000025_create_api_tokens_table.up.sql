CREATE TABLE IF NOT EXISTS api_tokens (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL,
    permission TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(token)
);