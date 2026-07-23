-- Add role column to users table
ALTER TABLE "public"."users" ADD COLUMN "role" text NOT NULL DEFAULT 'user';

-- Set existing users as admin
UPDATE "public"."users" SET "role" = 'admin';

-- Create kb_users table for user-kb permissions
CREATE TABLE "public"."kb_users" (
    "id" BIGSERIAL NOT NULL,
    "kb_id" text NOT NULL,
    "user_id" text NOT NULL,
    "perm" text NOT NULL DEFAULT 'full_control',
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

-- Add unique constraint for kb_id and user_id
ALTER TABLE "public"."kb_users" ADD CONSTRAINT "uniq_kb_users_kb_id_user_id" UNIQUE ("kb_id", "user_id");

-- Update auth_configs constraints
ALTER TABLE auth_configs DROP CONSTRAINT auth_configs_source_type_key;
ALTER TABLE auth_configs ADD CONSTRAINT uniq_auth_configs_source_type_kb_id UNIQUE (source_type, kb_id);