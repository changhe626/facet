import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import { useNavigate } from 'react-router-dom';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { useLocaleStore } from '@/stores/localeStore';

export default function NotFound() {
  const navigate = useNavigate();
  const { t } = useLocaleStore();
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', textAlign: 'center', px: 2 }}>
      <Box
        sx={{
          width: 100, height: 100, borderRadius: '50%',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'linear-gradient(135deg, #e8eaf6 0%, #fce4ec 100%)',
          mb: 3,
        }}
      >
        <ErrorOutlineIcon sx={{ fontSize: 48, color: 'text.secondary' }} />
      </Box>
      <Typography variant="h4" sx={{ fontWeight: 800, mb: 1 }}>{t.notFound.title}</Typography>
      <Typography variant="h6" color="text.secondary" sx={{ mb: 0.5 }}>{t.notFound.subtitle}</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 3, maxWidth: 400 }}>
        {t.notFound.desc}
      </Typography>
      <Button variant="contained" onClick={() => navigate('/overview', { replace: true })}>
        {t.notFound.back}
      </Button>
    </Box>
  );
}
