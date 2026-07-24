import { useState, useEffect, useCallback } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useLocaleStore } from '@/stores/localeStore';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
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
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Tooltip from '@mui/material/Tooltip';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import SearchIcon from '@mui/icons-material/Search';
import DnsIcon from '@mui/icons-material/Dns';

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { getEnvironmentsApi, createEnvironmentApi, updateEnvironmentApi, deleteEnvironmentApi } from '@/api/environment';
import { formatDateTime } from '@/utils/formatters';
import type { Environment } from '@/types';

export default function Environments() {
  const { t } = useLocaleStore();
  const currentUser = useAuthStore((s) => s.user);
  const userCanWrite = currentUser?.role && currentUser.role !== 'tenant_viewer';
  const [data, setData] = useState<Environment[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ name: '', description: '', isActive: true });
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; onConfirm: () => void } | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params: Record<string, unknown> = { page: page + 1, pageSize: rowsPerPage };
      if (keyword) params.keyword = keyword;
      if (statusFilter !== '') params.is_active = statusFilter === 'true';
      const res = await getEnvironmentsApi(params);
      setData(res.data?.items || []);
      setTotal(res.data?.total || 0);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, [page, rowsPerPage, keyword, statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSearch = () => { setPage(0); fetchData(); };

  const openCreate = () => {
    setEditingId(null);
    setForm({ name: '', description: '', isActive: true });
    setFormError('');
    setDialogOpen(true);
  };

  const openEdit = (row: Environment) => {
    setEditingId(row.id);
    setForm({ name: row.name, description: row.description || '', isActive: row.isActive ?? true });
    setFormError('');
    setDialogOpen(true);
  };

  const handleSave = async () => {
    if (!form.name.trim()) { setFormError(t.common.nameRequired); return; }
    setSaving(true);
    setFormError('');
    try {
      if (editingId) {
        await updateEnvironmentApi(editingId, { name: form.name, description: form.description, isActive: form.isActive });
      } else {
        await createEnvironmentApi({ name: form.name, description: form.description });
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
      message: t.environments.deleteConfirm,
      onConfirm: async () => {
        try {
          await deleteEnvironmentApi(id);
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
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.environments.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} disabled={!userCanWrite}>{t.environments.create}</Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
        <TextField
          size="small"
          placeholder={t.environments.searchPlaceholder}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          sx={{ minWidth: 240 }}
        />
        <FormControl size="small" sx={{ minWidth: 140 }}>
          <InputLabel>{t.environments.statusFilter}</InputLabel>
          <Select value={statusFilter} label={t.environments.statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
            <MenuItem value="">{t.common.all}</MenuItem>
            <MenuItem value="true">{t.common.active}</MenuItem>
            <MenuItem value="false">{t.common.inactive}</MenuItem>
          </Select>
        </FormControl>
        <Button variant="outlined" startIcon={<SearchIcon />} onClick={handleSearch}>{t.common.search}</Button>
      </Paper>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.common.name}</TableCell>
              <TableCell>{t.environments.sdkKey}</TableCell>
              <TableCell>{t.common.status}</TableCell>
              <TableCell>{t.common.createdAt}</TableCell>
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
                            ? 'linear-gradient(135deg, rgba(38,166,154,0.2) 0%, rgba(92,107,192,0.2) 100%)'
                            : 'linear-gradient(135deg, #e0f2f1 0%, #e8eaf6 100%)',
                        mb: 2,
                      }}
                    >
                      <DnsIcon sx={{ fontSize: 40, color: 'primary.main' }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 0.5 }}>
                      {t.environments.emptyTitle}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5, maxWidth: 360, mx: 'auto' }}>
                      {t.environments.emptyDesc}
                    </Typography>
                    {userCanWrite && (
                      <Button
                        variant="contained"
                        onClick={openCreate}
                        sx={{ fontWeight: 700, letterSpacing: '0.5px', px: 3 }}
                      >
                        {t.environments.emptyAction}
                      </Button>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id}>
                <TableCell sx={{ fontWeight: 600 }}>{row.name}</TableCell>
                <TableCell sx={{ fontFamily: 'monospace', fontSize: 12 }}>{row.sdkKey || '-'}</TableCell>
                <TableCell>
                  <Chip
                    label={row.isActive ? t.common.active : t.common.inactive}
                    color={row.isActive ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>{formatDateTime(row.createdAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.edit}><span><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)} disabled={!userCanWrite}><EditIcon /></IconButton></span></Tooltip>
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
          rowsPerPageOptions={[5, 10, 20]}
        />
      </TableContainer>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingId ? t.environments.editEnv : t.environments.createEnv}</DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField
            autoFocus
            fullWidth
            label={t.environments.fieldName}
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            sx={{ mt: 2, mb: 2 }}
          />
          <TextField
            fullWidth
            label={t.environments.fieldDescription}
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            multiline
            rows={3}
            sx={{ mb: 2 }}
          />
          <FormControlLabel
            control={<Switch checked={form.isActive} onChange={(e) => setForm({ ...form, isActive: e.target.checked })} />}
            label={t.environments.fieldActive}
          />
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
