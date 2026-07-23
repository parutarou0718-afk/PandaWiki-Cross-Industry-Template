'use client';

import MarkDown from '@/components/markdown';
import { useBasePath } from '@/hooks';
import { useStore } from '@/provider';
import {
  createReport,
  CreateReportPayload,
  getReport,
  getReportProfiles,
  getReports,
  PublicReportProfile,
  Report,
  ReportCitation,
  ReportDetail,
  ReportInputValues,
  ReportResultCode,
} from '@/request/Report';
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  Chip,
  CircularProgress,
  Divider,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import { useRouter } from 'next/navigation';
import { useEffect, useMemo, useState } from 'react';

const resultText: Record<
  ReportResultCode,
  { label: string; severity: 'success' | 'info' | 'error' | 'warning' }
> = {
  '': { label: 'In progress', severity: 'info' },
  success: { label: 'Completed', severity: 'success' },
  insufficient_evidence: {
    label: 'Insufficient materials',
    severity: 'warning',
  },
  model_error: { label: 'Model generation failed', severity: 'error' },
  retrieval_error: { label: 'Retrieval failed', severity: 'error' },
  invalid_citation: { label: 'Invalid citation', severity: 'error' },
  timeout: { label: 'Generation timed out', severity: 'error' },
};
const errorText = (error: unknown) => {
  const code = (error as { code?: number })?.code;
  if (code === 403) return 'You do not have permission to access this report.';
  if (code === 404) return 'Report not found or you do not have access to it.';
  return 'Request failed. Please try again.';
};
const formatDate = (value?: string) =>
  value ? new Date(value).toLocaleString() : '—';
const citationMarkdown = (content: string, citations: ReportCitation[]) => {
  const known = new Set(citations.map(item => item.citation_index));
  return content.replace(/\[(\d+)\]/g, (match, raw) =>
    known.has(Number(raw)) ? `[${raw}](#citation-${raw})` : match,
  );
};

export const ReportsList = () => {
  const router = useRouter();
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState<'all' | Report['status']>('all');
  useEffect(() => {
    getReports()
      .then(setReports)
      .catch(error => setError(errorText(error)))
      .finally(() => setLoading(false));
  }, []);
  const filtered = useMemo(
    () =>
      reports.filter(
        report =>
          (status === 'all' || report.status === status) &&
          `${report.title} ${report.profile_name}`
            .toLowerCase()
            .includes(search.toLowerCase()),
      ),
    [reports, status, search],
  );
  if (loading)
    return (
      <PageFrame>
        <CircularProgress />
      </PageFrame>
    );
  return (
    <PageFrame>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent='space-between'
        spacing={2}
      >
        <Typography variant='h4'>Reports</Typography>
        <Button variant='contained' onClick={() => router.push('/reports/new')}>
          New report
        </Button>
      </Stack>
      {error && <Alert severity='error'>{error}</Alert>}
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
        <TextField
          label='Search reports'
          value={search}
          onChange={event => setSearch(event.target.value)}
        />
        <TextField
          select
          label='Status'
          value={status}
          onChange={event => setStatus(event.target.value as typeof status)}
        >
          <MenuItem value='all'>All</MenuItem>
          {(['pending', 'running', 'completed', 'failed'] as const).map(
            item => (
              <MenuItem key={item} value={item}>
                {item}
              </MenuItem>
            ),
          )}
        </TextField>
      </Stack>
      {filtered.length === 0 ? (
        <Alert severity='info'>No reports yet.</Alert>
      ) : (
        <Stack spacing={1}>
          {filtered.map(report => (
            <Card key={report.id}>
              <CardActionArea
                onClick={() => router.push(`/reports/${report.id}`)}
              >
                <Stack spacing={1} sx={{ p: 2 }}>
                  <Typography variant='h6'>
                    {report.title || report.profile_name || 'Untitled report'}
                  </Typography>
                  <Typography variant='body2'>
                    {report.profile_name} · v{report.profile_version}
                  </Typography>
                  <Stack direction='row' spacing={1} flexWrap='wrap'>
                    <Chip label={report.status} />
                    <Chip
                      color={resultText[report.result_code].severity}
                      label={resultText[report.result_code].label}
                    />
                    <Chip label={`${report.citation_count} citations`} />
                  </Stack>
                  <Typography variant='caption'>
                    Created {formatDate(report.created_at)} · Completed{' '}
                    {formatDate(report.completed_at)}
                  </Typography>
                </Stack>
              </CardActionArea>
            </Card>
          ))}
        </Stack>
      )}
    </PageFrame>
  );
};

