ALTER TABLE nodes DROP COLUMN permissions;


-- Drop tables
DROP TABLE IF EXISTS auth_groups;
DROP TABLE IF EXISTS node_auth_groups;

--Drop columns
ALTER TABLE "public"."nodes" DROP COLUMN "creator_id";
ALTER TABLE "public"."nodes" DROP COLUMN "editor_id";
ALTER TABLE "public"."nodes" DROP COLUMN "edit_time";
