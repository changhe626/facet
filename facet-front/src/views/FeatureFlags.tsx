import { useState, useEffect, useCallback } from 'react';
import { useOutletContext } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { useLocaleStore } from '@/stores/localeStore';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TablePagination from '@mui/material/TablePagination';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import VisibilityIcon from '@mui/icons-material/Visibility';
import HistoryIcon from '@mui/icons-material/History';
import SearchIcon from '@mui/icons-material/Search';
import Slider from '@mui/material/Slider';
import Autocomplete from '@mui/material/Autocomplete';
import Checkbox from '@mui/material/Checkbox';
import ArchiveIcon from '@mui/icons-material/Archive';
import FlagIcon from '@mui/icons-material/Flag';
import UnarchiveIcon from '@mui/icons-material/Unarchive';

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { NEVER_EXPIRES } from '@/utils/constants';
import { getFeatureFlagsApi, getFeatureFlagDetailsApi, createFeatureFlagApi, updateFeatureFlagApi, deleteFeatureFlagApi, getFeatureFlagHistoryApi, archiveFeatureFlagApi, unarchiveFeatureFlagApi } from '@/api/featureFlags';
import { getSegmentsApi } from '@/api/segments';
import { formatDateTime, formatFlagType } from '@/utils/formatters';
import type { FeatureFlag, FeatureFlagDetail, FeatureFlagHistory, WhitelistRule, Segment, CreateFeatureFlagRequest, UpdateFeatureFlagRequest } from '@/types';

const flagTypes = ['RELEASE', 'EXPERIMENT', 'OPERATIONAL', 'PERMISSION'];

interface WhitelistRuleForm {
  accountId: string;
  userIds: string[];
}

interface FeatureFlagForm {
  name: string;
  description: string;
  flagType: string;
  enabled: boolean;
  rolloutPercentage: number;
  expiresAt: string;
  isPermanent: boolean;
  whitelistRules: WhitelistRuleForm[];
  segmentIds: string[];
}

const defaultForm: FeatureFlagForm = {
  name: '', description: '', flagType: 'RELEASE', enabled: true,
  rolloutPercentage: 0, expiresAt: 'never', isPermanent: true,
  whitelistRules: [{ accountId: '', userIds: [''] }], segmentIds: [],
};

