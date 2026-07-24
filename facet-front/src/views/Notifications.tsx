import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';

import { getErrorMessage } from '@/utils/errors';
import { useLocaleStore } from '@/stores/localeStore';
import { getNotificationsApi, markNotificationsAsReadApi } from '@/api/notifications';
import { formatDateTime } from '@/utils/formatters';
import type { Notification } from '@/types';

const levelColors: Record<string, 'default' | 'info' | 'warning' | 'error'> = {
  normal: 'info',
  important: 'warning',
  urgent: 'error',
};

export default function Notifications() {
  const { t } = useLocaleStore();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);
  const [marking, setMarking] = useState(false);
  const [error, setError] = useState('');

  const fetchNotifications = async () => {
    setLoading(true);
    try {
      const res = await getNotificationsApi();
      setNotifications(res.data || []);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchNotifications(); }, []);

  const handleMarkAsRead = async (ids: string[]) => {
    setMarking(true);
    try {
      await markNotificationsAsReadApi(ids);
      await fetchNotifications();
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setMarking(false);
    }
  };

  const unreadIds = notifications.filter((n) => !n.isRead).map((n) => n.id);

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>{t.notifications.title}</Typography>
        {unreadIds.length > 0 && (
          <Button variant="outlined" size="small" onClick={() => handleMarkAsRead(unreadIds)} disabled={marking}>
            {marking ? <CircularProgress size={16} /> : `${t.notifications.markAllAsRead} (${unreadIds.length})`}
          </Button>
        )}
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}><CircularProgress /></Box>
      ) : notifications.length === 0 ? (
        <Typography color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>{t.notifications.noNotifications}</Typography>
      ) : (
        <List>
          {notifications.map((n, i) => (
            <Box key={n.id}>
              {i > 0 && <Divider component="li" />}
              <ListItem
                alignItems="flex-start"
                sx={{ bgcolor: n.isRead ? 'transparent' : 'action.hover', borderRadius: 1, mb: 0.5 }}
                secondaryAction={
                  !n.isRead && (
                    <Button size="small" onClick={() => handleMarkAsRead([n.id])}>
                      {t.notifications.markRead}
                    </Button>
                  )
                }
              >
                <ListItemText
                  primary={
                    <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 0.5 }}>
                      <Chip label={n.level} size="small" color={levelColors[n.level] || 'default'} />
                      <Typography variant="subtitle2" sx={{ fontWeight: n.isRead ? 400 : 600 }}>
                        {n.title}
                      </Typography>
                    </Box>
                  }
                  secondary={
                    <>
                      <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5 }}>
                        {n.content}
                      </Typography>
                      <Typography variant="caption" color="text.disabled">
                        {formatDateTime(n.createdAt)}
                      </Typography>
                    </>
                  }
                />
              </ListItem>
            </Box>
          ))}
        </List>
      )}
    </Box>
  );
}
