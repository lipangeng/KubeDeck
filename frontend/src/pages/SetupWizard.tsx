import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Container from '@mui/material/Container';
import CssBaseline from '@mui/material/CssBaseline';
import Step from '@mui/material/Step';
import StepContent from '@mui/material/StepContent';
import StepLabel from '@mui/material/StepLabel';
import Stepper from '@mui/material/Stepper';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import SettingsIcon from '@mui/icons-material/Settings';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormControl from '@mui/material/FormControl';
import FormLabel from '@mui/material/FormLabel';

const steps = [
  {
    title: '欢迎使用 KubeDeck',
    description: 'KubeDeck 是一个现代化的 Kubernetes 控制平面，帮助您轻松管理多个集群。',
  },
  {
    title: '配置数据库',
    description: '选择数据库类型并配置连接信息。',
  },
  {
    title: '配置认证方式',
    description: '选择认证方式：本地账户、OAuth2 或两者都启用。',
  },
  {
    title: '创建管理员账号',
    description: '设置管理员账户用于首次登录。',
  },
  {
    title: '完成',
    description: '初始化完成，开始使用 KubeDeck。',
  },
];

export function SetupWizard() {
  const navigate = useNavigate();
  const [activeStep, setActiveStep] = useState(0);
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    dbType: 'sqlite',
    dbHost: '',
    dbName: '',
    authMode: 'local', // 'local', 'oauth2', 'both'
    oauthProvider: 'github',
    oauthClientId: '',
    oauthClientSecret: '',
    adminUsername: '',
    adminPassword: '',
    adminEmail: '',
  });

  const handleNext = () => {
    const nextStep = activeStep + 1;
    if (nextStep >= steps.length) {
      saveConfiguration();
      navigate('/login');
    } else {
      setActiveStep(nextStep);
    }
  };

  const handleBack = () => {
    setActiveStep((prevStep) => prevStep - 1);
  };

  const saveConfiguration = () => {
    console.log('Saving configuration:', formData);
    localStorage.setItem('setupCompleted', 'true');
    // TODO: Call API to save configuration
  };

  return (
    <Container component="main" maxWidth="md">
      <CssBaseline />
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          py: 4,
        }}
      >
        <Card variant="outlined" sx={{ width: '100%', mb: 3 }}>
          <CardContent sx={{ p: 3 }}>
            <Stack spacing={2} alignItems="center">
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
                <SettingsIcon sx={{ color: 'white', fontSize: 32 }} />
              </Box>
              <Typography component="h1" variant="h4" sx={{ fontWeight: 700 }}>
                KubeDeck 初始化向导
              </Typography>
              <Typography variant="body2" color="text.secondary">
                完成以下步骤以开始使用
              </Typography>
            </Stack>
          </CardContent>
        </Card>

        <Card variant="outlined" sx={{ width: '100%' }}>
          <CardContent sx={{ p: 3 }}>
            <Stepper activeStep={activeStep} orientation="vertical">
              {steps.map((step, index) => (
                <Step key={step.title}>
                  <StepLabel>{step.title}</StepLabel>
                  <StepContent>
                    <Typography variant="body2" sx={{ mb: 2 }}>
                      {step.description}
                    </Typography>

                    {/* Step 1: Database */}
                    {index === 1 && (
                      <Stack spacing={2} sx={{ mb: 2 }}>
                        <TextField
                          select
                          fullWidth
                          label="数据库类型"
                          value={formData.dbType}
                          onChange={(e) =>
                            setFormData({ ...formData, dbType: e.target.value })
                          }
                          SelectProps={{ native: true }}
                        >
                          <option value="sqlite">SQLite (推荐)</option>
                          <option value="mysql">MySQL</option>
                          <option value="postgres">PostgreSQL</option>
                        </TextField>

                        {formData.dbType !== 'sqlite' && (
                          <>
                            <TextField
                              fullWidth
                              label="主机地址"
                              value={formData.dbHost}
                              onChange={(e) =>
                                setFormData({ ...formData, dbHost: e.target.value })
                              }
                              placeholder="localhost:3306"
                            />
                            <TextField
                              fullWidth
                              label="数据库名称"
                              value={formData.dbName}
                              onChange={(e) =>
                                setFormData({ ...formData, dbName: e.target.value })
                              }
                              placeholder="kubedeck"
                            />
                          </>
                        )}
                      </Stack>
                    )}

                    {/* Step 2: Auth Mode */}
                    {index === 2 && (
                      <Stack spacing={3} sx={{ mb: 2 }}>
                        <FormControl component="fieldset">
                          <FormLabel component="legend">认证方式</FormLabel>
                          <RadioGroup
                            value={formData.authMode}
                            onChange={(e) =>
                              setFormData({ ...formData, authMode: e.target.value })
                            }
                          >
                            <FormControlLabel
                              value="local"
                              control={<Radio />}
                              label="本地账户认证（用户名/密码）"
                            />
                            <FormControlLabel
                              value="oauth2"
                              control={<Radio />}
                              label="OAuth2 认证（GitHub/Google）"
                            />
                            <FormControlLabel
                              value="both"
                              control={<Radio />}
                              label="两者都启用"
                            />
                          </RadioGroup>
                        </FormControl>

                        {(formData.authMode === 'oauth2' || formData.authMode === 'both') && (
                          <Stack spacing={2} sx={{ mt: 2, pl: 2, borderLeft: 2, borderColor: 'divider' }}>
                            <TextField
                              select
                              fullWidth
                              label="OAuth2 提供商"
                              value={formData.oauthProvider}
                              onChange={(e) =>
                                setFormData({ ...formData, oauthProvider: e.target.value })
                              }
                              SelectProps={{ native: true }}
                            >
                              <option value="github">GitHub</option>
                              <option value="google">Google</option>
                              <option value="gitlab">GitLab</option>
                            </TextField>

                            <TextField
                              fullWidth
                              label="Client ID"
                              value={formData.oauthClientId}
                              onChange={(e) =>
                                setFormData({ ...formData, oauthClientId: e.target.value })
                              }
                              placeholder="输入 Client ID"
                            />

                            <TextField
                              fullWidth
                              label="Client Secret"
                              type="password"
                              value={formData.oauthClientSecret}
                              onChange={(e) =>
                                setFormData({ ...formData, oauthClientSecret: e.target.value })
                              }
                              placeholder="输入 Client Secret"
                            />

                            <Typography variant="caption" color="text.secondary">
                              Callback URL: http://localhost:8080/api/auth/callback
                            </Typography>
                          </Stack>
                        )}
                      </Stack>
                    )}

                    {/* Step 3: Admin Account */}
                    {index === 3 && (
                      <Stack spacing={2} sx={{ mb: 2 }}>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                          创建管理员账户用于首次登录系统。
                        </Typography>

                        <TextField
                          fullWidth
                          label="用户名"
                          value={formData.adminUsername}
                          onChange={(e) =>
                            setFormData({ ...formData, adminUsername: e.target.value })
                          }
                          placeholder="admin"
                          required
                        />

                        <TextField
                          fullWidth
                          label="邮箱"
                          type="email"
                          value={formData.adminEmail}
                          onChange={(e) =>
                            setFormData({ ...formData, adminEmail: e.target.value })
                          }
                          placeholder="admin@example.com"
                          required
                        />

                        <TextField
                          fullWidth
                          label="密码"
                          type={showPassword ? 'text' : 'password'}
                          value={formData.adminPassword}
                          onChange={(e) =>
                            setFormData({ ...formData, adminPassword: e.target.value })
                          }
                          placeholder="至少 8 位"
                          required
                          InputProps={{
                            endAdornment: (
                              <InputAdornment position="end">
                                <IconButton
                                  onClick={() => setShowPassword(!showPassword)}
                                  edge="end"
                                >
                                  {showPassword ? <VisibilityOff /> : <Visibility />}
                                </IconButton>
                              </InputAdornment>
                            ),
                          }}
                        />

                        {formData.adminPassword && formData.adminPassword.length < 8 && (
                          <Typography variant="caption" color="error">
                            密码长度至少 8 位
                          </Typography>
                        )}
                      </Stack>
                    )}

                    {/* Step 4: Complete */}
                    {index === 4 && (
                      <Box sx={{ bgcolor: 'success.light', p: 2, borderRadius: 1, mb: 2 }}>
                        <Typography variant="body2" color="success.dark">
                          ✓ 配置已完成！点击"开始使用"进入登录页面。
                        </Typography>
                      </Box>
                    )}

                    <Box sx={{ mb: 2 }}>
                      <Button
                        variant="contained"
                        onClick={handleNext}
                        sx={{ mt: 1, mr: 1 }}
                        disabled={
                          (index === 3 && (!formData.adminUsername || !formData.adminPassword || formData.adminPassword.length < 8))
                        }
                      >
                        {index === steps.length - 1 ? '开始使用' : '下一步'}
                      </Button>
                      <Button
                        variant="outlined"
                        onClick={handleBack}
                        disabled={index === 0}
                        sx={{ mt: 1, mr: 1 }}
                      >
                        上一步
                      </Button>
                    </Box>
                  </StepContent>
                </Step>
              ))}
            </Stepper>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
}
