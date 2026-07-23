-- delete embedding and rerank models
DELETE FROM models WHERE type = 'embedding' OR type = 'rerank';
