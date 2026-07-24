import { useMemo } from 'react';
import { RouterProvider } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material';
import { router } from '@/router';
import { useThemeStore } from '@/stores/themeStore';
import ErrorBoundary from '@/components/ErrorBoundary';

export default function App() {
  const mode = useThemeStore((s) => s.mode);

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode,
          primary: { main: '#5c6bc0', light: '#7986cb', dark: '#3949ab' },
          secondary: { main: '#26a69a', light: '#4db6ac', dark: '#00897b' },
          ...(mode === 'dark'
            ? {
                background: { default: '#0f0f1a', paper: '#1a1a2e' },
                text: { primary: '#e0e0e0', secondary: '#a0a0b8' },
                divider: 'rgba(255,255,255,0.08)',
              }
            : {
                background: { default: '#f4f6fb', paper: '#ffffff' },
                text: { primary: '#1a1a2e', secondary: '#5a5a72' },
                divider: 'rgba(0,0,0,0.08)',
              }),
        },
        shape: { borderRadius: 10 },
        typography: {
          fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
          button: { textTransform: 'none', fontWeight: 600 },
        },
        components: {
          MuiCard: {
            styleOverrides: {
              root: {
                borderRadius: 12,
                boxShadow: mode === 'dark'
                  ? '0 2px 12px rgba(0,0,0,0.3)'
                  : '0 2px 12px rgba(0,0,0,0.06)',
              },
            },
          },
          MuiPaper: {
            styleOverrides: {
              root: {
                borderRadius: 12,
              },
            },
          },
          MuiButton: {
            styleOverrides: {
              root: {
                borderRadius: 8,
                padding: '8px 20px',
              },
              contained: {
                boxShadow: 'none',
                '&:hover': {
                  boxShadow: '0 4px 12px rgba(92,107,192,0.3)',
                },
              },
            },
          },
          MuiTableCell: {
            styleOverrides: {
              head: {
                fontWeight: 600,
                color: mode === 'dark' ? '#a0a0b8' : '#5a5a72',
              },
            },
          },
          MuiChip: {
            styleOverrides: {
              root: { borderRadius: 6 },
            },
          },
          MuiDrawer: {
            styleOverrides: {
              paper: {
                borderRight: mode === 'dark'
                  ? '1px solid rgba(255,255,255,0.08)'
                  : '1px solid rgba(0,0,0,0.08)',
              },
            },
          },
          MuiAppBar: {
            styleOverrides: {
              root: {
                backdropFilter: 'blur(8px)',
                backgroundColor: mode === 'dark'
                  ? 'rgba(26,26,46,0.9)'
                  : 'rgba(255,255,255,0.9)',
              },
            },
          },
          MuiTable: {
            styleOverrides: {
              root: {
                '& .MuiTableRow-root:not(:last-child) td': {
                  borderBottom: mode === 'dark'
                    ? '1px solid rgba(255,255,255,0.06)'
                    : '1px solid rgba(0,0,0,0.06)',
                },
              },
            },
          },
        },
      }),
    [mode]
  );

  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <RouterProvider router={router} />
      </ThemeProvider>
    </ErrorBoundary>
  );
}
