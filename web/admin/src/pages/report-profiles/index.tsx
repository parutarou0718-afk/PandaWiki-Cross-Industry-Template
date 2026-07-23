import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  FormControlLabel,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography,
} from '@mui/material';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Card from '@/components/Card';
import { ConstsUserRole } from '@/request/types';
import { useAppSelector } from '@/store';
import {
  createAdminReportProfile,
  deleteAdminReportProfile,
  EditionID,
  listAdminReportProfiles,
  ReportInputType,
  ReportProfile,
  ReportProfileInputField,
  ReportProfilePayload,
  ReportProfileSection,
  restoreAdminReportProfileDefault,
  setAdminReportProfileEnabled,
  updateAdminReportProfile,
} from '@/request/ReportProfile';

const editionIDs: EditionID[] = ['common', 'research', 'legal', 'finance'];
const inputTypes: ReportInputType[] = [
  'text',
  'textarea',
  'select',
  'date',
  'date_range',
];
const blankField = (): ReportProfileInputField => ({
  key: '',
  label: '',
  type: 'text',
  required: false,
  placeholder: '',
  default_value: '',
  options: [],
});
const blankSection = (order: number): ReportProfileSection => ({
  key: '',
  title: '',
  instruction: '',
  required: false,
  order,
});
const blankProfile = (): ReportProfilePayload => ({
  name: '',
  description: '',
  edition_ids: ['common'],
  system_prompt: '',
  input_fields: [blankField()],
  sections: [blankSection(1)],
  citation_required: true,
  enabled: true,
  version: '1.0.0',
});
const payloadOf = (profile: ReportProfile): ReportProfilePayload => ({
  name: profile.name,
  description: profile.description,
  edition_ids: profile.edition_ids,
  system_prompt: profile.system_prompt,
  input_fields: profile.input_fields,
  sections: profile.sections,
  citation_required: profile.citation_required,
  enabled: profile.enabled,
  version: profile.version,
});
const clone = <T,>(value: T): T => JSON.parse(JSON.stringify(value)) as T;

