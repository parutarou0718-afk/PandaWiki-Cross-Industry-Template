ALTER TABLE reports
    ADD COLUMN result_code TEXT NOT NULL DEFAULT '';

ALTER TABLE reports
    ADD CONSTRAINT chk_reports_result_code
    CHECK (result_code IN ('', 'success', 'insufficient_evidence', 'model_error', 'retrieval_error', 'invalid_citation', 'timeout'));
