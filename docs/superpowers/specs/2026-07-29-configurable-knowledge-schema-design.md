# Configurable Knowledge Schema Design

## Goal

Allow a PandaWiki administrator to define knowledge fields for one knowledge
base. The active schema controls server-side graph extraction, the LLM Wiki
Knowledge navigation, filters, and node property display.

## Boundaries

- PandaWiki remains the sole owner of document content, extraction, graph
  persistence, and permission filtering.
- LLM Wiki renders server DTOs only. It does not infer schema fields, persist
  remote attributes, or perform local entity extraction.
- System edition templates supply defaults. A knowledge base may override them
  without changing its existing documents or vector data.
- Updating a schema never silently rewrites historical graph data. An admin
  explicitly queues a graph rebuild when existing nodes must be re-extracted.

## Schema contract

Each knowledge base has one versioned schema with ordered navigation sections
and field definitions. A field has a stable key, display label, extraction
instruction, target (`entity`, `relation`, or `document`), supported entity
types, value type, multiple/filterable/enabled flags, and optional enum values.

Allowed value types are `text`, `number`, `date`, `boolean`, and `select`.
Keys are lower-case snake case and cannot change after field creation; label and
instruction can change. Unknown fields and values incompatible with their type
are rejected by the server.

Entity values are stored in `graph_entities.attributes` as JSONB. The response
also carries a field-safe node/evidence projection. Prompt construction includes
only enabled fields and their extraction instructions. The raw model response,
prompt, and full source document are never stored in graph attributes or sent
to the client.

## API

- `GET /api/v1/knowledge_base/graph/schema?kb_id=...` requires normal KB
  access and returns the safe effective schema.
- `PUT /api/v1/knowledge_base/graph/schema` requires document-management
  permission and validates the complete administrator schema.
- `POST /api/v1/knowledge_base/graph/rebuild?kb_id=...` uses the existing
  graph queue and current effective schema.
- `GET /api/v1/knowledge_base/graph?kb_id=...` returns effective navigation,
  entities with allowed attributes, relationships, and only evidence backed by
  source nodes the caller can visit.

## Data flow

1. Resolve the knowledge-base schema, falling back to the common template.
2. Generate a strict JSON extraction contract from enabled fields.
3. The graph worker reads the source release on the server and calls the
   configured model.
4. Validate entity types, relations, attributes, field value types, bounds,
   and enum values.
5. Transactionally replace that node's graph evidence and attributes.
6. At read time derive group IDs from the authenticated user and filter source
   evidence by current node visitability before returning entities or relations.

## Client behaviour

The Knowledge panel renders sections from `schema.navigation`. It groups only
the entities returned by the server. The graph surface uses the same graph DTO,
supports filtering by section and field, and opens a source through the remote
node reader. A field unknown to an older client is displayed as a plain,
escaped key/value property; it is never interpreted as HTML or code.

## Failure behaviour

Invalid administrator schema updates are rejected before persistence. An LLM
response with invalid or unknown fields is discarded for that node and logged
with IDs only. Model failures do not affect source documents, vector indexing,
or existing graph data. If the schema or graph endpoint is unavailable, LLM
Wiki hides the remote graph/navigation enhancement rather than falling back to
local document analysis.

## Tests

Unit and integration coverage must verify schema validation, prompt projection,
attribute validation, permission-filtered graph response, schema fallback,
rebuild queueing, and no leakage of full content or group IDs. Client coverage
must verify DTO mapping, server-controlled sections, safe property rendering,
and hidden capability when a server lacks the schema endpoint.
