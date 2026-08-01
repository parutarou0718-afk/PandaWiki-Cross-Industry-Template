CREATE TABLE IF NOT EXISTS graph_entity_summaries (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
    node_id TEXT NOT NULL,
    node_release_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT graph_entity_summaries_entity_node_unique UNIQUE (entity_id, node_id)
);

CREATE INDEX IF NOT EXISTS idx_graph_entity_summaries_kb_node ON graph_entity_summaries (kb_id, node_id);
CREATE INDEX IF NOT EXISTS idx_graph_entity_summaries_entity_id ON graph_entity_summaries (entity_id);
