ALTER TABLE nodes ADD COLUMN IF NOT EXISTS rag_info jsonb default '{}';
