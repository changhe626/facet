import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Skeleton from '@mui/material/Skeleton';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';
import ScienceIcon from '@mui/icons-material/Science';
import GroupIcon from '@mui/icons-material/Group';
import BoltIcon from '@mui/icons-material/Bolt';
import DnsIcon from '@mui/icons-material/Dns';
import FlagIcon from '@mui/icons-material/Flag';
import PeopleIcon from '@mui/icons-material/People';
import { getEnvironmentsApi } from '@/api/environment';
import { getUsersApi } from '@/api/user';
import { useLocaleStore } from '@/stores/localeStore';

export default function Overview() {
  const { t } = useLocaleStore();
  const [envCount, setEnvCount] = useState<number | null>(null);
  const [userCount, setUserCount] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  const featureCards = [
    { title: t.overview.featureFlagsTitle, desc: t.overview.featureFlagsDesc, icon: <ToggleOnIcon sx={{ fontSize: 40 }} />, color: '#5c6bc0' },
    { title: t.overview.abTestingTitle, desc: t.overview.abTestingDesc, icon: <ScienceIcon sx={{ fontSize: 40 }} />, color: '#26a69a' },
    { title: t.overview.multiTenantTitle, desc: t.overview.multiTenantDesc, icon: <GroupIcon sx={{ fontSize: 40 }} />, color: '#ff7043' },
    { title: t.overview.realtimeEvalTitle, desc: t.overview.realtimeEvalDesc, icon: <BoltIcon sx={{ fontSize: 40 }} />, color: '#ab47bc' },
  ];

  const statConfig = [
    { label: t.overview.environments, icon: <DnsIcon />, gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
    { label: t.overview.featureFlags, icon: <FlagIcon />, gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' },
    { label: t.overview.users, icon: <PeopleIcon />, gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)' },
  ];

  useEffect(() => {
    Promise.all([
      getEnvironmentsApi({ page: 1, pageSize: 1 }).catch(() => null),
      getUsersApi({ page: 1, pageSize: 1 }).catch(() => null),
    ]).then(([envRes, userRes]) => {
      if (envRes?.data) setEnvCount(envRes.data.total ?? 0);
      if (userRes?.data) setUserCount(userRes.data.total ?? 0);
      setLoading(false);
    });
  }, []);

  const statValues = [envCount, null, userCount];

  return (
    <Box>
      <Typography variant="h4" sx={{ fontWeight: 800, mb: 0.5, letterSpacing: '-0.5px' }}>
        {t.overview.title}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 4, fontSize: 15 }}>
        {t.overview.subtitle}
      </Typography>

      <Grid container spacing={3} sx={{ mb: 5 }}>
        {statConfig.map((s, i) => (
          <Grid size={{ xs: 12, sm: 4 }} key={s.label}>
            <Card
              sx={{
                position: 'relative',
                overflow: 'hidden',
                '&::before': {
                  content: '""',
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: 4,
                  background: s.gradient,
                },
              }}
            >
              <CardContent sx={{ p: 3, '&:last-child': { pb: 3 } }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 2.5,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: s.gradient,
                      color: '#fff',
                    }}
                  >
                    {s.icon}
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 500, lineHeight: 1.2 }}>
                      {s.label}
                    </Typography>
                    {loading ? (
                      <Skeleton width={50} height={32} sx={{ mt: 0.5 }} />
                    ) : (
                      <Typography variant="h4" sx={{ fontWeight: 800, mt: 0.25 }}>
                        {statValues[i] !== null ? statValues[i]!.toLocaleString() : '—'}
                      </Typography>
                    )}
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Typography variant="h6" sx={{ fontWeight: 700, mb: 2.5, letterSpacing: '-0.3px' }}>
        {t.overview.capabilities}
      </Typography>
      <Grid container spacing={3}>
        {featureCards.map((f) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={f.title}>
            <Card
              sx={{
                height: '100%',
                transition: 'transform 0.25s, box-shadow 0.25s',
                '&:hover': {
                  transform: 'translateY(-4px)',
                  boxShadow: '0 12px 28px rgba(0,0,0,0.1)',
                },
              }}
            >
              <CardContent sx={{ p: 3, '&:last-child': { pb: 3 } }}>
                <Box
                  sx={{
                    width: 52,
                    height: 52,
                    borderRadius: 2.5,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: `${f.color}15`,
                    color: f.color,
                    mb: 2,
                  }}
                >
                  {f.icon}
                </Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 0.75 }}>
                  {f.title}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.6 }}>
                  {f.desc}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
