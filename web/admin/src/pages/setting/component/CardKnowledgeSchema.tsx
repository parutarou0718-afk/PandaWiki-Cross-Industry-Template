import { getKnowledgeSchema, putKnowledgeSchema, rebuildKnowledgeGraph } from '@/api/KnowledgeSchema';
import { useAppSelector } from '@/store';
import { message } from '@ctzhian/ui';
import { Add, Delete } from '@mui/icons-material';
import { Box, Button, Checkbox, Divider, FormControlLabel, MenuItem, Select, Stack, TextField, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { ENTITY_TYPES, type KnowledgeSchema, type KnowledgeSchemaField, type KnowledgeSchemaNavigation, validateKnowledgeSchema } from './knowledge-schema';
import { canLoadKnowledgeSchema, knowledgeSchemaFieldRenderKey } from './knowledge-schema-editor';

const blankField = (): KnowledgeSchemaField => ({ key: '', label: '', target: 'entity', entity_types: ['concept'], value_type: 'text', multiple: false, filterable: true, enabled: true, options: [], extract_instruction: '' });
const blankSection = (order: number): KnowledgeSchemaNavigation => ({ id: '', label: '', entity_types: ['concept'], field_keys: [], order, enabled: true });

const CardKnowledgeSchema = () => {
  const { kb_id } = useAppSelector(state => state.config);
  const [schema, setSchema] = useState<KnowledgeSchema | null>(null);
  const [saving, setSaving] = useState(false);
  const [rebuilding, setRebuilding] = useState(false);
  const load = () => { if (canLoadKnowledgeSchema(kb_id)) getKnowledgeSchema(kb_id).then(setSchema).catch(() => setSchema(null)); };
  useEffect(() => {
    setSchema(null);
    load();
    // The current knowledge-base ID is the only value that should reload the schema.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [kb_id]);
  if (!kb_id || !schema) return <Box p={3}>Loading knowledge model…</Box>;
  const updateField = (index: number, patch: Partial<KnowledgeSchemaField>) => setSchema(current => current && ({ ...current, fields: current.fields.map((field, i) => i === index ? { ...field, ...patch } : field) }));
  const updateSection = (index: number, patch: Partial<KnowledgeSchemaNavigation>) => setSchema(current => current && ({ ...current, navigation: current.navigation.map((section, i) => i === index ? { ...section, ...patch } : section) }));
  const save = () => {
    const validation = validateKnowledgeSchema(schema);
    if (validation) return message.error(validation);
    setSaving(true);
    putKnowledgeSchema(kb_id, schema).then(next => { setSchema(next); message.success('Knowledge model saved. New graph extraction will use it.'); }).finally(() => setSaving(false));
  };
  return <Box p={3} maxWidth={980}>
    <Typography variant='h6'>Knowledge model</Typography>
    <Typography color='text.secondary' fontSize={14} mt={1} mb={3}>Server-defined fields guide LLM extraction and the desktop Knowledge navigation.</Typography>
    <Stack gap={2}>{schema.fields.map((field, index) => <Box key={knowledgeSchemaFieldRenderKey(index)} border='1px solid' borderColor='divider' borderRadius={1} p={2}>
      <Stack direction='row' justifyContent='space-between' alignItems='center' mb={2}><Typography fontWeight={600}>Field {index + 1}</Typography><Button color='error' size='small' startIcon={<Delete />} onClick={() => setSchema(current => current && ({ ...current, fields: current.fields.filter((_, i) => i !== index) }))}>Remove</Button></Stack>
      <Stack direction='row' gap={2} flexWrap='wrap'>
        <TextField label='Key' value={field.key} onChange={event => updateField(index, { key: event.target.value })} helperText='lowercase_with_underscores' sx={{ minWidth: 180 }} />
        <TextField label='Display name' value={field.label} onChange={event => updateField(index, { label: event.target.value })} sx={{ minWidth: 180 }} />
        <Select value={field.value_type} onChange={event => updateField(index, { value_type: event.target.value as KnowledgeSchemaField['value_type'] })} size='small' sx={{ minWidth: 130 }}>{['text', 'number', 'date', 'boolean', 'select'].map(value => <MenuItem key={value} value={value}>{value}</MenuItem>)}</Select>
        <Select multiple value={field.entity_types} onChange={event => updateField(index, { entity_types: event.target.value as string[] })} size='small' sx={{ minWidth: 210 }}>{ENTITY_TYPES.map(type => <MenuItem key={type} value={type}>{type}</MenuItem>)}</Select>
      </Stack>
      {field.value_type === 'select' && <TextField fullWidth sx={{ mt: 2 }} label='Options (comma separated)' value={field.options.join(', ')} onChange={event => updateField(index, { options: event.target.value.split(',').map(value => value.trim()).filter(Boolean) })} />}
      <TextField fullWidth multiline minRows={2} sx={{ mt: 2 }} label='LLM extraction instruction' value={field.extract_instruction} onChange={event => updateField(index, { extract_instruction: event.target.value })} />
      <FormControlLabel control={<Checkbox checked={field.multiple} onChange={event => updateField(index, { multiple: event.target.checked })} />} label='Allow multiple values' /><FormControlLabel control={<Checkbox checked={field.filterable} onChange={event => updateField(index, { filterable: event.target.checked })} />} label='Filterable' /><FormControlLabel control={<Checkbox checked={field.enabled} onChange={event => updateField(index, { enabled: event.target.checked })} />} label='Enabled' />
    </Box>)}<Button startIcon={<Add />} variant='outlined' onClick={() => setSchema(current => current && ({ ...current, fields: [...current.fields, blankField()] }))}>Add field</Button>
    <Divider /><Typography variant='subtitle1' fontWeight={600}>Knowledge navigation</Typography>
    {schema.navigation.map((section, index) => <Stack key={`${section.id}-${index}`} direction='row' gap={2} alignItems='center' flexWrap='wrap'><TextField label='ID' value={section.id} onChange={event => updateSection(index, { id: event.target.value })} /><TextField label='Label' value={section.label} onChange={event => updateSection(index, { label: event.target.value })} /><Select multiple value={section.entity_types} onChange={event => updateSection(index, { entity_types: event.target.value as string[] })} size='small' sx={{ minWidth: 210 }}>{ENTITY_TYPES.map(type => <MenuItem key={type} value={type}>{type}</MenuItem>)}</Select><TextField label='Order' type='number' value={section.order} onChange={event => updateSection(index, { order: Number(event.target.value) })} sx={{ width: 100 }} /><FormControlLabel control={<Checkbox checked={section.enabled} onChange={event => updateSection(index, { enabled: event.target.checked })} />} label='Enabled' /><Button color='error' size='small' onClick={() => setSchema(current => current && ({ ...current, navigation: current.navigation.filter((_, i) => i !== index) }))}>Remove</Button></Stack>)}
    <Button startIcon={<Add />} variant='outlined' onClick={() => setSchema(current => current && ({ ...current, navigation: [...current.navigation, blankSection(current.navigation.length + 1)] }))}>Add navigation group</Button>
    <Stack direction='row' gap={2} mt={2}><Button variant='contained' onClick={save} disabled={saving}>{saving ? 'Saving…' : 'Save knowledge model'}</Button><Button variant='outlined' onClick={load}>Discard changes</Button><Button variant='outlined' disabled={rebuilding} onClick={() => { setRebuilding(true); rebuildKnowledgeGraph(kb_id).then(result => message.success(`Queued ${result.queued} documents for graph extraction.`)).finally(() => setRebuilding(false)); }}>{rebuilding ? 'Queueing…' : 'Rebuild graph from documents'}</Button></Stack>
    </Stack>
  </Box>;
};
export default CardKnowledgeSchema;
