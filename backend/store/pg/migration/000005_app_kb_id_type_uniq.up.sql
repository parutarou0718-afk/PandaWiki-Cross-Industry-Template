-- Create unique index "idx_apps_kb_id_type" to table: "apps"
CREATE UNIQUE INDEX "idx_apps_kb_id_type" ON "public"."apps" ("kb_id", "type");

-- Drop index "idx_apps_kb_id" to table: "apps"
DROP INDEX IF EXISTS "idx_apps_kb_id";
