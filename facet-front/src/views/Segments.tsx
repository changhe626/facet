import { useState, useEffect, useCallback } from 'react';
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
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import SearchIcon from '@mui/icons-material/Search';
import FlagIcon from '@mui/icons-material/Flag';

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { getSegmentsApi, createSegmentApi, updateSegmentApi, deleteSegmentApi } from '@/api/segments';
import { formatDateTime } from '@/utils/formatters';
import type { Segment } from '@/types';

const operators = [
  { value: 'equals', label: 'equals' },
  { value: 'not_equals', label: 'not equals' },
  { value: 'contains', label: 'contains' },
  { value: 'not_contains', label: 'not contains' },
  { value: 'starts_with', label: 'starts with' },
  { value: 'ends_with', label: 'ends with' },
  { value: 'in', label: 'in (comma-sep)' },
  { value: 'not_in', label: 'not in (comma-sep)' },
  { value: 'gt', label: 'greater than' },
  { value: 'gte', label: 'greater or equal' },
  { value: 'lt', label: 'less than' },
  { value: 'lte', label: 'less or equal' },
];

interface ConditionForm {
  attributeKey: string;
  operator: string;
  value: string;
}

interface SegmentForm {
  name: string;
  description: string;
  matchType: string;
  conditions: ConditionForm[];
}

