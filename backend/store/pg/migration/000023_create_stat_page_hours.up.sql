CREATE TABLE IF NOT EXISTS stat_page_hours (
    id BIGSERIAL PRIMARY KEY,
    kb_id TEXT NOT NULL,
    hour timestamptz NOT NULL,
    ip_count BIGINT NOT NULL DEFAULT 0,
    session_count BIGINT NOT NULL DEFAULT 0,
    page_visit_count BIGINT NOT NULL DEFAULT 0,
    conversation_count BIGINT NOT NULL DEFAULT 0,
    geo_count JSONB NULL,
    conversation_distribution JSONB NULL,
    hot_referer_host JSONB NULL,
    hot_page JSONB NULL,
    hot_os JSONB NULL,
    hot_browser JSONB NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE(kb_id, hour)
);

CREATE INDEX IF NOT EXISTS idx_stat_page_hours_hour ON stat_page_hours (hour);
