-- Add parameters column to models table
ALTER TABLE "public"."models" ADD COLUMN "parameters" JSONB;