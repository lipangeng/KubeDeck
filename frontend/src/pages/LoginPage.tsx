import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Container from '@mui/material/Container';
import CssBaseline from '@mui/material/CssBaseline';
import Link from '@mui/material/Link';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import GitHubIcon from '@mui/icons-material/GitHub';
import GoogleIcon from '@mui/icons-material/Google';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';

export function LoginPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);

  const handleOAuth2Login = (provider: string) => {
    setLoading(true);
    // Redirect to OAuth2 provider
    window.location.href = `/api/auth/login?provider=${provider}`;
  };

  return (
    <Container component="main" maxWidth="xs">
      <CssBaseline />
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Card variant="outlined" sx={{ width: '100%' }}>
          <CardContent sx={{ p: 4 }}>
            <Stack spacing={3} alignItems="center">
              {/* Logo */}
              <Box
                sx={{
                  width: 64,
                  height: 64,
                  borderRadius: 2,
                  bgcolor: 'primary.main',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <LockOutlinedIcon sx={{ color: 'white', fontSize: 32 }} />
              </Box>

              {/* Title */}
              <Typography component="h1" variant="h4" sx={{ fontWeight: 700 }}>
                KubeDeck
              </Typography>

              <Typography variant="body2" color="text.secondary" align="center">
                Kubernetes 控制平面
              </Typography>

              {/* OAuth2 Buttons */}
              <Stack spacing={2} sx={{ width: '100%', mt: 2 }}>
                <Button
                  fullWidth
                  variant="outlined"
                  size="large"
                  startIcon={<GitHubIcon />}
                  onClick={() => handleOAuth2Login('github')}
                  disabled={loading}
                  sx={{ py: 1.5 }}
                >
                  使用 GitHub 登录
                </Button>

                <Button
                  fullWidth
                  variant="outlined"
                  size="large"
                  startIcon={<GoogleIcon />}
                  onClick={() => handleOAuth2Login('google')}
                  disabled={loading}
                  sx={{ py: 1.5 }}
                >
                  使用 Google 登录
                </Button>
              </Stack>

              {/* Divider */}
              <Box sx={{ width: '100%', textAlign: 'center' }}>
                <Typography variant="body2" color="text.secondary">
                  通过 OAuth2 安全登录
                </Typography>
              </Box>

              {/* Footer */}
              <Box sx={{ mt: 3, textAlign: 'center' }}>
                <Typography variant="caption" color="text.secondary">
                  登录即表示您同意我们的{' '}
                  <Link href="#" underline="hover">
                    服务条款
                  </Link>
                  {' '}和{' '}
                  <Link href="#" underline="hover">
                    隐私政策
                  </Link>
                </Typography>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        {/* Version */}
        <Typography variant="caption" color="text.secondary" sx={{ mt: 3 }}>
          KubeDeck v0.9.0
        </Typography>
      </Box>
    </Container>
  );
}
