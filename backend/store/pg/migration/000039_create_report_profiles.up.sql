CREATE TABLE report_profiles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    edition_ids JSONB NOT NULL,
    system_prompt TEXT NOT NULL,
    input_fields JSONB NOT NULL,
    sections JSONB NOT NULL,
    citation_required BOOLEAN NOT NULL DEFAULT true,
    enabled BOOLEAN NOT NULL DEFAULT true,
    version TEXT NOT NULL,
    is_builtin BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_report_profiles_visible ON report_profiles (enabled) WHERE deleted_at IS NULL;

CREATE TABLE reports (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    profile_id TEXT NOT NULL REFERENCES report_profiles(id),
    profile_version TEXT NOT NULL,
    profile_snapshot JSONB NOT NULL,
    created_by BIGINT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','running','completed','failed')),
    title TEXT NOT NULL DEFAULT '',
    input_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    content TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_reports_owner ON reports (created_by, created_at DESC);

CREATE TABLE report_citations (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    citation_index INTEGER NOT NULL,
    node_id TEXT NOT NULL,
    document_name TEXT NOT NULL,
    locator TEXT NOT NULL DEFAULT '',
    excerpt TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(report_id, citation_index)
);
