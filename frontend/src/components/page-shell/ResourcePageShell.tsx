import type { ReactNode } from 'react';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

interface BaseShellProps {
  title: string;
  children: ReactNode;
}

interface ListPageShellProps extends BaseShellProps {
  toolbar?: ReactNode;
}

interface DetailPageShellProps extends BaseShellProps {
  summary?: ReactNode;
  sidePanel?: ReactNode;
}

export function ListPageShell({ title, toolbar, children }: ListPageShellProps) {
  return (
    <Paper variant="outlined" sx={{ p: 2 }}>
      <Stack spacing={1.5}>
        <Stack direction="row" spacing={1.5} alignItems="center" justifyContent="space-between">
          <Typography variant="h6" component="h2">
            {title}
          </Typography>
          {toolbar ? <Box>{toolbar}</Box> : null}
        </Stack>
        <Box>{children}</Box>
      </Stack>
    </Paper>
  );
}

export function DetailPageShell({
  title,
  summary,
  sidePanel,
  children,
}: DetailPageShellProps) {
  return (
    <Stack spacing={2}>
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="h6" component="h2" sx={{ mb: summary ? 1 : 0 }}>
          {title}
        </Typography>
        {summary ? <Box>{summary}</Box> : null}
      </Paper>
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: sidePanel ? { xs: '1fr', md: 'minmax(0, 1fr) 280px' } : '1fr',
          gap: 2,
        }}
      >
        <Paper variant="outlined" sx={{ p: 2 }}>
          {children}
        </Paper>
        {sidePanel ? (
          <Paper variant="outlined" sx={{ p: 2 }}>
            {sidePanel}
          </Paper>
        ) : null}
      </Box>
    </Stack>
  );
}
