ALTER TABLE auth_groups ADD COLUMN IF NOT EXISTS parent_id INTEGER DEFAULT NULL;

ALTER TABLE auth_groups ADD COLUMN IF NOT EXISTS position FLOAT8 DEFAULT 0;

-- Update existing records with default positions (1000, 2000, 3000, etc.)
UPDATE auth_groups SET position = (id * 1000)::FLOAT8;