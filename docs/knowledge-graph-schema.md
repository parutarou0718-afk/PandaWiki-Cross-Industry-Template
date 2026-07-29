# Configurable Knowledge Graph Schema

## What it controls

Each knowledge base can define its own entity fields and Knowledge navigation
sections. PandaWiki sends enabled field definitions and extraction instructions
to the configured server-side chat model during graph extraction. The model
returns structured values, PandaWiki validates them, and stores them as JSONB
attributes on graph entities.

This is not a client-side analysis feature. LLM Wiki only displays the schema
and graph projection returned by PandaWiki.

## Administrator workflow

1. Open a knowledge base in PandaWiki Admin.
2. Open **Settings → Knowledge model**.
3. Add fields such as `argument`, `risk_level`, or `publication_year`.
4. Use lower-case snake-case keys. Select fields require at least one option.
5. Add or edit navigation groups to decide which entity types appear under the
   remote Knowledge panel.
6. Save. New document graph extraction uses the saved schema.
7. Choose **Rebuild graph from documents** to re-extract existing documents.

Only a user with `full_control` permission for that knowledge base can save a
schema. Graph reads still use each reader's server-side permissions.

## V0.1 limits

- Custom fields apply to graph entities. Relation and document fields are not
  supported yet.
- Changing a schema does not modify documents, vectors, RAG settings, or old
  graph facts automatically. Use the explicit rebuild action.
- The configured PandaWiki server model performs extraction. Model prompts and
  full source documents are not returned to desktop clients.

## Upgrade sequence

For a self-hosted deployment, update the source branch, rebuild the API,
consumer, and admin images with the repository's normal deployment scripts,
then start the stack. The API's normal migration startup applies
`000040_create_knowledge_graph_schema` before serving requests.

After the containers are healthy, open **Settings → Knowledge model**, save
the intended schema, and queue a graph rebuild. Keep the existing PostgreSQL,
MinIO, Qdrant, and `.env` data directories; this upgrade does not require a
destructive data reset.

## API summary

- `GET /api/v1/knowledge_base/graph/schema?kb_id=<id>`: effective schema for
  an authorized reader.
- `PUT /api/v1/knowledge_base/graph/schema`: saves `{ "kb_id": "...",
  "schema": { ... } }` for a full-control user.
- `POST /api/v1/knowledge_base/graph/rebuild`: queues `{ "kb_id": "..." }`
  for graph extraction using the current schema.
