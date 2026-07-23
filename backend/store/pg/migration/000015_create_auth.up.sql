-- create table auths
CREATE TABLE IF NOT EXISTS auths (
    id SERIAL PRIMARY KEY,
    user_info JSONB NULL,
    union_id text NOT NULL,
    ip text NOT NULL,
    kb_id text NOT NULL,
    source_type text NOT NULL,
    last_login_time timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
-- create table auth_configs
CREATE TABLE IF NOT EXISTS auth_configs (
    id SERIAL PRIMARY KEY,
    kb_id text NOT NULL,
    auth_setting JSONB NULL,
    source_type text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
