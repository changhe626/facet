import { useEffect, useRef, useState, type ReactNode } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { styled } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Badge from '@mui/material/Badge';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Menu from '@mui/material/Menu';
import Divider from '@mui/material/Divider';
import Tooltip from '@mui/material/Tooltip';
import Avatar from '@mui/material/Avatar';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';
import LightModeIcon from '@mui/icons-material/LightMode';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import QuestionAnswerOutlined from '@mui/icons-material/QuestionAnswerOutlined';
import NotificationsOutlined from '@mui/icons-material/NotificationsOutlined';
import AccountCircleOutlined from '@mui/icons-material/AccountCircleOutlined';
import LogoutIcon from '@mui/icons-material/Logout';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';
import SettingsIcon from '@mui/icons-material/Settings';
import GroupIcon from '@mui/icons-material/Group';
import BusinessIcon from '@mui/icons-material/Business';
import CampaignIcon from '@mui/icons-material/Campaign';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import TranslateIcon from '@mui/icons-material/Translate';

import { useAuthStore } from '@/stores/authStore';
import { useThemeStore } from '@/stores/themeStore';
import { useEnvStore } from '@/stores/envStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useLocaleStore } from '@/stores/localeStore';
import { setNavigate } from '@/utils/navigate';
import { getUserDetailsApi } from '@/api/user';
import type { User } from '@/types';

const DRAWER_WIDTH = 240;

const Main = styled('main')(({ theme }) => ({
  flexGrow: 1,
  minHeight: '100vh',
  backgroundColor: theme.palette.mode === 'dark'
    ? theme.palette.background.default
    : '#f4f6fb',
}));

const StyledAppBar = styled(AppBar)(({ theme }) => ({
  zIndex: theme.zIndex.drawer + 1,
}));

