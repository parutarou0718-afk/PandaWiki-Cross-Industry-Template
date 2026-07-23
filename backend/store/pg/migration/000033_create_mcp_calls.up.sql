CREATE TABLE IF NOT EXISTS mcp_calls (
    id SERIAL PRIMARY KEY,
    mcp_session_id TEXT NOT NULL,
    kb_id TEXT NOT NULL,
    remote_ip TEXT,
    initialize_req JSONB,
    initialize_resp JSONB,
    tool_call_req JSONB,
    tool_call_resp TEXT,
    created_at timestamptz NOT NULL DEFAULT NOW()
);
