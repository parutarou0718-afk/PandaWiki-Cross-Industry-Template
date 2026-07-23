ALTER Table comments ADD COLUMN status smallint NOT NULL DEFAULT 0;

UPDATE comments SET status = 1;