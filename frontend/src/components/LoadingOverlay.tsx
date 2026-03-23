import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

interface LoadingOverlayProps {
  message?: string;
  fullScreen?: boolean;
}

export function LoadingOverlay({ message = '加载中...', fullScreen = false }: LoadingOverlayProps) {
  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: fullScreen ? '100vh' : '200px',
        width: '100%',
      }}
    >
      <Stack spacing={2} alignItems="center">
        <CircularProgress size={40} />
        {message && (
          <Typography variant="body2" color="text.secondary">
            {message}
          </Typography>
        )}
      </Stack>
    </Box>
  );
}

export function PageLoader() {
  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '400px',
      }}
    >
      <CircularProgress />
    </Box>
  );
}
