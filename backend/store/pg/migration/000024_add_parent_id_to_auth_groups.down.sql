-- Remove parent_id column
ALTER TABLE auth_groups DROP COLUMN IF EXISTS parent_id;

-- Remove position column from auth_groups table
ALTER TABLE auth_groups DROP COLUMN IF EXISTS position;