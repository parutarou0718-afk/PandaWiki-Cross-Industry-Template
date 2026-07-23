-- create table stats_pages for 24-hour retention
CREATE TABLE IF NOT EXISTS stat_pages (
    id BIGSERIAL PRIMARY KEY,
    kb_id TEXT NOT NULL,
    node_id TEXT NOT NULL,
    user_id TEXT,
    session_id TEXT,
    scene INT NOT NULL,
    ip TEXT,
    ua TEXT,
    browser_name TEXT,
    browser_os TEXT,
    referer TEXT,
    referer_host TEXT,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stat_pages_kb_id_node_id ON stat_pages(kb_id, node_id);
