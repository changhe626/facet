import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import EmailOutlined from '@mui/icons-material/EmailOutlined';
import LockOutlined from '@mui/icons-material/LockOutlined';
import ShieldOutlined from '@mui/icons-material/ShieldOutlined';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';

import { getErrorMessage } from '@/utils/errors';
import { useAuthStore } from '@/stores/authStore';
import { useLocaleStore } from '@/stores/localeStore';
import { loginApi, getCaptchaApi } from '@/api/auth';

export default function Login() {
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const { t } = useLocaleStore();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [captchaAnswer, setCaptchaAnswer] = useState('');
  const [captchaId, setCaptchaId] = useState('');
  const [captchaImage, setCaptchaImage] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [loading, setLoading] = useState(false);
  const [captchaLoading, setCaptchaLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchCaptcha = async () => {
    setCaptchaLoading(true);
    try {
      const res = await getCaptchaApi();
      setCaptchaId(res.data.captchaId);
      setCaptchaImage(res.data.base64Image);
    } catch {
    } finally {
      setCaptchaLoading(false);
    }
  };

  useEffect(() => {
    fetchCaptcha();
  }, []);

  const handleLogin = async () => {
    if (!email.trim() || !password || !captchaAnswer) {
      setError(t.login.allFieldsRequired);
      return;
    }
    setLoading(true);
    setError('');
    try {
      const res = await loginApi({ email: email.trim(), password, captchaId, captchaAnswer });
      const { token, user: userInfo } = res.data;
      login(token, userInfo);
      navigate('/overview', { replace: true });
    } catch (e) {
      setError(getErrorMessage(e) || t.login.loginFailed);
      fetchCaptcha();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
        background: (t) =>
          t.palette.mode === 'dark'
            ? 'linear-gradient(135deg, #0f0f1a 0%, #1a1a2e 50%, #16213e 100%)'
            : 'linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: '-50%',
          right: '-20%',
          width: 600,
          height: 600,
          borderRadius: '50%',
          background: (t) =>
            t.palette.mode === 'dark'
              ? 'radial-gradient(circle, rgba(92,107,192,0.15) 0%, transparent 70%)'
              : 'radial-gradient(circle, rgba(255,255,255,0.15) 0%, transparent 70%)',
          pointerEvents: 'none',
        },
        '&::after': {
          content: '""',
          position: 'absolute',
          bottom: '-30%',
          left: '-10%',
          width: 400,
          height: 400,
          borderRadius: '50%',
          background: (t) =>
            t.palette.mode === 'dark'
              ? 'radial-gradient(circle, rgba(38,166,154,0.1) 0%, transparent 70%)'
              : 'radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%)',
          pointerEvents: 'none',
        },
      }}
    >
      <Card
        sx={{
          width: 420,
          maxWidth: '90vw',
          borderRadius: 4,
          boxShadow: (t) =>
            t.palette.mode === 'dark'
              ? '0 8px 40px rgba(0,0,0,0.4)'
              : '0 8px 40px rgba(102,126,234,0.2)',
          backdropFilter: 'blur(20px)',
          backgroundColor: (t) =>
            t.palette.mode === 'dark'
              ? 'rgba(26,26,46,0.95)'
              : 'rgba(255,255,255,0.95)',
          position: 'relative',
          zIndex: 1,
        }}
      >
        <CardContent sx={{ p: 4 }}>
          <Box sx={{ textAlign: 'center', mb: 3 }}>
            <Box
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: 56,
                height: 56,
                borderRadius: 3,
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                color: '#fff',
                mb: 1.5,
              }}
            >
              <ToggleOnIcon sx={{ fontSize: 32 }} />
            </Box>
            <Typography variant="h5" sx={{ fontWeight: 800, letterSpacing: '-0.5px' }}>
              Facet
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
              {t.app.subtitle}
            </Typography>
          </Box>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
              {error}
            </Alert>
          )}

          <TextField
            fullWidth
            label={t.login.email}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start"><EmailOutlined fontSize="small" /></InputAdornment>
                ),
              },
            }}
            sx={{ mb: 2 }}
          />

          <TextField
            fullWidth
            label={t.login.password}
            type={showPwd ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start"><LockOutlined fontSize="small" /></InputAdornment>
                ),
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton aria-label={showPwd ? t.login.hidePassword : t.login.showPassword} onClick={() => setShowPwd(!showPwd)} edge="end" size="small">
                      {showPwd ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                ),
              },
            }}
            sx={{ mb: 2 }}
          />

          <Box sx={{ display: 'flex', gap: 1, mb: 3, alignItems: 'center' }}>
            <TextField
              fullWidth
              label={t.login.captcha}
              value={captchaAnswer}
              onChange={(e) => setCaptchaAnswer(e.target.value)}
              size="small"
              slotProps={{
                input: {
                  startAdornment: (
                    <InputAdornment position="start"><ShieldOutlined fontSize="small" color="action" /></InputAdornment>
                  ),
                },
              }}
            />
            <Box
              sx={{
                width: 120,
                height: 40,
                borderRadius: 1.5,
                overflow: 'hidden',
                cursor: 'pointer',
                flexShrink: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                border: '1px solid',
                borderColor: 'divider',
                transition: 'border-color 0.2s',
                '&:hover': { borderColor: 'primary.main' },
              }}
              onClick={fetchCaptcha}
            >
              {captchaLoading ? (
                <CircularProgress size={20} />
              ) : (
                <Box
                  component="img"
                  src={captchaImage}
                  alt={t.login.captcha}
                  sx={{ width: '100%', height: '100%', objectFit: 'cover' }}
                />
              )}
            </Box>
          </Box>

          <Button
            fullWidth
            variant="contained"
            size="large"
            onClick={handleLogin}
            disabled={loading}
            sx={{
              py: 1.3,
              borderRadius: 2,
              fontSize: 16,
              fontWeight: 700,
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #5c6bc0 0%, #6a3d9a 100%)',
                boxShadow: '0 4px 16px rgba(102,126,234,0.4)',
              },
            }}
          >
            {loading ? <CircularProgress size={22} color="inherit" /> : t.login.signIn}
          </Button>
        </CardContent>
      </Card>
    </Box>
  );
}