export const NewReport = () => {
  const router = useRouter();
  const { kbDetail } = useStore();
  const kbID = kbDetail?.kb_id || '';
  const [profiles, setProfiles] = useState<PublicReportProfile[]>([]);
  const [profile, setProfile] = useState<PublicReportProfile>();
  const [values, setValues] = useState<ReportInputValues>({});
  const [title, setTitle] = useState('');
  const [scopeText, setScopeText] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => {
    if (!kbID) return;
    getReportProfiles(kbID)
      .then(items => {
        setProfiles(items);
        setProfile(items[0]);
      })
      .catch(error => setError(errorText(error)))
      .finally(() => setLoading(false));
  }, [kbID]);
  useEffect(() => {
    if (!profile) return;
    const defaults: ReportInputValues = {};
    profile.input_fields.forEach(field => {
      if (field.default_value) defaults[field.key] = field.default_value;
    });
    setValues(defaults);
  }, [profile?.id]);
  const missingRequired = profile?.input_fields.some(field => {
    const input = values[field.key];
    if (!field.required || !input) return false;
    if (typeof input === 'string') return !input.trim();
    return !input.start || !input.end;
  });
  const submit = async () => {
    if (!profile || !kbID || missingRequired || saving) return;
    setSaving(true);
    setError('');
    const scope = scopeText
      .split(/[\n,]/)
      .map(item => item.trim())
      .filter(Boolean);
    const body: CreateReportPayload = {
      kb_id: kbID,
      profile_id: profile.id,
      title: title.trim(),
      input_values: values,
      scope: { node_ids: scope },
    };
    try {
      const detail = await createReport(body);
      router.replace(`/reports/${detail.report.id}`);
    } catch (error) {
      setError(errorText(error));
    } finally {
      setSaving(false);
    }
  };
  if (loading)
    return (
      <PageFrame>
        <CircularProgress />
      </PageFrame>
    );
  if (!kbID)
    return (
      <PageFrame>
        <Alert severity='error'>Knowledge base context is unavailable.</Alert>
      </PageFrame>
    );
  if (!profiles.length)
    return (
      <PageFrame>
        <Typography variant='h4'>New report</Typography>
        <Alert severity='info'>
          No report templates are available for this knowledge base.
        </Alert>
      </PageFrame>
    );
  return (
    <PageFrame>
      <Button onClick={() => router.push('/reports')}>Back to reports</Button>
      <Typography variant='h4'>New report</Typography>
      <Alert severity='info'>
        Generation may take some time. Keep this page open and do not submit
        twice.
      </Alert>
      {error && <Alert severity='error'>{error}</Alert>}
      <Stack spacing={2}>
        <TextField
          select
          label='Template'
          value={profile?.id || ''}
          onChange={event =>
            setProfile(profiles.find(item => item.id === event.target.value))
          }
        >
          {profiles.map(item => (
            <MenuItem key={item.id} value={item.id}>
              {item.name} · v{item.version}
            </MenuItem>
          ))}
        </TextField>
        {profile && (
          <Card sx={{ p: 2 }}>
            <Typography variant='h6'>{profile.name}</Typography>
            <Typography variant='body2'>{profile.description}</Typography>
            <Typography variant='caption'>
              {profile.citation_required
                ? 'Citations required'
                : 'Citations optional'}{' '}
              · {profile.sections.map(section => section.title).join(' · ')}
            </Typography>
          </Card>
        )}
        <TextField
          label='Title (optional)'
          value={title}
          onChange={event => setTitle(event.target.value)}
        />
        {profile?.input_fields.map(field => (
          <DynamicField
            key={field.key}
            field={field}
            value={values[field.key]}
            onChange={value =>
              setValues(current => ({ ...current, [field.key]: value }))
            }
          />
        ))}
        <TextField
          label='Limit to node IDs (optional)'
          multiline
          minRows={2}
          value={scopeText}
          onChange={event => setScopeText(event.target.value)}
          helperText='Comma or line separated IDs. This only narrows the authorized retrieval scope.'
        />
        <Button
          variant='contained'
          disabled={saving || !!missingRequired}
          onClick={() => void submit()}
        >
          {saving ? 'Generating report…' : 'Generate report'}
        </Button>
      </Stack>
    </PageFrame>
  );
};

