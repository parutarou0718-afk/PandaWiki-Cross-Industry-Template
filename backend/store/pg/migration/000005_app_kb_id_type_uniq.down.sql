-- Drop index "idx_apps_kb_id_type" to table: "apps"
DROP INDEX IF EXISTS "idx_apps_kb_id_type";

-- Create index "idx_apps_kb_id" to table: "apps"
CREATE INDEX "idx_apps_kb_id" ON "public"."apps" ("kb_id");
