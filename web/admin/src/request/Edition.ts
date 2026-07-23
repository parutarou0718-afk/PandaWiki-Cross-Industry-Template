import request from '@/api/request';

export type EditionConfig = {
  schema_version: number;
  edition_id: 'common' | 'research' | 'legal' | 'finance';
  edition_version: string;
  product_name: string;
  short_name: string;
  branding: { logo?: string };
  home_description: string;
  terminology: Record<string, string>;
  enabled_features: string[];
  document_types: string[];
  relation_types: string[];
  default_prompts: { chat: string; summary: string };
};

export type EditionOverrides = Partial<
  Omit<EditionConfig, 'schema_version' | 'edition_id' | 'edition_version'>
>;

export const getEdition = () =>
  request<EditionConfig>({ url: '/api/v1/system/edition', method: 'GET' });

export const updateEdition = (
  edition_id: EditionConfig['edition_id'],
  overrides: EditionOverrides,
) =>
  request<EditionConfig>({
    url: '/api/v1/system/edition',
    method: 'PUT',
    data: { edition_id, overrides },
  });
