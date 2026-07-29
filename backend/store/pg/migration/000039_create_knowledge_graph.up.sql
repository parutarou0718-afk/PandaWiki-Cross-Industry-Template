CREATE TABLE IF NOT EXISTS graph_entities (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT graph_entities_kb_name_unique UNIQUE (kb_id, name_key)
);

CREATE INDEX IF NOT EXISTS idx_graph_entities_kb_id ON graph_entities (kb_id);

CREATE TABLE IF NOT EXISTS graph_relations (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    source_entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
    target_entity_id TEXT NOT NULL REFERENCES graph_entities(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT graph_relations_kb_source_target_type_unique UNIQUE (kb_id, source_entity_id, target_entity_id, type)
);

CREATE INDEX IF NOT EXISTS idx_graph_relations_kb_id ON graph_relations (kb_id);

CREATE TABLE IF NOT EXISTS graph_evidence (
    id TEXT PRIMARY KEY,
    kb_id TEXT NOT NULL,
    relation_id TEXT NOT NULL REFERENCES graph_relations(id) ON DELETE CASCADE,
    node_id TEXT NOT NULL,
    node_release_id TEXT NOT NULL,
    excerpt TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT graph_evidence_relation_node_unique UNIQUE (relation_id, node_id)
);

CREATE INDEX IF NOT EXISTS idx_graph_evidence_kb_node ON graph_evidence (kb_id, node_id);
CREATE INDEX IF NOT EXISTS idx_graph_evidence_relation_id ON graph_evidence (relation_id);
