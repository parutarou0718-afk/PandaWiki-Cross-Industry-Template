-- Create auth_groups table
CREATE TABLE IF NOT EXISTS auth_groups (
    id SERIAL PRIMARY KEY,
    kb_id TEXT NOT NULL,
    name VARCHAR(100) NOT NULL UNIQUE,
    auth_ids INTEGER[] DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create node_auth_groups table
CREATE TABLE IF NOT EXISTS node_auth_groups (
    id SERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    auth_group_id INTEGER NOT NULL,
    perm TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(node_id, auth_group_id, perm)
);


ALTER TABLE nodes ADD COLUMN permissions jsonb default '{}';
UPDATE nodes set permissions='{"answerable":"open","visitable":"open","visible":"open"}'::jsonb;


-- update nodes table
ALTER TABLE "public"."nodes" ADD COLUMN "creator_id" TEXT NOT NULL DEFAULT '';
ALTER TABLE "public"."nodes" ADD COLUMN "editor_id" TEXT NOT NULL DEFAULT '';

UPDATE nodes SET creator_id = u.id, editor_id = u.id FROM "users" u WHERE u.account = 'admin';

UPDATE nodes set "permissions"='{"answerable":"closed","visitable":"closed","visible":"closed"}'::jsonb, "status"=1 where "visibility"=1;

ALTER TABLE nodes ADD COLUMN edit_time TIMESTAMP;

UPDATE nodes SET edit_time=updated_at ;
