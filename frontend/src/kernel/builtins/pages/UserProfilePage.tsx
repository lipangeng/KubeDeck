import { useState, FormEvent } from 'react';
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
import CircularProgress from '@mui/material/CircularProgress';

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
  const [formError, setFormError] = useState('');

  const handleProfileUpdate = async (e: FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    console.log('=== 提交个人资料 ===');
    setLoading(true);
    setSuccessMessage('');
    setFormError('');
    
    try {
      // TODO: Call API to update profile
      console.log('Updating profile:', user);
      await new Promise(resolve => setTimeout(resolve, 800));
      
      setSuccessMessage('✅ 个人资料已成功更新');
      console.log('Profile updated successfully');
    } catch (error) {
      setFormError('❌ 更新失败：' + (error as Error).message);
      console.error('Profile update failed:', error);
    } finally {
      setLoading(false);
      // Clear success message after 5 seconds
      setTimeout(() => setSuccessMessage(''), 5000);
    }
  };

  const handlePasswordChange = async (e: FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    console.log('=== 提交密码修改 ===');
    setPasswordError('');
    setSuccessMessage('');
    
    // Validate passwords
    if (passwords.newPassword !== passwords.confirmPassword) {
      setPasswordError('❌ 两次输入的新密码不一致');
      return;
    }
    
    if (passwords.newPassword.length < 6) {
      setPasswordError('❌ 新密码长度至少 6 位');
      return;
    }
    
    if (!passwords.currentPassword) {
      setPasswordError('❌ 请输入当前密码');
      return;
    }
    
    setLoading(true);
    
    try {
      // TODO: Call API to change password
      console.log('Changing password:', { 
        hasCurrentPassword: !!passwords.currentPassword,
        newPasswordLength: passwords.newPassword.length 
      });
      await new Promise(resolve => setTimeout(resolve, 800));
      
      setSuccessMessage('✅ 密码已成功修改');
      setPasswords({ currentPassword: '', newPassword: '', confirmPassword: '' });
      console.log('Password changed successfully');
    } catch (error) {
      setPasswordError('❌ 修改失败：' + (error as Error).message);
      console.error('Password change failed:', error);
    } finally {
      setLoading(false);
      // Clear success message after 5 seconds
      setTimeout(() => setSuccessMessage(''), 5000);
    }
  };

  const handleLogout = async () => {
    console.log('=== 退出登录 ===');
    try {
      await fetch('/api/auth/logout', { 
        method: 'POST',
        credentials: 'include'
      });
      console.log('Logout successful');
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const initials = user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);

  return (
    <Stack spacing={3} sx={{ maxWidth: 800 }}>
      {/* Header */}
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

      {/* Profile Form */}
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
              onChange={(e) => {
                console.log('Name changed:', e.target.value);
                setUser({ ...user, name: e.target.value });
              }}
            />
            <TextField
              fullWidth
              label="邮箱"
              type="email"
              value={user.email}
              onChange={(e) => {
                console.log('Email changed:', e.target.value);
                setUser({ ...user, email: e.target.value });
              }}
            />
            
            {successMessage && (
              <Alert severity="success" onClose={() => setSuccessMessage('')}>
                <AlertTitle>成功</AlertTitle>
                {successMessage}
              </Alert>
            )}
            
            {formError && (
              <Alert severity="error" onClose={() => setFormError('')}>
                <AlertTitle>错误</AlertTitle>
                {formError}
              </Alert>
            )}
            
            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button 
                type="submit" 
                variant="contained" 
                disabled={loading}
                sx={{ minWidth: 120 }}
              >
                {loading ? <CircularProgress size={24} /> : '保存修改'}
              </Button>
              <Button 
                type="button"
                variant="outlined"
                onClick={() => {
                  setUser({
                    name: 'Admin User',
                    email: 'admin@kubedeck.io',
                    username: 'admin',
                  });
                  setFormError('');
                  setSuccessMessage('');
                }}
              >
                重置
              </Button>
            </Box>
          </Stack>
        </form>
      </Paper>

      {/* Password Change */}
      <Paper variant="outlined" sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>修改密码</Typography>
        <form onSubmit={handlePasswordChange}>
          <Stack spacing={3}>
            {successMessage && (
              <Alert severity="success" onClose={() => setSuccessMessage('')}>
                <AlertTitle>成功</AlertTitle>
                {successMessage}
              </Alert>
            )}
            
            {passwordError && (
              <Alert severity="error" onClose={() => setPasswordError('')}>
                <AlertTitle>错误</AlertTitle>
                {passwordError}
              </Alert>
            )}
            
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
            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button 
                type="submit" 
                variant="contained" 
                color="secondary"
                disabled={loading || !passwords.currentPassword || !passwords.newPassword || !passwords.confirmPassword}
                sx={{ minWidth: 120 }}
              >
                {loading ? <CircularProgress size={24} /> : '修改密码'}
              </Button>
              <Button 
                type="button"
                variant="outlined"
                onClick={() => setPasswords({ currentPassword: '', newPassword: '', confirmPassword: '' })}
                disabled={loading}
              >
                清空
              </Button>
            </Box>
          </Stack>
        </form>
      </Paper>

      {/* Danger Zone */}
      <Paper variant="outlined" sx={{ p: 3, borderColor: 'error.light' }}>
        <Typography variant="h6" color="error" sx={{ mb: 2 }}>危险区域</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          退出登录后需要重新登录才能访问系统。
        </Typography>
        <Button 
          variant="outlined" 
          color="error" 
          onClick={handleLogout}
          startIcon={loading ? <CircularProgress size={20} /> : null}
          disabled={loading}
        >
          退出登录
        </Button>
      </Paper>
    </Stack>
  );
}
