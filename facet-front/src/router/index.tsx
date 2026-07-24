import { lazy, Suspense, type ReactNode } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '@/layouts/MainLayout';
import { useAuthStore } from '@/stores/authStore';
import CircularProgress from '@mui/material/CircularProgress';
import Box from '@mui/material/Box';

const Login = lazy(() => import('@/views/Login'));
const Overview = lazy(() => import('@/views/Overview'));
const FeatureFlags = lazy(() => import('@/views/FeatureFlags'));
const Environments = lazy(() => import('@/views/Environments'));
const Users = lazy(() => import('@/views/Users'));
const SystemTenants = lazy(() => import('@/views/SystemTenants'));
const Notifications = lazy(() => import('@/views/Notifications'));
const HelpCenter = lazy(() => import('@/views/HelpCenter'));
const SystemNotifications = lazy(() => import('@/views/SystemNotifications'));
const Segments = lazy(() => import('@/views/Segments'));
const NotFound = lazy(() => import('@/views/NotFound'));

const LazyLoad = ({ children }: { children: ReactNode }) => (
  <Suspense fallback={<Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}><CircularProgress /></Box>}>
    {children}
  </Suspense>
);

const ProtectedRoute = ({ children }: { children: ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return children;
};

const PublicRoute = ({ children }: { children: ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (token) return <Navigate to="/overview" replace />;
  return children;
};

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <PublicRoute><LazyLoad><Login /></LazyLoad></PublicRoute>,
  },
  {
    path: '/',
    element: <ProtectedRoute><MainLayout /></ProtectedRoute>,
    children: [
      { index: true, element: <Navigate to="/overview" replace /> },
      { path: 'overview', element: <LazyLoad><Overview /></LazyLoad> },
      { path: 'feature-flags', element: <LazyLoad><FeatureFlags /></LazyLoad> },
      { path: 'environments', element: <LazyLoad><Environments /></LazyLoad> },
      { path: 'users', element: <LazyLoad><Users /></LazyLoad> },
      { path: 'system/tenants', element: <LazyLoad><SystemTenants /></LazyLoad> },
      { path: 'segments', element: <LazyLoad><Segments /></LazyLoad> },
      { path: 'notifications', element: <LazyLoad><Notifications /></LazyLoad> },
      { path: 'help', element: <LazyLoad><HelpCenter /></LazyLoad> },
      { path: 'system/notifications', element: <LazyLoad><SystemNotifications /></LazyLoad> },
    ],
  },
  {
    path: '*',
    element: <LazyLoad><NotFound /></LazyLoad>,
  },
]);
