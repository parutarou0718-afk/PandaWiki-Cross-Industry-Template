# Knowledge Schema Admin Design

## Goal

Allow a knowledge-base full-control administrator to define the entity fields
and navigation sections that PandaWiki uses for server-side graph extraction
and that LLM Wiki renders for a remote knowledge base.

## Design

The existing Knowledge Base Settings screen gains a `Knowledge model` tab. It
uses the existing authenticated graph-schema endpoints and never stores schema
state in the browser as an authority. The server remains the only source of
truth.

The first UI supports entity fields only. A field has a stable lowercase key,
display label, entity types, value type, multiple/filterable flags, select
options and an extraction instruction. Navigation sections have a label,
entity types, optional field keys, order and enabled flag. The backend rejects
unknown fields and invalid schemas.

## Safety

Only `full_control` knowledge-base users can save. Readers may retrieve the
effective schema indirectly through the graph projection but cannot edit it.
The UI never accepts arbitrary JSON; it sends the explicit schema structure.

## Scope

Included: schema editing, reset to server defaults, field/section validation,
and server save/load feedback.

Excluded: relation/document custom fields, automatic extraction migration of
historical graph records, and a new global system-settings page.
