-- Reverse auth_configs constraints
ALTER TABLE auth_configs DROP CONSTRAINT IF EXISTS uniq_auth_configs_source_type_kb_id;
ALTER TABLE auth_configs ADD CONSTRAINT auth_configs_source_type_key UNIQUE (source_type);

-- Drop kb_users table and constraints
ALTER TABLE "public"."kb_users" DROP CONSTRAINT IF EXISTS "uniq_kb_users_kb_id_user_id";
DROP TABLE IF EXISTS "public"."kb_users";

-- Remove role column from users table
ALTER TABLE "public"."users" DROP COLUMN IF EXISTS "role";