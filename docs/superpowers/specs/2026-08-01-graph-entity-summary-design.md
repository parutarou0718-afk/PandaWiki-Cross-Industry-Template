# Graph Entity Summary Design

## Goal

Provide a durable, permission-safe LLM summary for each PandaWiki knowledge-graph entity so desktop clients can show a useful entity inspector without generating text locally.

## Scope

The backend owns summary generation and persistence. The existing graph endpoint returns the stored summary with each visible entity. The research-work desktop application renders it in the existing graph inspector.

This work does not add a second graph store, local extraction, local content cache, client-side LLM call, new user roles, or an API that bypasses knowledge-base permissions.

## Data Flow

```text
Node graph extraction finishes
  -> graph facts, relations, and bounded evidence are persisted
  -> affected entities are queued for summary refresh
  -> server loads the entity's schema-approved attributes, relations, and bounded evidence
  -> configured graph model generates a concise entity summary
  -> summary is stored with the entity
  -> GET /api/v1/knowledge_base/graph returns only entities already visible to the caller
  -> desktop graph inspector displays the returned summary and source links
```

## Data Contract

`GraphEntity` gains an optional `summary` text field. Its value is generated only from the entity's existing, server-authorized graph projection:

- entity name and type;
- schema-approved attributes;
- direct relation labels and adjacent entity names;
- bounded evidence excerpts already associated with those relations.

The field is optional. Existing entities and failed refreshes return an empty summary; clients must show a neutral `Summary unavailable` state rather than invent text.

The graph API remains permission-filtered before serialization. It returns `summary` only alongside an entity that the caller may already receive. It never exposes an unbounded source node body, raw model prompt, model response trace, or another user's groups.

## Refresh and Failure Rules

- A node graph refresh schedules summaries for only the entities affected by that node's extracted facts.
- A knowledge-base rebuild may schedule all currently visible graph entities in bounded batches.
- Summary failure is non-fatal to graph extraction: facts and relations remain available, and the prior stored summary is kept.
- Model output is length-bounded and plain text. It is logged only as status and counts, never in full.
- The database migration is forward-compatible: nullable summary starts empty for existing graph rows.

## Desktop Behavior

The desktop client does not call an LLM. On graph-node selection, it reads the `summary` received in the normal graph projection and renders this inspector order:

1. entity name, type, and Schema fields;
2. server-generated entity summary;
3. relationship count and relation labels;
4. existing bounded evidence cards; clicking a card opens the permission-checked PandaWiki node detail.

Loading, empty summary, graph not generated, unauthorized, and request-error states remain distinct. Local projects retain their existing local graph behavior.

## Verification

Backend tests cover: migration/default empty value, summary context excludes unapproved fields, model failure preserves previous summary, and graph response includes summary only through the existing permission-filtered path.

Desktop tests cover: mapper preserves a server summary, inspector displays a summary, and an empty summary displays the neutral state without a client LLM request.
