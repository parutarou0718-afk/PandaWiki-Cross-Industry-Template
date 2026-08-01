CREATE TABLE IF NOT EXISTS knowledge_base_plugin_records (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    plugin_id TEXT NOT NULL,
    record_type TEXT NOT NULL,
    owner_user_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'knowledge_base', 'groups')),
    shared_group_ids BIGINT[] NOT NULL DEFAULT '{}',
    allow_collaborative_edit BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    CONSTRAINT knowledge_base_plugin_records_access_shape CHECK (
        (visibility IN ('private', 'knowledge_base') AND cardinality(shared_group_ids) = 0)
        OR (visibility = 'groups' AND cardinality(shared_group_ids) > 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_plugin_records_visible_lookup
    ON knowledge_base_plugin_records (kb_id, plugin_id, record_type, updated_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_records_owner_lookup
    ON knowledge_base_plugin_records (kb_id, owner_user_id, updated_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_records_shared_groups
    ON knowledge_base_plugin_records USING GIN (shared_group_ids)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_plugin_records_deleted
    ON knowledge_base_plugin_records (kb_id, plugin_id, record_type, deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS knowledge_base_plugin_groups (
    id BIGSERIAL PRIMARY KEY,
    kb_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT knowledge_base_plugin_groups_name_length CHECK (char_length(trim(name)) BETWEEN 1 AND 100)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_groups_name_per_kb
    ON knowledge_base_plugin_groups (kb_id, lower(name));

CREATE TABLE IF NOT EXISTS knowledge_base_plugin_group_members (
    group_id BIGINT NOT NULL REFERENCES knowledge_base_plugin_groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    PRIMARY KEY (group_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_plugin_group_members_user
    ON knowledge_base_plugin_group_members (user_id, group_id);
