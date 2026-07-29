import request from './request';
import type { KnowledgeSchema } from '@/pages/setting/component/knowledge-schema';

export const getKnowledgeSchema = (kb_id: string): Promise<KnowledgeSchema> =>
  request({ url: '/api/v1/knowledge_base/graph/schema', method: 'get', params: { kb_id } });

export const putKnowledgeSchema = (kb_id: string, schema: KnowledgeSchema): Promise<KnowledgeSchema> =>
  request({ url: '/api/v1/knowledge_base/graph/schema', method: 'put', data: { kb_id, schema } });
