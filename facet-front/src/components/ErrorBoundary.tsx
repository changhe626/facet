import { Component, type ReactNode, type ErrorInfo } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

function DefaultFallback() {
  return (
    <Box sx={{ p: 4, maxWidth: 600, mx: 'auto', mt: 4 }}>
      <Alert severity="error" sx={{ mb: 2 }}>
        <Typography variant="h6" sx={{ mb: 1 }}>Something went wrong</Typography>
        <Typography variant="body2" sx={{ mb: 2, fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-word' }}>
          An unexpected error occurred. Please try reloading the page.
        </Typography>
        <Button variant="outlined" size="small" onClick={() => window.location.reload()}>
          Reload Page
        </Button>
      </Alert>
    </Box>
  );
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  static ErrorFallback = DefaultFallback;
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    if (import.meta.env.DEV) {
      console.error('ErrorBoundary caught:', error, info);
    }
  }

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <Box sx={{ p: 4, maxWidth: 600, mx: 'auto', mt: 4 }}>
          <Alert severity="error" sx={{ mb: 2 }}>
            <Typography variant="h6" sx={{ mb: 1 }}>Something went wrong</Typography>
            <Typography variant="body2" sx={{ mb: 2, fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-word' }}>
              {this.state.error?.message}
            </Typography>
            <Button variant="outlined" size="small" onClick={() => window.location.reload()}>
              Reload Page
            </Button>
          </Alert>
        </Box>
      );
    }

    return this.props.children;
  }
}