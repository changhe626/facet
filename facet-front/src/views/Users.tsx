import { useState, useEffect, useCallback } from 'react';
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
import IconButton from '@mui/material/IconButton';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Tooltip from '@mui/material/Tooltip';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import VisibilityIcon from '@mui/icons-material/Visibility';
import LockResetIcon from '@mui/icons-material/LockReset';
import SearchIcon from '@mui/icons-material/Search';
import PeopleIcon from '@mui/icons-material/People';

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { getUsersApi, createUserApi, updateUserApi, deleteUserApi, resetUserPasswordApi } from '@/api/user';
import { formatRole, formatDateTime } from '@/utils/formatters';
import { useAuthStore } from '@/stores/authStore';
import { useLocaleStore } from '@/stores/localeStore';
import type { User } from '@/types';

const roles = [
  { value: 'tenant_admin', labelKey: 'roleReadWrite' },
  { value: 'tenant_viewer', labelKey: 'roleReadOnly' },
  { value: 'tenant_super_admin', labelKey: 'roleAdmin' },
];

export default function Users() {
  const { t } = useLocaleStore();
  const currentUser = useAuthStore((s) => s.user);
  const userCanWrite = currentUser?.role && currentUser.role !== 'tenant_viewer';
  const canResetPwd = currentUser?.role === 'tenant_super_admin' || currentUser?.role === 'system_admin';

  const [data, setData] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ displayName: '', email: '', role: 'tenant_viewer', initialPassword: '' });
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailData, setDetailData] = useState<User | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; confirmLabel?: string; onConfirm: () => void } | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params: Record<string, unknown> = { page: page + 1, pageSize: rowsPerPage };
      if (keyword) params.keyword = keyword;
      const res = await getUsersApi(params);
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
    setForm({ displayName: '', email: '', role: 'tenant_viewer', initialPassword: '' });
    setFormError('');
    setDialogOpen(true);
  };

  const openEdit = (row: User) => {
    setEditingId(row.id);
    setForm({ displayName: row.displayName || '', email: row.email || '', role: row.role, initialPassword: '' });
    setFormError('');
    setDialogOpen(true);
  };

  const handleSave = async () => {
    if (!form.displayName.trim() || !form.email.trim()) {
      setFormError(t.users.errors.nameAndEmailRequired);
      return;
    }
    if (!editingId && !form.initialPassword.trim()) {
      setFormError(t.users.errors.passwordRequired);
      return;
    }
    setSaving(true);
    setFormError('');
    try {
      if (editingId) {
        await updateUserApi(editingId, { displayName: form.displayName, email: form.email, role: form.role });
      } else {
        await createUserApi({ displayName: form.displayName, email: form.email, role: form.role, initialPassword: form.initialPassword });
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
      message: t.users.deleteConfirm,
      onConfirm: async () => {
        try {
          await deleteUserApi(id);
          fetchData();
        } catch (e) {
          setError(getErrorMessage(e));
        }
      },
    });
  };

  const handleViewDetails = (row: User) => {
    setDetailData(row);
    setDetailOpen(true);
  };

  const handleResetPwd = (id: string) => {
    setConfirmDialog({
      message: t.users.resetPwdConfirm,
      confirmLabel: t.common.reset,
      onConfirm: async () => {
        try {
          await resetUserPasswordApi(id);
          setError(t.users.pwdResetSuccess);
        } catch (e) {
          setError(getErrorMessage(e));
        }
      },
    });
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.users.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} disabled={!userCanWrite}>{t.users.create}</Button>
      </Box>

      {error && <Alert severity={error.includes('successfully') ? 'success' : 'error'} sx={{ mb: 2 }}>{error}</Alert>}

      <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 2, alignItems: 'center' }}>
        <TextField
          size="small"
          placeholder={t.users.searchPlaceholder}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => (e as React.KeyboardEvent<HTMLDivElement>).key === 'Enter' && handleSearch()}
          sx={{ minWidth: 240 }}
        />
        <Button variant="outlined" startIcon={<SearchIcon />} onClick={handleSearch}>{t.common.search}</Button>
      </Paper>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.users.displayName}</TableCell>
              <TableCell>{t.users.email}</TableCell>
              <TableCell>{t.users.role}</TableCell>
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
                            ? 'linear-gradient(135deg, rgba(250,112,154,0.2) 0%, rgba(254,225,64,0.2) 100%)'
                            : 'linear-gradient(135deg, #fce4ec 0%, #fff9c4 100%)',
                        mb: 2,
                      }}
                    >
                      <PeopleIcon sx={{ fontSize: 40, color: 'secondary.main' }} />
                    </Box>
                    <Typography variant="h6" sx={{ fontWeight: 700, mb: 0.5 }}>
                      {t.users.emptyTitle}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5, maxWidth: 360, mx: 'auto' }}>
                      {t.users.emptyDesc}
                    </Typography>
                    {userCanWrite && (
                      <Button
                        variant="contained"
                        onClick={openCreate}
                        sx={{ fontWeight: 700, letterSpacing: '0.5px', px: 3 }}
                      >
                        {t.users.emptyAction}
                      </Button>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id}>
                <TableCell sx={{ fontWeight: 600 }}>{row.displayName || '-'}</TableCell>
                <TableCell>{row.email}</TableCell>
                <TableCell>{formatRole(row.role)}</TableCell>
                <TableCell>{formatDateTime(row.createdAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.viewDetails}><IconButton aria-label={t.common.viewDetails} size="small" onClick={() => handleViewDetails(row)}><VisibilityIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.edit}><span><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)} disabled={!userCanWrite}><EditIcon /></IconButton></span></Tooltip>
                  <Tooltip title={t.common.delete}><span><IconButton aria-label={t.common.delete} size="small" color="error" onClick={() => handleDelete(row.id)} disabled={!userCanWrite}><DeleteIcon /></IconButton></span></Tooltip>
                  {canResetPwd && userCanWrite && (
                    <Tooltip title={t.common.reset}><IconButton aria-label={t.common.reset} size="small" onClick={() => handleResetPwd(row.id)}><LockResetIcon /></IconButton></Tooltip>
                  )}
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
        <DialogTitle>{editingId ? t.users.editUser : t.users.createUser}</DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField autoFocus fullWidth label={t.users.fieldDisplayName} value={form.displayName} onChange={(e) => setForm({ ...form, displayName: e.target.value })} sx={{ mt: 2, mb: 2 }} />
          <TextField fullWidth label={t.users.fieldEmail} value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} sx={{ mb: 2 }} />
          {!editingId && (
            <TextField fullWidth label={t.users.fieldPassword} type="password" value={form.initialPassword} onChange={(e) => setForm({ ...form, initialPassword: e.target.value })} sx={{ mb: 2 }} />
          )}
          <FormControl fullWidth>
            <InputLabel>{t.users.fieldRole}</InputLabel>
            <Select value={form.role} label={t.users.fieldRole} onChange={(e) => setForm({ ...form, role: e.target.value })}>
              {roles.map((r) => <MenuItem key={r.value} value={r.value}>{t.systemTenants[r.labelKey as keyof typeof t.systemTenants] as string}</MenuItem>)}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>{t.common.cancel}</Button>
          <Button variant="contained" onClick={handleSave} disabled={saving}>
            {saving ? <CircularProgress size={20} /> : t.common.save}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t.users.userDetails}</DialogTitle>
        <DialogContent>
          {detailData && (
            <Table size="small">
              <TableBody>
                {([
                  [t.users.fieldDisplayName, detailData.displayName || '-'],
                  [t.users.fieldEmail, detailData.email],
                  [t.users.fieldRole, formatRole(detailData.role)],
                  [t.common.createdAt, formatDateTime(detailData.createdAt)],
                  [t.users.createdBy, detailData.createdBy || '-'],
                  [t.common.updatedAt, formatDateTime(detailData.updatedAt)],
                  [t.users.updatedBy, detailData.updatedBy || '-'],
                ] as [string, string][]).map(([k, v]) => (
                  <TableRow key={k}>
                    <TableCell sx={{ fontWeight: 600, width: 140 }}>{k}</TableCell>
                    <TableCell>{v}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>{t.common.close}</Button>
        </DialogActions>
      </Dialog>

      <ConfirmDialog
        open={confirmDialog !== null}
        message={confirmDialog?.message ?? ''}
        confirmLabel={confirmDialog?.confirmLabel}
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
