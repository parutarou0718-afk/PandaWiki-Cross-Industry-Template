ALTER TABLE stat_pages
ALTER COLUMN user_id TYPE text USING user_id::text;
UPDATE stat_pages SET user_id = '' WHERE user_id = NULL;
