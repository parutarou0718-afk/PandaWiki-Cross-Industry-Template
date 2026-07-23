ALTER TABLE auth_groups ADD COLUMN IF NOT EXISTS sync_id text NOT NULL DEFAULT '';
ALTER TABLE auth_groups ADD COLUMN IF NOT EXISTS sync_parent_id text NOT NULL DEFAULT '';
ALTER TABLE auth_groups ADD COLUMN IF NOT EXISTS source_type text NOT NULL DEFAULT '';
ALTER TABLE auth_groups DROP CONSTRAINT IF EXISTS auth_groups_name_key;

