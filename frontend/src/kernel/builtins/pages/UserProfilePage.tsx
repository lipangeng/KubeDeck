import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Avatar from '@mui/material/Avatar';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Divider from '@mui/material/Divider';

export function UserProfilePage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  
  const [user, setUser] = useState({
    name: 'Admin User',
    email: 'admin@kubedeck.io',
    username: 'admin',
  });

  const [passwords, setPasswords] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [passwordError, setPasswordError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setSuccessMessage('');
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      setSuccessMessage('个人资料已更新');
    } catch (error) {
      setPasswordError('更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setSuccessMessage('');
    
    if (passwords.newPassword !== passwords.confirmPassword) {
      setPasswordError('两次输入的新密码不一致');
      return;
    }
    
    if (passwords.newPassword.length < 6) {
      setPasswordError('新密码长度至少 6 位');
      return;
    }
    
    setLoading(true);
    
    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      setSuccessMessage('密码已修改');
      setPasswords({ currentPassword: '', newPassword: '', confirmPassword: '' });
    } catch (error) {
      setPasswordError('修改失败');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await fetch('/api/auth/logout', { method: 'POST' });
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const initials = user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);

  return (
    <Stack spacing={3} sx={{ maxWidth: 800 }}>
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center">
          <Avatar sx={{ width: 56, height: 56, bgcolor: 'primary.main', fontSize: 24 }}>
            {initials}
          </Avatar>
          <Stack spacing={0.5}>
            <Typography variant="h6">{user.name}</Typography>
            <Typography variant="body2" color="text.secondary">{user.email}</Typography>
          </Stack>
        </Stack>
      </Paper>

      <Paper variant="outlined" sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>个人资料</Typography>
        <form onSubmit={handleProfileUpdate}>
          <Stack spacing={3}>
            <TextField
              fullWidth
              label="用户名"
              value={user.username}
              disabled
              helperText="用户名不可修改"
            />
            <TextField
              fullWidth
              label="姓名"
              value={user.name}
              onChange={(e) => setUser({ ...user, name: e.target.value })}
            />
            <TextField
              fullWidth
              label="邮箱"
              type="email"
              value={user.email}
              onChange={(e) => setUser({ ...user, email: e.target.value })}
            />
            {successMessage && (
              <Alert severity="success">
                <AlertTitle>成功</AlertTitle>
                {successMessage}
              </Alert>
            )}
            {passwordError && (
              <Alert severity="error">
                <AlertTitle>错误</AlertTitle>
                {passwordError}
              </Alert>
            )}
            <Box>
              <Button type="submit" variant="contained" disabled={loading}>
                {loading ? '保存中...' : '保存修改'}
              </Button>
            </Box>
          </Stack>
        </form>
      </Paper>

      <Paper variant="outlined" sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>修改密码</Typography>
        <form onSubmit={handlePasswordChange}>
          <Stack spacing={3}>
            <TextField
              fullWidth
              label="当前密码"
              type="password"
              value={passwords.currentPassword}
              onChange={(e) => setPasswords({ ...passwords, currentPassword: e.target.value })}
              required
            />
            <TextField
              fullWidth
              label="新密码"
              type="password"
              value={passwords.newPassword}
              onChange={(e) => setPasswords({ ...passwords, newPassword: e.target.value })}
              required
              helperText="至少 6 位"
            />
            <TextField
              fullWidth
              label="确认新密码"
              type="password"
              value={passwords.confirmPassword}
              onChange={(e) => setPasswords({ ...passwords, confirmPassword: e.target.value })}
              required
              error={passwords.newPassword !== passwords.confirmPassword && passwords.confirmPassword !== ''}
              helperText={passwords.newPassword !== passwords.confirmPassword && passwords.confirmPassword !== '' ? '两次输入的密码不一致' : ''}
            />
            <Box>
              <Button 
                type="submit" 
                variant="contained" 
                color="secondary"
                disabled={loading || !passwords.currentPassword || !passwords.newPassword || !passwords.confirmPassword}
              >
                {loading ? '修改中...' : '修改密码'}
              </Button>
            </Box>
          </Stack>
        </form>
      </Paper>

      <Paper variant="outlined" sx={{ p: 3, borderColor: 'error.light' }}>
        <Typography variant="h6" color="error" sx={{ mb: 2 }}>危险区域</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          退出登录后需要重新登录才能访问系统。
        </Typography>
        <Button variant="outlined" color="error" onClick={handleLogout}>
          退出登录
        </Button>
      </Paper>
    </Stack>
  );
}
