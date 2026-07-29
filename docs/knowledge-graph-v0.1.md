# Knowledge Graph V0.1

## Scope

PandaWiki extracts a compact, structured graph from server-side node content
after the existing vector-upsert queue successfully processes a node. The
desktop client is a graph viewer only: it neither downloads a corpus nor runs
local embedding, entity extraction, or permission filtering.

## Data retained

- `graph_entities`: normalized entity name, type and knowledge-base scope.
- `graph_relations`: typed directed relation between two entities.
- `graph_evidence`: relation-to-node provenance, node release ID, and an
  excerpt capped at 512 characters.

The graph tables do not retain a full node body, a model prompt, or model raw
response. Reprocessing a node replaces that node's old evidence transactionally.

## Permissions

`GET /api/v1/knowledge_base/graph?kb_id=<id>` requires normal PandaWiki login
and knowledge-base access. It derives the caller's group IDs on the server and
returns only evidence whose source node is currently `visitable` by that user.
Clients cannot submit `group_ids`.

## Queue behaviour

The existing vector `upsert` consumer invokes graph extraction only after RAG
upsert succeeds. A graph-model failure is logged with IDs only and does not
roll back vector indexing, node editing, or document reading. A later node
update retries graph extraction naturally.

## Client contract

The future `GraphProvider` consumes the endpoint above and renders entity and
relation data. Selecting a source opens the already-existing node-detail route.
It must not use local project paths or local graph-analysis code for a
PandaWiki virtual project.
