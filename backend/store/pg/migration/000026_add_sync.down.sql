ALTER TABLE auth_groups DROP COLUMN IF EXISTS sync_id;
ALTER TABLE auth_groups DROP COLUMN IF EXISTS sync_parent_id;
ALTER TABLE auth_groups DROP COLUMN IF EXISTS source_type;