const ProfileEditor = ({
  value,
  onChange,
  builtin,
  saving,
}: {
  value: ReportProfilePayload;
  onChange: (value: ReportProfilePayload) => void;
  builtin: boolean;
  saving: boolean;
}) => {
  const set = <K extends keyof ReportProfilePayload>(
    key: K,
    next: ReportProfilePayload[K],
  ) => onChange({ ...value, [key]: next });
  const inputKeys = value.input_fields.map(field => field.key).filter(Boolean);
  const sectionKeys = value.sections
    .map(section => section.key)
    .filter(Boolean);
  const orders = value.sections.map(section => section.order);
  return (
    <Stack spacing={2}>
      <TextField
        label='Name'
        value={value.name}
        disabled={saving}
        onChange={event => set('name', event.target.value)}
      />
      <TextField
        label='Description'
        value={value.description}
        disabled={saving}
        multiline
        minRows={2}
        onChange={event => set('description', event.target.value)}
      />
      <TextField
        label='Version'
        value={value.version}
        disabled={saving}
        onChange={event => set('version', event.target.value)}
        helperText='Template version, not database row version'
      />
      <Typography variant='subtitle1'>Applicable editions</Typography>
      <Stack direction='row' flexWrap='wrap'>
        {editionIDs.map(edition => (
          <FormControlLabel
            key={edition}
            control={
              <Checkbox
                checked={value.edition_ids.includes(edition)}
                disabled={saving}
                onChange={event =>
                  set(
                    'edition_ids',
                    event.target.checked
                      ? [...value.edition_ids, edition]
                      : value.edition_ids.filter(item => item !== edition),
                  )
                }
              />
            }
            label={edition}
          />
        ))}
      </Stack>
      <TextField
        label='System prompt'
        multiline
        minRows={6}
        value={value.system_prompt}
        disabled={saving}
        inputProps={{ maxLength: 16000 }}
        helperText={`${value.system_prompt.length}/16000`}
        onChange={event => set('system_prompt', event.target.value)}
      />
      <FormControlLabel
        control={
          <Switch
            checked={value.citation_required}
            disabled={saving}
            onChange={event => set('citation_required', event.target.checked)}
          />
        }
        label='Citations required'
      />
      <FormControlLabel
        control={
          <Switch
            checked={value.enabled}
            disabled={saving}
            onChange={event => set('enabled', event.target.checked)}
          />
        }
        label='Enabled'
      />
      <Typography variant='subtitle1'>Input fields</Typography>
      {value.input_fields.map((field, index) => (
        <Card key={index}>
          <Stack spacing={1}>
            <Stack direction='row' spacing={1}>
              <TextField
                label='Key'
                value={field.key}
                error={
                  field.key !== '' &&
                  inputKeys.filter(key => key === field.key).length > 1
                }
                helperText={
                  field.key !== '' &&
                  inputKeys.filter(key => key === field.key).length > 1
                    ? 'Duplicate key'
                    : ''
                }
                onChange={event => {
                  const next = clone(value.input_fields);
                  next[index].key = event.target.value;
                  set('input_fields', next);
                }}
              />
              <TextField
                label='Label'
                value={field.label}
                onChange={event => {
                  const next = clone(value.input_fields);
                  next[index].label = event.target.value;
                  set('input_fields', next);
                }}
              />
              <TextField
                select
                label='Type'
                value={field.type}
                onChange={event => {
                  const next = clone(value.input_fields);
                  next[index].type = event.target.value as ReportInputType;
                  set('input_fields', next);
                }}
              >
                {inputTypes.map(type => (
                  <MenuItem key={type} value={type}>
                    {type}
                  </MenuItem>
                ))}
              </TextField>
            </Stack>
            <TextField
              label='Placeholder'
              value={field.placeholder}
              onChange={event => {
                const next = clone(value.input_fields);
                next[index].placeholder = event.target.value;
                set('input_fields', next);
              }}
            />
            <TextField
              label='Default value'
              value={field.default_value}
              onChange={event => {
                const next = clone(value.input_fields);
                next[index].default_value = event.target.value;
                set('input_fields', next);
              }}
            />
            {field.type === 'select' && (
              <TextField
                label='Options (one per line)'
                required
                error={!field.options.length}
                helperText={
                  !field.options.length
                    ? 'Select needs at least one option'
                    : ''
                }
                multiline
                value={field.options.join('\n')}
                onChange={event => {
                  const next = clone(value.input_fields);
                  next[index].options = event.target.value
                    .split('\n')
                    .map(item => item.trim())
                    .filter(Boolean);
                  set('input_fields', next);
                }}
              />
            )}
            <FormControlLabel
              control={
                <Checkbox
                  checked={field.required}
                  onChange={event => {
                    const next = clone(value.input_fields);
                    next[index].required = event.target.checked;
                    set('input_fields', next);
                  }}
                />
              }
              label='Required'
            />
            <Button
              color='error'
              disabled={saving || value.input_fields.length === 1}
              onClick={() =>
                set(
                  'input_fields',
                  value.input_fields.filter(
                    (_, itemIndex) => itemIndex !== index,
                  ),
                )
              }
            >
              Remove field
            </Button>
          </Stack>
        </Card>
      ))}
      <Button
        disabled={saving}
        onClick={() =>
          set('input_fields', [...value.input_fields, blankField()])
        }
      >
        Add input field
      </Button>
      <Typography variant='subtitle1'>Sections</Typography>
      {value.sections.map((section, index) => (
        <Card key={index}>
          <Stack spacing={1}>
            <Stack direction='row' spacing={1}>
              <TextField
                label='Key'
                value={section.key}
                error={
                  section.key !== '' &&
                  sectionKeys.filter(key => key === section.key).length > 1
                }
                helperText={
                  section.key !== '' &&
                  sectionKeys.filter(key => key === section.key).length > 1
                    ? 'Duplicate key'
                    : ''
                }
                onChange={event => {
                  const next = clone(value.sections);
                  next[index].key = event.target.value;
                  set('sections', next);
                }}
              />
              <TextField
                label='Title'
                value={section.title}
                onChange={event => {
                  const next = clone(value.sections);
                  next[index].title = event.target.value;
                  set('sections', next);
                }}
              />
              <TextField
                label='Order'
                type='number'
                value={section.order}
                error={
                  orders.filter(order => order === section.order).length > 1
                }
                helperText={
                  orders.filter(order => order === section.order).length > 1
                    ? 'Duplicate order'
                    : ''
                }
                onChange={event => {
                  const next = clone(value.sections);
                  next[index].order = Number(event.target.value);
                  set('sections', next);
                }}
              />
            </Stack>
            <TextField
              label='Instruction'
              multiline
              minRows={3}
              value={section.instruction}
              onChange={event => {
                const next = clone(value.sections);
                next[index].instruction = event.target.value;
                set('sections', next);
              }}
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={section.required}
                  onChange={event => {
                    const next = clone(value.sections);
                    next[index].required = event.target.checked;
                    set('sections', next);
                  }}
                />
              }
              label='Required'
            />
            <Button
              color='error'
              disabled={saving || value.sections.length === 1}
              onClick={() =>
                set(
                  'sections',
                  value.sections.filter((_, itemIndex) => itemIndex !== index),
                )
              }
            >
              Remove section
            </Button>
          </Stack>
        </Card>
      ))}
      <Button
        disabled={saving}
        onClick={() =>
          set('sections', [
            ...value.sections,
            blankSection(value.sections.length + 1),
          ])
        }
      >
        Add section
      </Button>
      {builtin && (
        <Alert severity='info'>
          Built-in ID and built-in flag are server-owned. You may edit the
          template content or restore the current preset.
        </Alert>
      )}
    </Stack>
  );
};