interface NavItem {
  path: string;
  labelKey: string;
  icon: ReactNode;
  adminOnly: boolean;
}

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { mode, toggleTheme } = useThemeStore();
  const { environments, currentEnvId, setCurrentEnv, fetchEnvironments } = useEnvStore();
  const { unreadCount, fetchUnreadCount } = useNotificationStore();
  const { t, locale, setLocale } = useLocaleStore();

  useEffect(() => { setNavigate(navigate); }, [navigate]);

  const [profileOpen, setProfileOpen] = useState(false);
  const [userDetails, setUserDetails] = useState<User | null>(null);
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
  const intervalRef = useRef<number | null>(null);

  const isAdmin = user?.role === 'system_admin';

  const navItems = [
    { path: '/overview', labelKey: 'overview', icon: <TrendingUpIcon />, adminOnly: false },
    { path: '/feature-flags', labelKey: 'featureFlags', icon: <ToggleOnIcon />, adminOnly: false },
    { path: '/segments', labelKey: 'segments', icon: <FilterAltIcon />, adminOnly: false },
    { path: '/environments', labelKey: 'environments', icon: <SettingsIcon />, adminOnly: false },
    { path: '/users', labelKey: 'users', icon: <GroupIcon />, adminOnly: false },
  ] as NavItem[];

  const adminNavItems = [
    { path: '/system/tenants', labelKey: 'tenants', icon: <BusinessIcon />, adminOnly: true },
    { path: '/system/notifications', labelKey: 'systemNotifications', icon: <CampaignIcon />, adminOnly: true },
  ] as NavItem[];

  useEffect(() => {
    fetchEnvironments();
    fetchUnreadCount();
    intervalRef.current = setInterval(fetchUnreadCount, 60000);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, []);

  const handleProfileOpen = async () => {
    setProfileOpen(true);
    try {
      const res = await getUserDetailsApi(user!.id);
      setUserDetails(res.data);
    } catch {
      setUserDetails(user);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  const isActive = (path: string) => location.pathname === path;

  const toggleLocale = () => {
    setLocale(locale === 'en' ? 'zh' : 'en');
  };

  return (
    <Box sx={{ display: 'flex' }}>
      <StyledAppBar position="fixed" color="inherit" elevation={1}>
        <Toolbar sx={{ gap: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: DRAWER_WIDTH - 16 }}>
            <Box
              component="img"
              src="/logo.svg"
              sx={{ height: 32, width: 32 }}
              alt="Facet"
            />
            <Typography variant="h6" noWrap sx={{ fontWeight: 700 }}>
              Facet
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant="body2" color="text.secondary">
              {t.common.environment}
            </Typography>
            <Select
              size="small"
              value={currentEnvId ?? ''}
              onChange={(e) => setCurrentEnv(e.target.value)}
              sx={{ minWidth: 160 }}
            >
              {environments.map((env) => (
                <MenuItem key={env.id} value={env.id}>
                  {env.name}
                </MenuItem>
              ))}
            </Select>
          </Box>

          <Box sx={{ flexGrow: 1 }} />

          <Tooltip title={t.lang[locale === 'zh' ? 'switchToEn' : 'switchToZh']}>
            <IconButton aria-label={t.lang[locale === 'zh' ? 'switchToEn' : 'switchToZh']} onClick={toggleLocale}>
              <TranslateIcon />
            </IconButton>
          </Tooltip>

          <Tooltip title={mode === 'light' ? t.theme.darkMode : t.theme.lightMode}>
            <IconButton aria-label={mode === 'light' ? t.theme.darkMode : t.theme.lightMode} onClick={toggleTheme}>
              {mode === 'light' ? <DarkModeIcon /> : <LightModeIcon />}
            </IconButton>
          </Tooltip>

          <Tooltip title={t.theme.help}>
            <IconButton aria-label={t.theme.help} onClick={() => navigate('/help')}>
              <QuestionAnswerOutlined />
            </IconButton>
          </Tooltip>

          <Tooltip title={t.theme.notifications}>
            <IconButton aria-label={t.theme.notifications} onClick={() => navigate('/notifications')}>
              <Badge badgeContent={unreadCount} color="error">
                <NotificationsOutlined />
              </Badge>
            </IconButton>
          </Tooltip>

          <IconButton aria-label={t.nav.profile} onClick={(e) => setAnchorEl(e.currentTarget)}>
            <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main', fontSize: 14 }}>
              {user?.displayName?.charAt(0) || 'U'}
            </Avatar>
          </IconButton>
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={() => setAnchorEl(null)}
          >
            <MenuItem onClick={() => { setAnchorEl(null); handleProfileOpen(); }}>
              <AccountCircleOutlined sx={{ mr: 1 }} /> {t.nav.profile}
            </MenuItem>
            <MenuItem onClick={() => { setAnchorEl(null); handleLogout(); }}>
              <LogoutIcon sx={{ mr: 1 }} /> {t.nav.logout}
            </MenuItem>
          </Menu>
        </Toolbar>
      </StyledAppBar>

      <Drawer
        variant="permanent"
        sx={{
          width: DRAWER_WIDTH,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: DRAWER_WIDTH,
            boxSizing: 'border-box',
            borderRight: '1px solid',
            borderColor: 'divider',
            bgcolor: (t) => t.palette.mode === 'dark' ? '#141425' : '#ffffff',
          },
        }}
      >
        <Toolbar />
        <List sx={{ px: 1.5, pt: 1 }}>
          {navItems.map((item) => (
            <ListItem key={item.path} disablePadding sx={{ mb: 0.5 }}>
              <ListItemButton
                selected={isActive(item.path)}
                onClick={() => navigate(item.path)}
                sx={{
                  borderRadius: 2,
                  '&.Mui-selected': {
                    bgcolor: (t) => t.palette.mode === 'dark'
                      ? 'rgba(92,107,192,0.2)'
                      : 'rgba(92,107,192,0.1)',
                    '&::before': {
                      content: '""',
                      position: 'absolute',
                      left: 0,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      width: 3,
                      height: 20,
                      borderRadius: 2,
                      bgcolor: 'primary.main',
                    },
                  },
                  '&.Mui-selected:hover': {
                    bgcolor: (t) => t.palette.mode === 'dark'
                      ? 'rgba(92,107,192,0.3)'
                      : 'rgba(92,107,192,0.15)',
                  },
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 40,
                    color: isActive(item.path) ? 'primary.main' : undefined,
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                  primary={t.nav[item.labelKey as keyof typeof t.nav]}
                  slotProps={{
                    primary: {
                      fontWeight: isActive(item.path) ? 700 : 500,
                      fontSize: 14,
                    },
                  }}
                />
              </ListItemButton>
            </ListItem>
          ))}
        </List>

        {isAdmin && (
          <>
            <Divider sx={{ mx: 2, my: 1 }} />
            <Typography variant="caption" sx={{ px: 3, py: 1, color: 'text.secondary', fontWeight: 600 }}>
              {t.nav.systemAdmin}
            </Typography>
            <List sx={{ px: 1.5 }}>
              {adminNavItems.map((item) => (
                <ListItem key={item.path} disablePadding sx={{ mb: 0.5 }}>
                  <ListItemButton
                    selected={isActive(item.path)}
                    onClick={() => navigate(item.path)}
                    sx={{
                      borderRadius: 2,
                      '&.Mui-selected': {
                        bgcolor: (t) => t.palette.mode === 'dark'
                          ? 'rgba(92,107,192,0.2)'
                          : 'rgba(92,107,192,0.1)',
                        '&::before': {
                          content: '""',
                          position: 'absolute',
                          left: 0,
                          top: '50%',
                          transform: 'translateY(-50%)',
                          width: 3,
                          height: 20,
                          borderRadius: 2,
                          bgcolor: 'primary.main',
                        },
                      },
                    }}
                  >
                    <ListItemIcon
                      sx={{
                        minWidth: 40,
                        color: isActive(item.path) ? 'primary.main' : undefined,
                      }}
                    >
                      {item.icon}
                    </ListItemIcon>
                    <ListItemText primary={t.nav[item.labelKey as keyof typeof t.nav]} />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </>
        )}
      </Drawer>

      <Main>
        <Toolbar />
        <Box sx={{ p: 3 }}>
          <Outlet context={{ currentEnvId }} />
        </Box>
      </Main>

      <Dialog open={profileOpen} onClose={() => setProfileOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t.profile.title}</DialogTitle>
        <DialogContent>
          {(userDetails || user) && (
            <Table size="small">
              <TableBody>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600, width: 140 }}>{t.profile.userId}</TableCell>
                  <TableCell>{(userDetails || user).id}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{t.profile.displayName}</TableCell>
                  <TableCell>{(userDetails || user).displayName}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{t.profile.email}</TableCell>
                  <TableCell>{(userDetails || user).email}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{t.profile.role}</TableCell>
                  <TableCell>{(userDetails || user).role}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{t.profile.createdAt}</TableCell>
                  <TableCell>{(userDetails || user).createdAt || '-'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 600 }}>{t.profile.lastLogin}</TableCell>
                  <TableCell>{(userDetails || user).lastLoginAt || '-'}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          )}
        </DialogContent>
      </Dialog>
    </Box>
  );
}
