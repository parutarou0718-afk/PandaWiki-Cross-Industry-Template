import request from '@/api/request';

export type EditionID = 'common' | 'research' | 'legal' | 'finance';
export type ReportInputType =
  | 'text'
  | 'textarea'
  | 'select'
  | 'date'
  | 'date_range';

export type ReportProfileInputField = {
  key: string;
  label: string;
  type: ReportInputType;
  required: boolean;
  placeholder: string;
  default_value: string;
  options: string[];
};
export type ReportProfileSection = {
  key: string;
  title: string;
  instruction: string;
  required: boolean;
  order: number;
};
export type ReportProfile = {
  id: string;
  name: string;
  description: string;
  edition_ids: EditionID[];
  system_prompt: string;
  input_fields: ReportProfileInputField[];
  sections: ReportProfileSection[];
  citation_required: boolean;
  enabled: boolean;
  version: string;
  is_builtin: boolean;
  created_at: string;
  updated_at: string;
};
export type ReportProfilePayload = Omit<
  ReportProfile,
  'id' | 'is_builtin' | 'created_at' | 'updated_at'
>;

const base = '/api/v1/admin/report-profiles';
export const listAdminReportProfiles = () => request<ReportProfile[]>({ url: base, method: 'GET' });
export const getAdminReportProfile = (id: string) => request<ReportProfile>({ url: `${base}/${id}`, method: 'GET' });
export const createAdminReportProfile = (data: ReportProfilePayload) => request<ReportProfile>({ url: base, method: 'POST', data });
export const updateAdminReportProfile = (id: string, data: ReportProfilePayload) => request<ReportProfile>({ url: `${base}/${id}`, method: 'PUT', data });
export const setAdminReportProfileEnabled = (id: string, enabled: boolean) => request<{ enabled: boolean }>({ url: `${base}/${id}/enabled`, method: 'PATCH', data: { enabled } });
export const deleteAdminReportProfile = (id: string) => request<boolean>({ url: `${base}/${id}`, method: 'DELETE' });
export const restoreAdminReportProfileDefault = (id: string) => request<ReportProfile>({ url: `${base}/${id}/restore-default`, method: 'POST' });
