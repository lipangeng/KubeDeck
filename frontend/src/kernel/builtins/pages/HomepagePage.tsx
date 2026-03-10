import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { copy } from '../../../i18n/copy';

export function HomepagePage() {
  return (
    <Paper variant="outlined" sx={{ p: 3 }}>
      <Stack spacing={1.5}>
        <Typography variant="overline" color="primary.main">
          {copy('app.cleanup.badge')}
        </Typography>
        <Typography variant="h4" sx={{ fontWeight: 700 }}>
          {copy('homepage.title')}
        </Typography>
        <Typography color="text.secondary">{copy('homepage.description')}</Typography>
      </Stack>
    </Paper>
  );
}