const ReportProfiles = () => {
  const navigate = useNavigate();
  const user = useAppSelector(state => state.config.user);
  const isAdmin = user.role === ConstsUserRole.UserRoleAdmin;
  const [profiles, setProfiles] = useState<ReportProfile[]>([]);
  const [selected, setSelected] = useState<ReportProfile>();
  const [draft, setDraft] = useState<ReportProfilePayload>();
  const [search, setSearch] = useState('');
  const [edition, setEdition] = useState<'all' | EditionID>('all');
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const load = async () => {
    setLoading(true);
    try {
      setProfiles(await listAdminReportProfiles());
      setError('');
    } catch {
      setError('Unable to load report profiles.');
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    if (!user.id) return;
    if (!isAdmin) {
      navigate('/setting', { replace: true });
      return;
    }
    void load();
  }, [user.id, isAdmin]);
  const filtered = useMemo(
    () =>
      profiles.filter(
        profile =>
          (edition === 'all' || profile.edition_ids.includes(edition)) &&
          `${profile.name} ${profile.description}`
            .toLowerCase()
            .includes(search.toLowerCase()),
      ),
    [profiles, edition, search],
  );
  const startEdit = (profile: ReportProfile) => {
    setSelected(profile);
    setDraft(clone(payloadOf(profile)));
    setError('');
  };
  const valid =
    !!draft &&
    draft.name.trim() !== '' &&
    draft.system_prompt.length <= 16000 &&
    draft.edition_ids.length > 0 &&
    draft.input_fields.every(
      field =>
        field.key.trim() !== '' &&
        (field.type !== 'select' || field.options.length > 0),
    ) &&
    draft.sections.length > 0 &&
    new Set(draft.input_fields.map(field => field.key)).size ===
      draft.input_fields.length &&
    new Set(draft.sections.map(section => section.key)).size ===
      draft.sections.length &&
    new Set(draft.sections.map(section => section.order)).size ===
      draft.sections.length;
  const dirty =
    !!draft &&
    (!selected ||
      JSON.stringify(draft) !== JSON.stringify(payloadOf(selected)));
  const save = async () => {
    if (!draft || !valid || !dirty || saving) return;
    setSaving(true);
    try {
      const saved = selected
        ? await updateAdminReportProfile(selected.id, draft)
        : await createAdminReportProfile(draft);
      await load();
      setSelected(saved);
      setDraft(clone(payloadOf(saved)));
    } catch {
      setError('Save failed. Check required fields and try again.');
    } finally {
      setSaving(false);
    }
  };
  const toggle = async (profile: ReportProfile) => {
    try {
      await setAdminReportProfileEnabled(profile.id, !profile.enabled);
      await load();
    } catch {
      setError('Unable to update enabled state.');
    }
  };
  const remove = async (profile: ReportProfile) => {
    if (!window.confirm(`Delete ${profile.name}?`)) return;
    try {
      await deleteAdminReportProfile(profile.id);
      setSelected(undefined);
      setDraft(undefined);
      await load();
    } catch {
      setError('Only custom profiles can be deleted.');
    }
  };
  const restore = async (profile: ReportProfile) => {
    if (
      !window.confirm(
        'Restore the current built-in default? Your template edits will be replaced.',
      )
    )
      return;
    try {
      const restored = await restoreAdminReportProfileDefault(profile.id);
      await load();
      setSelected(restored);
      setDraft(clone(payloadOf(restored)));
    } catch {
      setError('Unable to restore this profile.');
    }
  };
  if (!user.id || !isAdmin) return <></>;
  if (loading) return <Card>Loading...</Card>;
  return (
    <Stack spacing={2}>
      <Typography variant='h5'>Report templates</Typography>
      {error && <Alert severity='error'>{error}</Alert>}
      <Stack direction='row' spacing={1}>
        <TextField
          label='Search'
          value={search}
          onChange={event => setSearch(event.target.value)}
        />
        <TextField
          select
          label='Edition'
          value={edition}
          onChange={event =>
            setEdition(event.target.value as 'all' | EditionID)
          }
        >
          <MenuItem value='all'>All</MenuItem>
          {editionIDs.map(item => (
            <MenuItem key={item} value={item}>
              {item}
            </MenuItem>
          ))}
        </TextField>
        <Button
          variant='contained'
          onClick={() => {
            setSelected(undefined);
            setDraft(blankProfile());
          }}
        >
          New custom template
        </Button>
      </Stack>
      <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2}>
        <Stack sx={{ minWidth: 360, flex: 1 }} spacing={1}>
          {filtered.map(profile => (
            <Card key={profile.id}>
              <Stack spacing={1}>
                <Typography>{profile.name}</Typography>
                <Typography variant='body2'>{profile.description}</Typography>
                <Stack direction='row' spacing={1} flexWrap='wrap'>
                  {profile.edition_ids.map(item => (
                    <Chip size='small' key={item} label={item} />
                  ))}
                  <Chip
                    size='small'
                    label={profile.is_builtin ? 'Built-in' : 'Custom'}
                  />
                </Stack>
                <Typography variant='caption'>
                  v{profile.version} ·{' '}
                  {profile.citation_required
                    ? 'Citations required'
                    : 'Citations optional'}{' '}
                  · {new Date(profile.updated_at).toLocaleString()}
                </Typography>
                <Stack direction='row' spacing={1}>
                  <Button onClick={() => startEdit(profile)}>Edit</Button>
                  <Button onClick={() => void toggle(profile)}>
                    {profile.enabled ? 'Disable' : 'Enable'}
                  </Button>
                  {profile.is_builtin ? (
                    <Button
                      color='warning'
                      onClick={() => void restore(profile)}
                    >
                      Restore default
                    </Button>
                  ) : (
                    <Button color='error' onClick={() => void remove(profile)}>
                      Delete
                    </Button>
                  )}
                </Stack>
              </Stack>
            </Card>
          ))}
        </Stack>
        <Box sx={{ flex: 2 }}>
          {draft && (
            <Card>
              <Stack spacing={2}>
                <Typography variant='h6'>
                  {selected ? `Edit: ${selected.name}` : 'New custom template'}
                </Typography>
                <ProfileEditor
                  value={draft}
                  onChange={setDraft}
                  builtin={!!selected?.is_builtin}
                  saving={saving}
                />
                <Stack direction='row' spacing={1}>
                  <Button
                    variant='contained'
                    disabled={!dirty || !valid || saving}
                    onClick={() => void save()}
                  >
                    Save
                  </Button>
                  <Button
                    disabled={!dirty || saving}
                    onClick={() =>
                      setDraft(
                        selected ? clone(payloadOf(selected)) : blankProfile(),
                      )
                    }
                  >
                    Discard unsaved changes
                  </Button>
                </Stack>
              </Stack>
            </Card>
          )}
        </Box>
      </Stack>
    </Stack>
  );
};
export default ReportProfiles;
