ALTER TABLE reports DROP CONSTRAINT IF EXISTS chk_reports_result_code;
ALTER TABLE reports DROP COLUMN IF EXISTS result_code;
