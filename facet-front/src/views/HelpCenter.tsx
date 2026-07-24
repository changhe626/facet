import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Alert from '@mui/material/Alert';
import BusinessIcon from '@mui/icons-material/Business';
import ToggleOnIcon from '@mui/icons-material/ToggleOn';
import ExploreIcon from '@mui/icons-material/Explore';
import AutoFixHighIcon from '@mui/icons-material/AutoFixHigh';
import { useLocaleStore } from '@/stores/localeStore';

export default function HelpCenter() {
  const { t } = useLocaleStore();
  const cards = [
    { icon: <BusinessIcon fontSize="large" />, title: t.help.multiTenant, desc: t.help.multiTenantDesc },
    { icon: <ToggleOnIcon fontSize="large" />, title: t.help.featureFlags, desc: t.help.featureFlagsDesc },
    { icon: <ExploreIcon fontSize="large" />, title: t.help.gettingStarted, desc: t.help.gettingStartedDesc },
    { icon: <AutoFixHighIcon fontSize="large" />, title: t.help.bestPractices, desc: t.help.bestPracticesDesc },
  ];
  const concepts = [
    { term: t.help.conceptTenant, def: t.help.conceptTenantDesc },
    { term: t.help.conceptEnvironment, def: t.help.conceptEnvironmentDesc },
    { term: t.help.conceptFlag, def: t.help.conceptFlagDesc },
    { term: t.help.conceptSdkKey, def: t.help.conceptSdkKeyDesc },
  ];
  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 600, mb: 3 }}>{t.help.title}</Typography>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        {cards.map((s) => (
          <Grid size={{ xs: 12, sm: 6 }} key={s.title}>
            <Card sx={{ height: '100%' }} elevation={1}>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}>
                  {s.icon}
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>{s.title}</Typography>
                </Box>
                <Typography variant="body2" color="text.secondary">{s.desc}</Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>{t.help.coreConcepts}</Typography>
      <Card sx={{ mb: 3 }} elevation={1}>
        <CardContent>
          {concepts.map((c, i) => (
            <Typography key={c.term} variant="body2" sx={{ mb: i < concepts.length - 1 ? 1 : 0 }}>
              <strong>{c.term}</strong> — {c.def}
            </Typography>
          ))}
        </CardContent>
      </Card>

      <Alert severity="info" sx={{ mb: 1 }}>
        {t.help.tipEnv}
      </Alert>
      <Alert severity="success">
        {t.help.tipFlagType}
      </Alert>
    </Box>
  );
}
