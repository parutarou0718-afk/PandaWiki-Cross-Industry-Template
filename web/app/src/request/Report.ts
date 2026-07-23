import httpRequest, { ContentType, RequestParams } from './httpClient';

type APIResponse<T> = { data?: T };

export type ReportInputType =
  | 'text'
  | 'textarea'
  | 'select'
  | 'date'
  | 'date_range';
export type ReportInputField = {
  key: string;
  label: string;
  type: ReportInputType;
  required: boolean;
  placeholder: string;
  default_value: string;
  options: string[];
};
export type ReportSection = {
  key: string;
  title: string;
  instruction: string;
  required: boolean;
  order: number;
};
export type PublicReportProfile = {
  id: string;
  name: string;
  description: string;
  input_fields: ReportInputField[];
  sections: ReportSection[];
  citation_required: boolean;
  version: string;
};
export type ReportDateRange = { start: string; end: string };
export type ReportInputValues = Record<string, string | ReportDateRange>;
export type ReportCitation = {
  citation_index: number;
  node_id: string;
  document_name: string;
  locator: string;
  excerpt: string;
};
export type ReportStatus = 'pending' | 'running' | 'completed' | 'failed';
export type ReportResultCode =
  | ''
  | 'success'
  | 'insufficient_evidence'
  | 'model_error'
  | 'retrieval_error'
  | 'invalid_citation'
  | 'timeout';
export type Report = {
  id: string;
  kb_id: string;
  profile_id: string;
  profile_version: string;
  profile_name: string;
  status: ReportStatus;
  result_code: ReportResultCode;
  title: string;
  input_values: ReportInputValues;
  content: string;
  error_message: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  citation_count: number;
};
export type ReportDetail = { report: Report; citations: ReportCitation[] };
export type CreateReportPayload = {
  kb_id: string;
  profile_id: string;
  title: string;
  input_values: ReportInputValues;
  scope: { node_ids: string[] };
};

export const getReportProfiles = (kbID: string, params: RequestParams = {}) =>
  httpRequest<APIResponse<PublicReportProfile[]>>({
    path: '/api/v1/report-profiles', method: 'GET', query: { kb_id: kbID },
    type: ContentType.Json, format: 'json', ...params,
  });
export const createReport = (body: CreateReportPayload, params: RequestParams = {}) =>
  httpRequest<APIResponse<ReportDetail>>({ path: '/api/v1/reports', method: 'POST', body, type: ContentType.Json, format: 'json', ...params });
export const getReports = (params: RequestParams = {}) =>
  httpRequest<APIResponse<Report[]>>({ path: '/api/v1/reports', method: 'GET', type: ContentType.Json, format: 'json', ...params });
export const getReport = (id: string, params: RequestParams = {}) =>
  httpRequest<APIResponse<ReportDetail>>({ path: `/api/v1/reports/${id}`, method: 'GET', type: ContentType.Json, format: 'json', ...params });
