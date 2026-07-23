-- add updated_at to node_releases
ALTER TABLE node_releases ADD COLUMN updated_at timestamptz NULL;

-- update existing node_releases
UPDATE node_releases SET updated_at = created_at;