const DynamicField = ({
  field,
  value,
  onChange,
}: {
  field: PublicReportProfile['input_fields'][number];
  value: ReportInputValues[string];
  onChange: (value: ReportInputValues[string]) => void;
}) => {
  if (field.type === 'select')
    return (
      <TextField
        select
        required={field.required}
        label={field.label || field.key}
        value={typeof value === 'string' ? value : ''}
        onChange={event => onChange(event.target.value)}
        helperText={field.placeholder}
      >
        {field.options.map(option => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </TextField>
    );
  if (field.type === 'date_range') {
    const range =
      typeof value === 'object' && value ? value : { start: '', end: '' };
    return (
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
        <TextField
          required={field.required}
          type='date'
          label={`${field.label || field.key} start`}
          InputLabelProps={{ shrink: true }}
          value={range.start}
          onChange={event => onChange({ ...range, start: event.target.value })}
        />
        <TextField
          required={field.required}
          type='date'
          label={`${field.label || field.key} end`}
          InputLabelProps={{ shrink: true }}
          value={range.end}
          onChange={event => onChange({ ...range, end: event.target.value })}
        />
      </Stack>
    );
  }
  return (
    <TextField
      required={field.required}
      type={field.type === 'date' ? 'date' : 'text'}
      multiline={field.type === 'textarea'}
      minRows={field.type === 'textarea' ? 4 : undefined}
      label={field.label || field.key}
      placeholder={field.placeholder}
      InputLabelProps={field.type === 'date' ? { shrink: true } : undefined}
      value={typeof value === 'string' ? value : ''}
      onChange={event => onChange(event.target.value)}
    />
  );
};

export const ReportPage = ({ id }: { id: string }) => {
  const router = useRouter();
  const [detail, setDetail] = useState<ReportDetail>();
  const [error, setError] = useState('');
  useEffect(() => {
    getReport(id)
      .then(setDetail)
      .catch(error => setError(errorText(error)));
  }, [id]);
  if (error)
    return (
      <PageFrame>
        <Alert severity='error'>{error}</Alert>
        <Button onClick={() => router.push('/reports')}>Back to reports</Button>
      </PageFrame>
    );
  if (!detail)
    return (
      <PageFrame>
        <CircularProgress />
      </PageFrame>
    );
  const { report, citations } = detail;
  const result = resultText[report.result_code];
  const content = citationMarkdown(report.content, citations);
  return (
    <PageFrame>
      <Button onClick={() => router.push('/reports')}>Back to reports</Button>
      <Stack spacing={2}>
        <Typography variant='h4'>
          {report.title || report.profile_name || 'Untitled report'}
        </Typography>
        <Typography>
          {report.profile_name} · v{report.profile_version}
        </Typography>
        <Stack direction='row' spacing={1}>
          <Chip label={report.status} />
          <Chip color={result.severity} label={result.label} />
        </Stack>
        <Typography variant='caption'>
          Created {formatDate(report.created_at)} · Completed{' '}
          {formatDate(report.completed_at)}
        </Typography>
        {report.result_code === 'insufficient_evidence' && (
          <Alert severity='warning'>
            There was not enough authorized material to provide a cited report.
          </Alert>
        )}
        {report.status === 'failed' && (
          <Alert severity='error'>
            {result.label}. {report.error_message || 'Please try again later.'}
          </Alert>
        )}
        <Divider />
        {report.content && <MarkDown content={content} />}
        <Divider />
        <Typography variant='h6'>Sources</Typography>
        {citations.length === 0 ? (
          <Typography variant='body2'>
            No sources are available for this report.
          </Typography>
        ) : (
          <Stack spacing={1}>
            {[...citations]
              .sort((a, b) => a.citation_index - b.citation_index)
              .map(citation => (
                <Card
                  id={`citation-${citation.citation_index}`}
                  key={citation.citation_index}
                  sx={{ p: 2 }}
                >
                  <Typography variant='subtitle2'>
                    [{citation.citation_index}] {citation.document_name}
                  </Typography>
                  <Typography variant='caption'>
                    {citation.locator || 'Location unavailable'}
                  </Typography>
                  <Typography sx={{ whiteSpace: 'pre-wrap' }}>
                    {citation.excerpt || 'Citation unavailable'}
                  </Typography>
                </Card>
              ))}
          </Stack>
        )}
      </Stack>
    </PageFrame>
  );
};

const PageFrame = ({ children }: { children: React.ReactNode }) => (
  <Stack
    spacing={2}
    sx={{ maxWidth: 960, mx: 'auto', p: { xs: 2, md: 4 }, mt: 8 }}
  >
    {children}
  </Stack>
);