export default function FeatureFlags() {
  const { currentEnvId } = useOutletContext<{ currentEnvId: string | null }>();
  const currentUser = useAuthStore((s) => s.user);
  const userCanWrite = currentUser?.role && currentUser.role !== 'tenant_viewer';
  const { t } = useLocaleStore();

  const [data, setData] = useState<FeatureFlag[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [includeArchived, setIncludeArchived] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailData, setDetailData] = useState<FeatureFlagDetail | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyData, setHistoryData] = useState<FeatureFlagHistory[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FeatureFlagForm>(defaultForm);
  const [segmentOptions, setSegmentOptions] = useState<Segment[]>([]);
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; onConfirm: () => void } | null>(null);

  const fetchData = useCallback(async () => {
    if (!currentEnvId) {
      setData([]);
      setTotal(0);
      return;
    }
    setLoading(true);
    setError('');
    try {
      const params: Record<string, unknown> = { page: page + 1, pageSize: rowsPerPage };
      if (keyword) params.keyword = keyword;
      if (includeArchived) params.includeArchived = true;
      const res = await getFeatureFlagsApi(currentEnvId, params);
      setData(res.data?.items || []);
      setTotal(res.data?.total || 0);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [currentEnvId, page, rowsPerPage, keyword, includeArchived]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    if (!currentEnvId) { setData([]); setTotal(0); }
    setPage(0);
    setKeyword('');
  }, [currentEnvId]);

  const handleSearch = () => { setPage(0); fetchData(); };

  const openCreate = () => {
    setEditingId(null);
    setForm({ ...defaultForm });
    setFormError('');
    setDialogOpen(true);
  };

  const openEdit = async (row: FeatureFlag) => {
    setEditingId(row.id);
    try {
      const res = await getFeatureFlagDetailsApi(currentEnvId!, row.id);
      const ff = res.data;
      const isPermanent = ff.expiresAt === NEVER_EXPIRES || !ff.expiresAt;
      if (segmentOptions.length === 0) {
        const segRes = await getSegmentsApi({ page: 1, pageSize: 100 });
        setSegmentOptions(segRes.data?.items || []);
      }
      setForm({
        name: ff.name,
        description: ff.description || '',
        flagType: ff.flagType || 'RELEASE',
        enabled: ff.enabled ?? true,
        rolloutPercentage: ff.rolloutPercentage ?? 0,
        expiresAt: ff.expiresAt || NEVER_EXPIRES,
        isPermanent,
        whitelistRules: (ff.whitelistRules && ff.whitelistRules.length > 0)
          ? ff.whitelistRules.map((r: WhitelistRule) => ({ accountId: r.accountId || '', userIds: (r.userIDs && r.userIDs.length > 0) ? r.userIDs : [''] }))
          : [{ accountId: '', userIds: [''] }],
        segmentIds: ff.segmentIds || [],
      });
    } catch (e) {
      setError(getErrorMessage(e));
      return;
    }
    setFormError('');
    setDialogOpen(true);
  };

  const handlePermanentChange = (perm: boolean) => {
    setForm({ ...form, isPermanent: perm, expiresAt: perm ? NEVER_EXPIRES : '' });
  };

  const updateRule = (idx: number, field: string, value: string) => {
    const rules = [...form.whitelistRules];
    rules[idx] = { ...rules[idx], [field]: value };
    setForm({ ...form, whitelistRules: rules });
  };

  const addRule = () => {
    setForm({ ...form, whitelistRules: [...form.whitelistRules, { accountId: '', userIds: [''] }] });
  };

  const removeRule = (idx: number) => {
    const rules = form.whitelistRules.filter((_, i) => i !== idx);
    setForm({ ...form, whitelistRules: rules.length > 0 ? rules : [{ accountId: '', userIds: [''] }] });
  };

  const updateUserIds = (idx: number, val: string) => {
    const ids = val.split(',').map((s) => s.trim()).filter(Boolean);
    const rules = [...form.whitelistRules];
    rules[idx] = { ...rules[idx], userIds: ids.length > 0 ? ids : [''] };
    setForm({ ...form, whitelistRules: rules });
  };

  const handleSave = async () => {
    if (!form.name.trim()) { setFormError(t.common.nameRequired); return; }
    setSaving(true);
    setFormError('');
    const basePayload = {
      description: form.description,
      flagType: form.flagType,
      enabled: form.enabled,
      rolloutPercentage: form.rolloutPercentage,
      expiresAt: form.isPermanent ? NEVER_EXPIRES : form.expiresAt,
      whitelistRules: form.whitelistRules.map((r) => ({
        accountId: r.accountId,
        userIds: r.userIds.filter((u) => u.trim() !== ''),
      })),
      segmentIds: form.segmentIds,
    };
    try {
      if (editingId) {
        await updateFeatureFlagApi(currentEnvId!, editingId, basePayload as UpdateFeatureFlagRequest);
      } else {
        await createFeatureFlagApi(currentEnvId!, { name: form.name, ...basePayload } as CreateFeatureFlagRequest);
      }
      setDialogOpen(false);
      fetchData();
    } catch (e) {
      setFormError(getErrorMessage(e));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = (id: string) => {
    setConfirmDialog({
      message: t.featureFlags.deleteConfirm,
      onConfirm: async () => {
        try {
          await deleteFeatureFlagApi(currentEnvId!, id);
          fetchData();
        } catch (e) {
          setError(getErrorMessage(e));
        }
      },
    });
  };

  const handleArchive = async (id: string) => {
    try {
      await archiveFeatureFlagApi(currentEnvId!, id);
      fetchData();
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  const handleUnarchive = async (id: string) => {
    try {
      await unarchiveFeatureFlagApi(currentEnvId!, id);
      fetchData();
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  const handleViewDetails = async (row: FeatureFlag) => {
    try {
      const res = await getFeatureFlagDetailsApi(currentEnvId!, row.id);
      setDetailData(res.data);
      setDetailOpen(true);
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  const handleViewHistory = async (row: FeatureFlag) => {
    try {
      const res = await getFeatureFlagHistoryApi(currentEnvId!, row.id);
      setHistoryData(res.data || []);
      setHistoryOpen(true);
    } catch (e) {
      setError(getErrorMessage(e));
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.featureFlags.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} disabled={!currentEnvId || !userCanWrite}>{t.featureFlags.create}</Button>
      </Box>

      {!currentEnvId && (
        <Alert severity="info" sx={{ mb: 2 }}>{t.featureFlags.noEnv}</Alert>
      )}
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      {currentEnvId && (
        <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
          <TextField
            size="small"
            placeholder={t.featureFlags.searchPlaceholder}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => (e as React.KeyboardEvent).key === 'Enter' && handleSearch()}
            sx={{ minWidth: 240 }}
          />
          <FormControlLabel control={<Checkbox size="small" checked={includeArchived} onChange={(e) => setIncludeArchived(e.target.checked)} />} label={t.featureFlags.showArchived} />
          <Button variant="outlined" startIcon={<SearchIcon />} onClick={handleSearch}>{t.common.search}</Button>
        </Paper>
      )}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.common.name}</TableCell>
              <TableCell>{t.common.type}</TableCell>
              <TableCell>{t.common.status}</TableCell>
              <TableCell>{t.featureFlags.rollout}</TableCell>
              <TableCell>{t.featureFlags.expires}</TableCell>
              <TableCell>{t.common.updatedAt}</TableCell>
              <TableCell align="right">{t.common.actions}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={7} align="center"><CircularProgress /></TableCell></TableRow>
            ) : data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: 8 }}>
                  <Box>
                    <Box
                      sx={{
                        width: 80,
                        height: 80,
                        borderRadius: '50%',
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: (t) =>
                          t.palette.mode === 'dark'
                            ? 'linear-gradient(135deg, rgba(92,107,192,0.2) 0%, rgba(38,166,154,0.2) 100%)'
                            : 'linear-gradient(135deg, #e8eaf6 0%, #e0f2f1 100%)',
                        mb: 2,
                      }}
                    >
                      <FlagIcon sx={{ fontSize: 40, color: 'primary.main' }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 0.5 }}>
                      {t.featureFlags.emptyTitle}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5, maxWidth: 360, mx: 'auto' }}>
                      {t.featureFlags.emptyDesc}
                    </Typography>
                    {userCanWrite && (
                      <Button
                        variant="contained"
                        onClick={openCreate}
                        sx={{ fontWeight: 700, letterSpacing: '0.5px', px: 3 }}
                      >
                        {t.featureFlags.emptyAction}
                      </Button>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id} sx={{ opacity: row.archivedAt ? 0.6 : 1 }}>
                <TableCell sx={{ fontWeight: 600 }}>
                  {row.name}
                  {row.archivedAt && <Chip label={t.featureFlags.archived} size="small" variant="outlined" color="default" sx={{ ml: 1 }} />}
                </TableCell>
                <TableCell>{formatFlagType(row.flagType)}</TableCell>
                <TableCell>
                  <Chip
                    label={row.enabled ? t.common.enabled : t.common.disabled}
                    color={row.enabled ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  {(row.rolloutPercentage ?? 0) > 0 ? `${row.rolloutPercentage}%` : '-'}
                </TableCell>
                <TableCell>{row.expiresAt === NEVER_EXPIRES ? t.common.never : formatDateTime(row.expiresAt)}</TableCell>
                <TableCell>{formatDateTime(row.updatedAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.viewDetails}><IconButton aria-label={t.common.viewDetails} size="small" onClick={() => handleViewDetails(row)}><VisibilityIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.history}><IconButton aria-label={t.common.history} size="small" onClick={() => handleViewHistory(row)}><HistoryIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.edit}><span><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)} disabled={!userCanWrite}><EditIcon /></IconButton></span></Tooltip>
                  {row.archivedAt ? (
                    <Tooltip title={t.common.unarchive}><span><IconButton aria-label={t.common.unarchive} size="small" onClick={() => handleUnarchive(row.id)} disabled={!userCanWrite}><UnarchiveIcon /></IconButton></span></Tooltip>
                  ) : (
                    <Tooltip title={t.common.archive}><span><IconButton aria-label={t.common.archive} size="small" onClick={() => handleArchive(row.id)} disabled={!userCanWrite}><ArchiveIcon /></IconButton></span></Tooltip>
                  )}
                  <Tooltip title={t.common.delete}><span><IconButton aria-label={t.common.delete} size="small" color="error" onClick={() => handleDelete(row.id)} disabled={!userCanWrite}><DeleteIcon /></IconButton></span></Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination
          component="div"
          count={total}
          page={page}
          onPageChange={(_, p) => setPage(p)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }}
          rowsPerPageOptions={[10, 20, 50]}
        />
      </TableContainer>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editingId ? t.featureFlags.editFlag : t.featureFlags.createFlag}</DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField autoFocus fullWidth label={t.featureFlags.fieldName} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} sx={{ mt: 2, mb: 2 }} disabled={!!editingId} />
          <TextField fullWidth label={t.featureFlags.fieldDescription} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} multiline rows={2} sx={{ mb: 2 }} />
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>{t.featureFlags.fieldType}</InputLabel>
            <Select value={form.flagType} label={t.featureFlags.fieldType} onChange={(e) => setForm({ ...form, flagType: e.target.value })}>
              {flagTypes.map((t) => <MenuItem key={t} value={t}>{formatFlagType(t)}</MenuItem>)}
            </Select>
          </FormControl>
          <FormControlLabel control={<Switch checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: (e.target as HTMLInputElement).checked })} />} label={t.featureFlags.fieldEnabled} sx={{ mb: 2 }} />

          <Typography gutterBottom>{t.featureFlags.rollout}: {form.rolloutPercentage}%</Typography>
          <Slider
            value={form.rolloutPercentage}
            onChange={(_, v) => setForm({ ...form, rolloutPercentage: Array.isArray(v) ? v[0] : v })}
            valueLabelDisplay="auto"
            step={5}
            marks={[{ value: 0, label: '0%' }, { value: 50, label: '50%' }, { value: 100, label: '100%' }]}
            min={0}
            max={100}
            sx={{ mb: 2, mx: 1 }}
          />

          <Autocomplete
            multiple
            options={segmentOptions}
            getOptionLabel={(o: Segment) => o.name}
            value={segmentOptions.filter((o) => form.segmentIds.includes(o.id))}
            onChange={(_, v) => setForm({ ...form, segmentIds: v.map((s: Segment) => s.id) })}
            onOpen={() => {
              if (segmentOptions.length === 0) {
                getSegmentsApi({ page: 1, pageSize: 100 }).then((res) => setSegmentOptions(res.data?.items || [])).catch(() => {});
              }
            }}
            renderInput={(params) => <TextField {...params} label={t.featureFlags.fieldSegments} placeholder={t.featureFlags.fieldSegments} />}
            sx={{ mb: 2 }}
          />

          <Box sx={{ mb: 2 }}>
            <FormControlLabel control={<Switch checked={form.isPermanent} onChange={(e) => handlePermanentChange((e.target as HTMLInputElement).checked)} />} label={t.featureFlags.fieldNeverExpires} />
          </Box>
          {!form.isPermanent && (
            <TextField fullWidth label={t.featureFlags.fieldExpiresAt} type="datetime-local" value={form.expiresAt} onChange={(e) => setForm({ ...form, expiresAt: e.target.value })} sx={{ mb: 2 }} slotProps={{ inputLabel: { shrink: true } }} />
          )}

          <Typography variant="subtitle2" sx={{ mb: 1 }}>{t.featureFlags.whitelistRules}</Typography>
          {form.whitelistRules.map((rule, idx) => (
            <Box key={idx} sx={{ display: 'flex', gap: 1, mb: 1, alignItems: 'center' }}>
              <TextField size="small" label={t.featureFlags.fieldAccountId} value={rule.accountId} onChange={(e) => updateRule(idx, 'accountId', e.target.value)} sx={{ minWidth: 160 }} />
              <TextField size="small" label={t.featureFlags.fieldUserIds} value={rule.userIds.filter((u) => u !== '').join(', ')} onChange={(e) => updateUserIds(idx, e.target.value)} sx={{ flexGrow: 1 }} />
              <IconButton aria-label={t.featureFlags.removeRule} size="small" color="error" onClick={() => removeRule(idx)}><DeleteIcon /></IconButton>
            </Box>
          ))}
          <Button size="small" startIcon={<AddIcon />} onClick={addRule}>{t.featureFlags.addRule}</Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>{t.common.cancel}</Button>
          <Button variant="contained" onClick={handleSave} disabled={saving}>
            {saving ? <CircularProgress size={20} /> : t.common.save}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{t.featureFlags.flagDetails}</DialogTitle>
        <DialogContent>
          {detailData && (
            <Table size="small">
              <TableBody>
                {([
                  [t.featureFlags.fieldName, detailData.name],
                  [t.featureFlags.fieldDescription, detailData.description || '-'],
                  [t.featureFlags.fieldType, formatFlagType(detailData.flagType)],
                  [t.featureFlags.fieldEnabled, detailData.enabled ? t.featureFlags.yes : t.featureFlags.no],
                  [t.featureFlags.rollout, (detailData.rolloutPercentage ?? 0) > 0 ? `${detailData.rolloutPercentage}%` : t.common.disabled],
                  [t.featureFlags.expires, detailData.expiresAt === NEVER_EXPIRES ? t.common.never : formatDateTime(detailData.expiresAt)],
                  [t.common.createdAt, formatDateTime(detailData.createdAt)],
                  [t.common.updatedAt, formatDateTime(detailData.updatedAt)],
                ] as [string, string][]).map(([k, v]) => (
                  <TableRow key={k}>
                    <TableCell sx={{ fontWeight: 600, width: 140 }}>{k}</TableCell>
                    <TableCell>{v}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
          {detailData?.segmentIds && detailData.segmentIds.length > 0 && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>{t.featureFlags.segments}</Typography>
              <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                {detailData.segmentIds.map((sid: string) => (
                  <Chip key={sid} size="small" label={sid} variant="outlined" />
                ))}
              </Box>
            </Box>
          )}
          {detailData?.whitelistRules && detailData.whitelistRules.length > 0 && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>{t.featureFlags.whitelistRules}</Typography>
              {detailData.whitelistRules.map((r, i) => (
                <Box key={r.accountId || i} sx={{ display: 'flex', gap: 1, mb: 0.5 }}>
                  <Chip size="small" label={`${t.featureFlags.account}: ${r.accountId || '*'}`} />
                  <Chip size="small" label={`${t.featureFlags.users}: ${(r.userIDs && r.userIDs.length > 0) ? r.userIDs.join(', ') : '*'}`} variant="outlined" />
                </Box>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>{t.common.close}</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={historyOpen} onClose={() => setHistoryOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{t.featureFlags.changeHistory}</DialogTitle>
        <DialogContent>
          {historyData.length === 0 ? (
            <Typography color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>{t.featureFlags.noHistory}</Typography>
          ) : (
            <Box>
              {historyData.map((h) => (
                <Paper key={h.historyId} variant="outlined" sx={{ p: 2, mb: 1 }}>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    {formatDateTime(h.changeTime)}
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                    <Chip size="small" label={`${t.featureFlags.fieldName}: ${h.name}`} />
                    <Chip size="small" label={`${t.featureFlags.fieldEnabled}: ${h.enabled ? t.featureFlags.yes : t.featureFlags.no}`} color={h.enabled ? 'success' : 'default'} />
                    <Chip size="small" label={`${t.featureFlags.expires}: ${h.expiresAt === NEVER_EXPIRES ? t.common.never : h.expiresAt}`} variant="outlined" />
                  </Box>
                </Paper>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setHistoryOpen(false)}>{t.common.close}</Button>
        </DialogActions>
      </Dialog>

      <ConfirmDialog
        open={confirmDialog !== null}
        message={confirmDialog?.message ?? ''}
        onConfirm={() => {
          const action = confirmDialog?.onConfirm;
          setConfirmDialog(null);
          action?.();
        }}
        onCancel={() => setConfirmDialog(null)}
      />
    </Box>
  );
}
