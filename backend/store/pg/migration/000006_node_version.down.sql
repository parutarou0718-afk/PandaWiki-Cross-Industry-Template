-- drop node_releases table
DROP TABLE "public"."node_releases";

-- drop kb_releases table
DROP TABLE "public"."kb_releases";

-- drop kb_release_node_releases table
DROP TABLE "public"."kb_release_node_releases";

-- alter nodes table
ALTER TABLE "public"."nodes" DROP COLUMN "status";
ALTER TABLE "public"."nodes" DROP COLUMN "visibility";

-- drop migrations table
DROP TABLE "public"."migrations";
