-- add dataset_id to knowledge_bases table
ALTER TABLE "public"."knowledge_bases" ADD COLUMN "dataset_id" text NULL;
