import Chip from '@mui/material/Chip';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { copy } from '../../i18n/copy';
import type { LocalizedText } from '../contracts/types';

export function createRemoteCapabilityPage(
  title: LocalizedText,
  description?: LocalizedText,
) {
  function RemoteCapabilityPage() {
    return (
      <Paper variant="outlined" sx={{ p: 3 }}>
        <Stack spacing={1.5}>
          <Chip color="secondary" label={copy('remoteCapability.badge')} size="small" />
          <Typography variant="h4" sx={{ fontWeight: 700 }}>
            {title.fallback}
          </Typography>
          {description ? (
            <Typography color="text.secondary">{description.fallback}</Typography>
          ) : null}
          <Typography variant="body2" color="text.secondary">
            {copy('remoteCapability.body')}
          </Typography>
        </Stack>
      </Paper>
    );
  }

  return RemoteCapabilityPage;
}