export default function Segments() {
  const currentUser = useAuthStore((s) => s.user);
  const userCanWrite = currentUser?.role && currentUser.role !== 'tenant_viewer';
  const { t } = useLocaleStore();
  const [data, setData] = useState<Segment[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<SegmentForm>({ name: '', description: '', matchType: 'ANY', conditions: [{ attributeKey: '', operator: 'equals', value: '' }] });
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; onConfirm: () => void } | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params: Record<string, unknown> = { page: page + 1, pageSize: rowsPerPage };
      if (keyword) params.keyword = keyword;
      const res = await getSegmentsApi(params);
      setData(res.data?.items || []);
      setTotal(res.data?.total || 0);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [page, rowsPerPage, keyword]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSearch = () => { setPage(0); fetchData(); };

  const openCreate = () => {
    setEditingId(null);
    setForm({ name: '', description: '', matchType: 'ANY', conditions: [{ attributeKey: '', operator: 'equals', value: '' }] });
    setFormError('');
    setDialogOpen(true);
  };

  const openEdit = async (row: Segment) => {
    setEditingId(row.id);
    setForm({
      name: row.name,
      description: row.description || '',
      matchType: row.matchType || 'ANY',
      conditions: (row.conditions && row.conditions.length > 0)
        ? row.conditions.map((c) => ({ attributeKey: c.attributeKey, operator: c.operator, value: c.value }))
        : [{ attributeKey: '', operator: 'equals', value: '' }],
    });
    setFormError('');
    setDialogOpen(true);
  };

  const updateCondition = (idx: number, field: string, value: string) => {
    const conditions = [...form.conditions];
    conditions[idx] = { ...conditions[idx], [field]: value };
    setForm({ ...form, conditions });
  };

  const addCondition = () => {
    setForm({ ...form, conditions: [...form.conditions, { attributeKey: '', operator: 'equals', value: '' }] });
  };

  const removeCondition = (idx: number) => {
    const conditions = form.conditions.filter((_, i) => i !== idx);
    setForm({ ...form, conditions: conditions.length > 0 ? conditions : [{ attributeKey: '', operator: 'equals', value: '' }] });
  };

  const handleSave = async () => {
    if (!form.name.trim()) { setFormError(t.common.nameRequired); return; }
    setSaving(true);
    setFormError('');
    const payload = { name: form.name, description: form.description, matchType: form.matchType, conditions: form.conditions.filter((c) => c.attributeKey.trim()).map((c) => ({ attributeKey: c.attributeKey, operator: c.operator, value: c.value })) };
    try {
      if (editingId) {
        await updateSegmentApi(editingId, payload);
      } else {
        await createSegmentApi(payload);
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
      message: t.segments.deleteConfirm,
      onConfirm: async () => {
        try {
          await deleteSegmentApi(id);
          fetchData();
        } catch (e) {
          setError(getErrorMessage(e));
        }
      },
    });
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.segments.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} disabled={!userCanWrite}>{t.segments.create}</Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
        <TextField size="small" placeholder={t.segments.searchPlaceholder} value={keyword} onChange={(e) => setKeyword(e.target.value)} onKeyDown={(e) => (e as React.KeyboardEvent).key === 'Enter' && handleSearch()} sx={{ minWidth: 240 }} />
        <Button variant="outlined" startIcon={<SearchIcon />} onClick={handleSearch}>{t.common.search}</Button>
      </Paper>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.common.name}</TableCell>
              <TableCell>{t.segments.matchType}</TableCell>
              <TableCell>{t.segments.conditions}</TableCell>
              <TableCell>{t.common.updatedAt}</TableCell>
              <TableCell align="right">{t.common.actions}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={5} align="center"><CircularProgress /></TableCell></TableRow>
            ) : data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: 8 }}>
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
                            ? 'linear-gradient(135deg, rgba(255,112,67,0.2) 0%, rgba(171,71,188,0.2) 100%)'
                            : 'linear-gradient(135deg, #fbe9e7 0%, #f3e5f5 100%)',
                        mb: 2,
                      }}
                    >
                      <FlagIcon sx={{ fontSize: 40, color: 'secondary.main' }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 0.5 }}>
                      {t.segments.emptyTitle}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5, maxWidth: 360, mx: 'auto' }}>
                      {t.segments.emptyDesc}
                    </Typography>
                    {userCanWrite && (
                      <Button
                        variant="contained"
                        onClick={openCreate}
                        sx={{ fontWeight: 700, letterSpacing: '0.5px', px: 3 }}
                      >
                        {t.segments.emptyAction}
                      </Button>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id}>
                <TableCell sx={{ fontWeight: 600 }}>{row.name}</TableCell>
                <TableCell><Chip label={row.matchType === 'ALL' ? 'ALL' : 'ANY'} size="small" color={row.matchType === 'ALL' ? 'info' : 'default'} /></TableCell>
                <TableCell>{row.conditions?.length || 0} {t.segments.conditionCount}</TableCell>
                <TableCell>{formatDateTime(row.updatedAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.edit}><span><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)} disabled={!userCanWrite}><EditIcon /></IconButton></span></Tooltip>
                  <Tooltip title={t.common.delete}><span><IconButton aria-label={t.common.delete} size="small" color="error" onClick={() => handleDelete(row.id)} disabled={!userCanWrite}><DeleteIcon /></IconButton></span></Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination component="div" count={total} page={page} onPageChange={(_, p) => setPage(p)} rowsPerPage={rowsPerPage} onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }} rowsPerPageOptions={[10, 20, 50]} />
      </TableContainer>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editingId ? t.segments.editSegment : t.segments.createSegment}</DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField autoFocus fullWidth label={t.segments.fieldName} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} sx={{ mt: 2, mb: 2 }} />
          <TextField fullWidth label={t.segments.fieldDescription} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} multiline rows={2} sx={{ mb: 2 }} />
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>{t.segments.fieldMatchType}</InputLabel>
            <Select value={form.matchType} label={t.segments.fieldMatchType} onChange={(e) => setForm({ ...form, matchType: e.target.value })}>
              <MenuItem value="ANY">{t.segments.matchAny}</MenuItem>
              <MenuItem value="ALL">{t.segments.matchAll}</MenuItem>
            </Select>
          </FormControl>

          <Typography variant="subtitle2" sx={{ mb: 1 }}>{t.segments.conditions}</Typography>
          {form.conditions.map((cond, idx) => (
            <Box key={idx} sx={{ display: 'flex', gap: 1, mb: 1, alignItems: 'center' }}>
              <TextField size="small" label={t.segments.fieldAttributeKey} value={cond.attributeKey} onChange={(e) => updateCondition(idx, 'attributeKey', e.target.value)} sx={{ minWidth: 140 }} />
              <FormControl size="small" sx={{ minWidth: 140 }}>
                <InputLabel>{t.segments.fieldOperator}</InputLabel>
                <Select value={cond.operator} label={t.segments.fieldOperator} onChange={(e) => updateCondition(idx, 'operator', e.target.value)}>
                  {operators.map((o) => <MenuItem key={o.value} value={o.value}>{o.label}</MenuItem>)}
                </Select>
              </FormControl>
              <TextField size="small" label={t.segments.fieldValue} value={cond.value} onChange={(e) => updateCondition(idx, 'value', e.target.value)} sx={{ flexGrow: 1 }} />
              <IconButton aria-label={t.segments.removeCondition} size="small" color="error" onClick={() => removeCondition(idx)}><DeleteIcon /></IconButton>
            </Box>
          ))}
          <Button size="small" startIcon={<AddIcon />} onClick={addCondition}>{t.segments.addCondition}</Button>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>{t.common.cancel}</Button>
          <Button variant="contained" onClick={handleSave} disabled={saving}>
            {saving ? <CircularProgress size={20} /> : t.common.save}
          </Button>
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
