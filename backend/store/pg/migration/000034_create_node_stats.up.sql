CREATE TABLE IF NOT EXISTS node_stats (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL UNIQUE,
    pv BIGINT NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

