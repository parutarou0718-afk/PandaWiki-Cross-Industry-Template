import {
  Alert,
  Box,
  Button,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Card from '@/components/Card';
import {
  getEdition,
  updateEdition,
  EditionConfig,
  EditionOverrides,
} from '@/request/Edition';
import { ConstsUserRole } from '@/request/types';
import { useAppSelector } from '@/store';

const editionIDs = ['common', 'research', 'legal', 'finance'] as const;
const featureIDs = ['chat', 'search'];
const terminologyKeys = [
  'knowledge_base',
  'document',
  'folder',
  'report',
] as const;

const clone = (value: EditionConfig) =>
  JSON.parse(JSON.stringify(value)) as EditionConfig;

const getOverrides = (
  base: EditionConfig,
  value: EditionConfig,
): EditionOverrides => {
  const overrides: EditionOverrides = {};
  (
    [
      'product_name',
      'short_name',
      'branding',
      'home_description',
      'terminology',
      'enabled_features',
      'document_types',
      'relation_types',
      'default_prompts',
    ] as const
  ).forEach(key => {
    if (JSON.stringify(base[key]) !== JSON.stringify(value[key]))
      overrides[key] = value[key] as never;
  });
  return overrides;
};

const Edition = () => {
  const navigate = useNavigate();
  const user = useAppSelector(state => state.config.user);
  const [loaded, setLoaded] = useState<EditionConfig>();
  const [draft, setDraft] = useState<EditionConfig>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const isAdmin = user.role === ConstsUserRole.UserRoleAdmin;
  const dirty = useMemo(
    () =>
      !!loaded && !!draft && JSON.stringify(loaded) !== JSON.stringify(draft),
    [loaded, draft],
  );

  const load = async () => {
    setLoading(true);
    try {
      const value = await getEdition();
      setLoaded(value);
      setDraft(clone(value));
      setError('');
    } catch {
      setError('加载行业版本配置失败，请稍后重试。');
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
  }, [isAdmin, user.id]);
  if (!user.id) return <Card>加载中...</Card>;
  if (!isAdmin) return <></>;
  if (loading || !loaded || !draft) return <Card>加载中...</Card>;

  const set = <K extends keyof EditionConfig>(
    key: K,
    value: EditionConfig[K],
  ) => setDraft({ ...draft, [key]: value });
  const save = async (overrides = getOverrides(loaded, draft)) => {
    if (saving) return;
    if (
      draft.edition_id !== loaded.edition_id &&
      !window.confirm(
        '切换行业版本不会删除知识库、文档、权限或向量数据，是否继续？',
      )
    )
      return;
    setSaving(true);
    setError('');
    try {
      // A different preset must never inherit the previous edition's resolved fields as overrides.
      await updateEdition(
        draft.edition_id,
        draft.edition_id === loaded.edition_id ? overrides : {},
      );
      await load();
    } catch {
      setError('保存失败：请检查字段内容或稍后重试。');
    } finally {
      setSaving(false);
    }
  };
  const restoreDefaults = async () => {
    if (
      !window.confirm(
        '将清空当前行业版本的所有管理员覆盖配置，并恢复预置默认值。是否继续？',
      )
    )
      return;
    setSaving(true);
    setError('');
    try {
      await updateEdition(loaded.edition_id, {});
      await load();
    } catch {
      setError('恢复版本默认值失败，请稍后重试。');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack spacing={2}>
      <Typography variant='h5'>行业版本</Typography>
      <Alert severity='info'>
        默认 Prompt 当前仅保存配置，问答和摘要运行时接入将在后续阶段完成。
      </Alert>
      {error && <Alert severity='error'>{error}</Alert>}
      <TextField
        select
        label='当前行业版本'
        value={draft.edition_id}
        disabled={saving}
        helperText={
          draft.edition_id !== loaded.edition_id
            ? '切换版本时将使用目标版本的预置默认值；请保存后再编辑其覆盖配置。'
            : ''
        }
        onChange={event =>
          set('edition_id', event.target.value as EditionConfig['edition_id'])
        }
      >
        {editionIDs.map(id => (
          <MenuItem key={id} value={id}>
            {id}
          </MenuItem>
        ))}
      </TextField>
      {(['product_name', 'short_name', 'home_description'] as const).map(
        key => (
          <TextField
            key={key}
            label={key}
            value={draft[key]}
            disabled={saving}
            onChange={event => set(key, event.target.value)}
          />
        ),
      )}
      <TextField
        label='Logo'
        value={draft.branding.logo || ''}
        disabled={saving}
        onChange={event =>
          set('branding', { ...draft.branding, logo: event.target.value })
        }
      />
      {terminologyKeys.map(key => (
        <TextField
          key={key}
          label={`术语：${key}`}
          value={draft.terminology[key] || ''}
          disabled={saving}
          onChange={event =>
            set('terminology', {
              ...draft.terminology,
              [key]: event.target.value,
            })
          }
        />
      ))}
      <TextField
        label='启用功能'
        value={draft.enabled_features.join(', ')}
        disabled={saving}
        helperText={`仅允许：${featureIDs.join(', ')}`}
        onChange={event =>
          set(
            'enabled_features',
            event.target.value
              .split(',')
              .map(value => value.trim())
              .filter(value => featureIDs.includes(value)),
          )
        }
      />
      <TextField
        label='文档类型（每行一项）'
        multiline
        minRows={3}
        value={draft.document_types.join('\n')}
        disabled={saving}
        onChange={event =>
          set('document_types', event.target.value.split('\n').filter(Boolean))
        }
      />
      <TextField
        label='关系类型（每行一项）'
        multiline
        minRows={3}
        value={draft.relation_types.join('\n')}
        disabled={saving}
        onChange={event =>
          set('relation_types', event.target.value.split('\n').filter(Boolean))
        }
      />
      <TextField
        label='问答 Prompt'
        multiline
        minRows={4}
        inputProps={{ maxLength: 16000 }}
        value={draft.default_prompts.chat}
        disabled={saving}
        helperText={`${draft.default_prompts.chat.length}/16000`}
        onChange={event =>
          set('default_prompts', {
            ...draft.default_prompts,
            chat: event.target.value,
          })
        }
      />
      <TextField
        label='摘要 Prompt'
        multiline
        minRows={4}
        inputProps={{ maxLength: 16000 }}
        value={draft.default_prompts.summary}
        disabled={saving}
        helperText={`${draft.default_prompts.summary.length}/16000`}
        onChange={event =>
          set('default_prompts', {
            ...draft.default_prompts,
            summary: event.target.value,
          })
        }
      />
      <Box>
        <Button
          variant='contained'
          disabled={!dirty || saving}
          onClick={() => void save()}
        >
          保存
        </Button>
        <Button
          sx={{ ml: 1 }}
          disabled={!dirty || saving}
          onClick={() => setDraft(clone(loaded))}
        >
          撤销未保存修改
        </Button>
        <Button
          sx={{ ml: 1 }}
          color='warning'
          disabled={saving}
          onClick={() => void restoreDefaults()}
        >
          恢复版本默认值
        </Button>
      </Box>
    </Stack>
  );
};
export default Edition;
