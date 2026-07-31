/**
 * UI-only helpers for the knowledge-schema editor.
 *
 * Field keys are user-editable domain data, so they must never be used as
 * React keys. A changing React key remounts the field editor and drops focus.
 */
export const knowledgeSchemaFieldRenderKey = (index: number): string =>
  `knowledge-schema-field-${index}`;

export const canLoadKnowledgeSchema = (
  kbID: string | null | undefined,
): boolean => Boolean(kbID?.trim());
