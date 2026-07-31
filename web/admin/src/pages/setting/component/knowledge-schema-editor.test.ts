import { strict as assert } from 'node:assert';
import {
  canLoadKnowledgeSchema,
  knowledgeSchemaFieldRenderKey,
} from './knowledge-schema-editor';

assert.equal(
  knowledgeSchemaFieldRenderKey(0),
  'knowledge-schema-field-0',
  'a field render key must not depend on the user-editable field key',
);
assert.equal(
  knowledgeSchemaFieldRenderKey(3),
  'knowledge-schema-field-3',
  'each field position needs a stable render key',
);
assert.equal(canLoadKnowledgeSchema(''), false);
assert.equal(canLoadKnowledgeSchema('kb-123'), true);
