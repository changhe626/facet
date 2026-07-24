import { useState, useEffect, useCallback } from 'react';
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
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import LockResetIcon from '@mui/icons-material/LockReset';
import SearchIcon from '@mui/icons-material/Search';

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { getTenantsApi, createTenantApi, updateTenantApi, resetTenantPasswordApi } from '@/api/tenants';
import { formatDateTime } from '@/utils/formatters';
import type { Tenant } from '@/types';

const levelOptions = ['free', 'pro', 'enterprise'];

interface TenantForm {
  name: string;
  level_id: string;
  description: string;
  website: string;
  status: string;
  admin_email: string;
  admin_password: string;
  admin_role: string;
  new_password: string;
}

const defaultForm: TenantForm = {
  name: '', level_id: 'free', description: '', website: '', status: 'active',
  admin_email: '', admin_password: '', admin_role: 'tenant_super_admin', new_password: '',
};

export default function SystemTenants() {
  const { t } = useLocaleStore();
  const [data, setData] = useState<Tenant[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<TenantForm>(defaultForm);
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; onConfirm: () => void } | null>(null);
  const [dialogMode, setDialogMode] = useState<'create' | 'edit' | 'resetPwd'>('create');

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params: Record<string, unknown> = { page: page + 1, pageSize: rowsPerPage };
      if (keyword) params.keyword = keyword;
      const res = await getTenantsApi(params);
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
    setForm(defaultForm);
    setFormError('');
    setDialogMode('create');
    setDialogOpen(true);
  };

  const openEdit = (row: Tenant) => {
    setEditingId(row.id);
    setForm({ ...defaultForm, name: row.name, level_id: row.levelId ?? 'free', status: row.status });
    setFormError('');
    setDialogMode('edit');
    setDialogOpen(true);
  };

  const openResetPwd = (row: Tenant) => {
    setEditingId(row.id);
    setForm({ ...defaultForm, name: row.name });
    setFormError('');
    setDialogMode('resetPwd');
    setDialogOpen(true);
  };

  const handleSave = async () => {
    setSaving(true);
    setFormError('');
    try {
      if (dialogMode === 'resetPwd') {
        await resetTenantPasswordApi(editingId!, { new_password: form.new_password });
      } else if (dialogMode === 'edit') {
        await updateTenantApi(editingId!, { name: form.name, levelId: form.level_id, status: form.status });
      } else {
        const body = {
          tenant_info: {
            name: form.name,
            levelId: form.level_id,
          },
          user_info: {
            role: form.admin_role,
            adminEmail: form.admin_email,
            adminDisplayName: form.admin_email.split('@')[0],
            adminInitialPassword: form.admin_password,
          },
        };
        await createTenantApi(body);
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
      message: t.systemTenants.deleteConfirm,
      onConfirm: async () => {
        try {
          await updateTenantApi(id, { status: 'archived' });
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
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.systemTenants.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>{t.systemTenants.create}</Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
        <TextField
          size="small"
          placeholder={t.systemTenants.searchPlaceholder}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => (e as React.KeyboardEvent).key === 'Enter' && handleSearch()}
          sx={{ minWidth: 240 }}
        />
        <Button variant="outlined" startIcon={<SearchIcon />} onClick={handleSearch}>{t.common.search}</Button>
      </Paper>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.common.name}</TableCell>
              <TableCell>{t.systemTenants.level}</TableCell>
              <TableCell>{t.common.status}</TableCell>
              <TableCell>{t.common.createdAt}</TableCell>
              <TableCell align="right">{t.common.actions}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={5} align="center"><CircularProgress /></TableCell></TableRow>
            ) : data.length === 0 ? (
              <TableRow><TableCell colSpan={5} align="center">{t.systemTenants.emptyText}</TableCell></TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id}>
                <TableCell sx={{ fontWeight: 600 }}>{row.name}</TableCell>
                <TableCell>{row.levelName || row.levelId}</TableCell>
                <TableCell>
                  <Chip label={row.status === 'active' ? t.systemTenants.statusActive : t.systemTenants.statusSuspended} color={row.status === 'active' ? 'success' : 'default'} size="small" />
                </TableCell>
                <TableCell>{formatDateTime(row.createdAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.edit}><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)}><EditIcon /></IconButton></Tooltip>
                  <Tooltip title={t.systemTenants.resetPassword}><IconButton aria-label={t.systemTenants.resetPassword} size="small" onClick={() => openResetPwd(row)}><LockResetIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.delete}><IconButton aria-label={t.common.delete} size="small" color="error" onClick={() => handleDelete(row.id)}><DeleteIcon /></IconButton></Tooltip>
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

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {dialogMode === 'create' ? t.systemTenants.createTenant : dialogMode === 'resetPwd' ? t.systemTenants.resetPassword : t.systemTenants.editTenant}
        </DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField autoFocus fullWidth label={t.common.name} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} sx={{ mt: 2, mb: 2 }} disabled={dialogMode !== 'create'} />
          {dialogMode === 'create' && (
            <>
              <TextField fullWidth label={t.systemTenants.fieldAdminEmail} value={form.admin_email} onChange={(e) => setForm({ ...form, admin_email: e.target.value })} sx={{ mb: 2 }} />
              <TextField fullWidth label={t.systemTenants.fieldAdminPassword} type="password" value={form.admin_password} onChange={(e) => setForm({ ...form, admin_password: e.target.value })} sx={{ mb: 2 }} />
              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>{t.systemTenants.fieldAdminRole}</InputLabel>
                <Select value={form.admin_role} label={t.systemTenants.fieldAdminRole} onChange={(e) => setForm({ ...form, admin_role: e.target.value })}>
                  <MenuItem value="tenant_super_admin">{t.systemTenants.roleAdmin}</MenuItem>
                  <MenuItem value="tenant_admin">{t.systemTenants.roleReadWrite}</MenuItem>
                  <MenuItem value="tenant_viewer">{t.systemTenants.roleReadOnly}</MenuItem>
                </Select>
              </FormControl>
            </>
          )}
          {dialogMode === 'resetPwd' && (
            <TextField fullWidth label={t.systemTenants.fieldNewPassword} type="password" value={form.new_password} onChange={(e) => setForm({ ...form, new_password: e.target.value })} sx={{ mt: 2 }} />
          )}
          {dialogMode === 'edit' && (
            <>
              <FormControl fullWidth sx={{ mb: 2 }}>
                <InputLabel>{t.systemTenants.fieldLevel}</InputLabel>
                <Select value={form.level_id} label={t.systemTenants.fieldLevel} onChange={(e) => setForm({ ...form, level_id: e.target.value })}>
                  {levelOptions.map((l) => <MenuItem key={l} value={l}>{l}</MenuItem>)}
                </Select>
              </FormControl>
              <FormControl fullWidth>
                <InputLabel>{t.common.status}</InputLabel>
                <Select value={form.status} label={t.common.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
                  <MenuItem value="active">{t.systemTenants.statusActive}</MenuItem>
                  <MenuItem value="suspended">{t.systemTenants.statusSuspended}</MenuItem>
                </Select>
              </FormControl>
            </>
          )}
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
