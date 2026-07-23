UPDATE stat_pages SET user_id = NULL WHERE user_id = '';
ALTER TABLE stat_pages
ALTER COLUMN user_id TYPE bigint USING user_id::bigint;