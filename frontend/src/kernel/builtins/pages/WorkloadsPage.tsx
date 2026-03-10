import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { ListPageShell } from '../../../components/page-shell/ResourcePageShell';
import { copy } from '../../../i18n/copy';

export function WorkloadsPage() {
  return (
    <ListPageShell
      title={copy('workloads.title')}
      toolbar={<Chip color="primary" label={copy('workloads.badge')} size="small" />}
    >
      <Stack spacing={1.5}>
        <Typography color="text.secondary">{copy('workloads.description')}</Typography>
        <Typography variant="body2" color="text.secondary">
          {copy('workloads.placeholder')}
        </Typography>
      </Stack>
    </ListPageShell>
  );
}
