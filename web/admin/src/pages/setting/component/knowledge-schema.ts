export type KnowledgeValueType = 'text' | 'number' | 'date' | 'boolean' | 'select';

export interface KnowledgeSchemaField {
  key: string;
  label: string;
  target: 'entity';
  entity_types: string[];
  value_type: KnowledgeValueType;
  multiple: boolean;
  filterable: boolean;
  enabled: boolean;
  options: string[];
  extract_instruction: string;
}

export interface KnowledgeSchemaNavigation {
  id: string;
  label: string;
  entity_types: string[];
  field_keys: string[];
  order: number;
  enabled: boolean;
}

export interface KnowledgeSchema {
  version: number;
  fields: KnowledgeSchemaField[];
  navigation: KnowledgeSchemaNavigation[];
}

export const ENTITY_TYPES = ['person', 'organization', 'concept', 'method', 'event', 'document', 'other'];

export function createDefaultKnowledgeSchema(): KnowledgeSchema {
  return { version: 1, fields: [], navigation: [
    { id: 'overview', label: 'Overview', entity_types: ['document', 'event'], field_keys: [], order: 1, enabled: true },
    { id: 'entities', label: 'Entities', entity_types: ['person', 'organization'], field_keys: [], order: 2, enabled: true },
    { id: 'concepts', label: 'Concepts', entity_types: ['concept', 'method'], field_keys: [], order: 3, enabled: true },
  ] };
}

export function validateKnowledgeSchema(schema: KnowledgeSchema): string | null {
  const keys = schema.fields.map(field => field.key.trim());
  if (keys.some(key => !/^[a-z][a-z0-9_]{0,63}$/.test(key))) return 'Field keys must use lowercase letters, numbers, and underscores.';
  if (new Set(keys).size !== keys.length) return 'Field keys must be unique.';
  if (schema.fields.some(field => !field.label.trim())) return 'Every field needs a display name.';
  if (schema.fields.some(field => field.value_type === 'select' && field.options.filter(Boolean).length === 0)) return 'Select fields need at least one option.';
  const sectionIDs = schema.navigation.map(section => section.id.trim());
  if (sectionIDs.some(id => !/^[a-z][a-z0-9_]{0,63}$/.test(id))) return 'Navigation IDs must use lowercase letters, numbers, and underscores.';
  if (new Set(sectionIDs).size !== sectionIDs.length) return 'Navigation IDs must be unique.';
  if (new Set(schema.navigation.map(section => section.order)).size !== schema.navigation.length) return 'Navigation order must be unique.';
  return null;
}
