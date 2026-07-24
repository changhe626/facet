import { useState, useEffect, useCallback } from 'react';
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

import { getErrorMessage } from '@/utils/errors';
import ConfirmDialog from '@/components/ConfirmDialog';
import { getSystemNotificationsApi, createSystemNotificationApi, getSystemNotificationDetailsApi, updateSystemNotificationApi, deleteSystemNotificationApi } from '@/api/systemNotifications';
import { formatDateTime } from '@/utils/formatters';
import type { SystemNotification } from '@/types';

const levels = ['normal', 'important', 'urgent'];
const types = ['announcement', 'product_update', 'security_alert'];

export default function SystemNotifications() {
  const { t } = useLocaleStore();
  const [data, setData] = useState<SystemNotification[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState({ title: '', content: '', level: 'normal', type: 'announcement' });
  const [formError, setFormError] = useState('');
  const [saving, setSaving] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailData, setDetailData] = useState<SystemNotification | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{ message: string; onConfirm: () => void } | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await getSystemNotificationsApi({ page: page + 1, pageSize: rowsPerPage });
      setData(res.data?.items || []);
      setTotal(res.data?.total || 0);
    } catch (e) { setError(getErrorMessage(e)); }
    finally { setLoading(false); }
  }, [page, rowsPerPage]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openCreate = () => {
    setEditingId(null);
    setForm({ title: '', content: '', level: 'normal', type: 'announcement' });
    setFormError('');
    setDialogOpen(true);
  };

  const openEdit = async (row: SystemNotification) => {
    setEditingId(row.id);
    try {
      const res = await getSystemNotificationDetailsApi(row.id);
      const n = res.data;
      setForm({ title: n.title || '', content: n.content || '', level: n.level || 'normal', type: n.type || 'announcement' });
    } catch (e) { setError(getErrorMessage(e)); return; }
    setFormError('');
    setDialogOpen(true);
  };

  const handleSave = async () => {
    if (!form.title.trim() || !form.content.trim()) { setFormError(t.systemNotifications.errors.titleAndContentRequired); return; }
    setSaving(true);
    setFormError('');
    try {
      if (editingId) {
        await updateSystemNotificationApi(editingId, { title: form.title, content: form.content, level: form.level, type: form.type });
      } else {
        await createSystemNotificationApi({ title: form.title, content: form.content, level: form.level, type: form.type });
      }
      setDialogOpen(false);
      fetchData();
    } catch (e) { setFormError(getErrorMessage(e)); }
    finally { setSaving(false); }
  };

  const handleDelete = (id: string) => {
    setConfirmDialog({
      message: t.systemNotifications.deleteConfirm,
      onConfirm: async () => {
        try { await deleteSystemNotificationApi(id); fetchData(); }
        catch (e) { setError(getErrorMessage(e)); }
      },
    });
  };

  const handleViewDetails = async (row: SystemNotification) => {
    try {
      const res = await getSystemNotificationDetailsApi(row.id);
      setDetailData(res.data);
      setDetailOpen(true);
    } catch (e) { setError(getErrorMessage(e)); }
  };

  const levelColor = (l: string): 'error' | 'warning' | 'info' | 'default' => l === 'urgent' ? 'error' : l === 'important' ? 'warning' : 'info';

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.systemNotifications.title}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>{t.systemNotifications.create}</Button>
      </Box>
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t.systemNotifications.fieldTitle}</TableCell>
              <TableCell>{t.systemNotifications.fieldLevel}</TableCell>
              <TableCell>{t.systemNotifications.fieldType}</TableCell>
              <TableCell>{t.common.createdAt}</TableCell>
              <TableCell align="right">{t.common.actions}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow><TableCell colSpan={5} align="center"><CircularProgress /></TableCell></TableRow>
            ) : data.length === 0 ? (
              <TableRow><TableCell colSpan={5} align="center">{t.systemNotifications.emptyText}</TableCell></TableRow>
            ) : data.map((row) => (
              <TableRow key={row.id}>
                <TableCell sx={{ fontWeight: 600, maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{row.title}</TableCell>
                <TableCell><Chip label={row.level} color={levelColor(row.level)} size="small" /></TableCell>
                <TableCell><Chip label={row.type} size="small" variant="outlined" /></TableCell>
                <TableCell>{formatDateTime(row.createdAt)}</TableCell>
                <TableCell align="right">
                  <Tooltip title={t.common.viewDetails}><IconButton aria-label={t.common.viewDetails} size="small" onClick={() => handleViewDetails(row)}><VisibilityIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.edit}><IconButton aria-label={t.common.edit} size="small" onClick={() => openEdit(row)}><EditIcon /></IconButton></Tooltip>
                  <Tooltip title={t.common.delete}><IconButton aria-label={t.common.delete} size="small" color="error" onClick={() => handleDelete(row.id)}><DeleteIcon /></IconButton></Tooltip>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination component="div" count={total} page={page} onPageChange={(_, p) => setPage(p)} rowsPerPage={rowsPerPage} onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value, 10)); setPage(0); }} rowsPerPageOptions={[10, 20, 50]} />
      </TableContainer>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingId ? t.systemNotifications.editNotification : t.systemNotifications.createNotification}</DialogTitle>
        <DialogContent>
          {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
          <TextField autoFocus fullWidth label={t.systemNotifications.fieldTitle} value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} sx={{ mt: 2, mb: 2 }} />
          <TextField fullWidth label={t.systemNotifications.fieldContent} value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })} multiline rows={4} sx={{ mb: 2 }} />
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>{t.systemNotifications.fieldLevel}</InputLabel>
            <Select value={form.level} label={t.systemNotifications.fieldLevel} onChange={(e) => setForm({ ...form, level: e.target.value })}>
              {levels.map((l) => <MenuItem key={l} value={l}>{l}</MenuItem>)}
            </Select>
          </FormControl>
          <FormControl fullWidth>
            <InputLabel>{t.systemNotifications.fieldType}</InputLabel>
            <Select value={form.type} label={t.systemNotifications.fieldType} onChange={(e) => setForm({ ...form, type: e.target.value })}>
              {types.map((t) => <MenuItem key={t} value={t}>{t}</MenuItem>)}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>{t.common.cancel}</Button>
          <Button variant="contained" onClick={handleSave} disabled={saving}>{saving ? <CircularProgress size={20} /> : t.common.save}</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t.systemNotifications.notificationDetails}</DialogTitle>
        <DialogContent>
          {detailData && (
            <Table size="small">
              <TableBody>
                {([
                  [t.systemNotifications.fieldTitle, detailData.title],
                  [t.systemNotifications.fieldContent, detailData.content],
                  [t.systemNotifications.fieldLevel, detailData.level],
                  [t.systemNotifications.fieldType, detailData.type],
                  [t.common.createdAt, formatDateTime(detailData.createdAt)],
                ] as [string, string][]).map(([k, v]) => (
                  <TableRow key={k}>
                    <TableCell sx={{ fontWeight: 600, width: 120 }}>{k}</TableCell>
                    <TableCell sx={{ whiteSpace: 'pre-wrap' }}>{v || '-'}</TableCell>
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
