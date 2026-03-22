import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import { copy } from '../../../i18n/copy';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';

export function HomepagePage() {
  const { activeCluster, namespaceScope } = useKernelRuntime();

  return (
    <Paper variant="outlined" sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="overline" color="primary.main">
          {copy('app.cleanup.badge')}
        </Typography>
        <Typography variant="h4" sx={{ fontWeight: 700 }}>
          {copy('homepage.title')}
        </Typography>
        <Typography color="text.secondary">{copy('homepage.description')}</Typography>

        <Box sx={{ mt: 2, p: 2, bgcolor: 'background.paper', borderRadius: 1 }}>
          <Stack spacing={1.5}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {copy('homepage.currentContext')}
            </Typography>
            <Stack direction="row" spacing={3}>
              <Box>
                <Typography variant="subtitle2" color="text.secondary">
                  {copy('homepage.activeCluster')}
                </Typography>
                <Typography variant="body1" sx={{ fontWeight: 500 }}>
                  {activeCluster}
                </Typography>
              </Box>
              <Box>
                <Typography variant="subtitle2" color="text.secondary">
                  {copy('homepage.namespaceScope')}
                </Typography>
                <Typography variant="body1" sx={{ fontWeight: 500 }}>
                  {namespaceScope.kind === 'single'
                    ? namespaceScope.namespaces.join(', ')
                    : 'All Namespaces'}
                </Typography>
              </Box>
            </Stack>
          </Stack>
        </Box>

        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          {copy('homepage.useNavigation')}
        </Typography>
      </Stack>
    </Paper>
  );
}
