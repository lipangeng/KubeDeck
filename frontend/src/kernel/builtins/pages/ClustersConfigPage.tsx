import { useState, useEffect } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import RefreshIcon from '@mui/icons-material/Refresh';
import UploadIcon from '@mui/icons-material/Upload';

interface Cluster {
  id: string;
  name: string;
  server: string;
  status: 'connected' | 'disconnected' | 'error';
  version?: string;
  nodes?: number;
  createdAt: string;
}

export function ClustersConfigPage() {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCluster, setEditingCluster] = useState<Cluster | null>(null);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    server: '',
    token: '',
    kubeconfig: '',
  });

  useEffect(() => {
    loadClusters();
  }, []);

  const loadClusters = async () => {
    try {
      const response = await fetch('/api/clusters/config');
      const data = await response.json();
      setClusters(data.clusters || []);
    } catch (error) {
      console.error('Failed to load clusters:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (cluster?: Cluster) => {
    if (cluster) {
      setEditingCluster(cluster);
      setFormData({
        name: cluster.name,
        server: cluster.server,
        token: '',
        kubeconfig: '',
      });
    } else {
      setEditingCluster(null);
      setFormData({
        name: '',
        server: '',
        token: '',
        kubeconfig: '',
      });
    }
    setTestResult(null);
    setDialogOpen(true);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditingCluster(null);
    setFormData({
      name: '',
      server: '',
      token: '',
      kubeconfig: '',
    });
    setTestResult(null);
  };

  const handleTestConnection = async () => {
    setTestResult(null);
    try {
      const response = await fetch('/api/clusters/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          server: formData.server,
          token: formData.token,
        }),
      });
      const result = await response.json();
      setTestResult(result);
    } catch (error) {
      setTestResult({
        success: false,
        message: '连接测试失败：' + (error as Error).message,
      });
    }
  };

  const handleSave = async () => {
    try {
      const response = await fetch('/api/clusters/config', {
        method: editingCluster ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: editingCluster?.id,
          name: formData.name,
          server: formData.server,
          token: formData.token,
          kubeconfig: formData.kubeconfig,
        }),
      });

      if (response.ok) {
        handleCloseDialog();
        loadClusters();
      }
    } catch (error) {
      console.error('Failed to save cluster:', error);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除这个集群配置吗？')) return;

    try {
      await fetch(`/api/clusters/config/${id}`, { method: 'DELETE' });
      loadClusters();
    } catch (error) {
      console.error('Failed to delete cluster:', error);
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setFormData({ ...formData, kubeconfig: content });
        // Try to parse kubeconfig
        try {
          // Simple parsing - in production use js-yaml
          const nameMatch = content.match(/name:\s*(.+)/);
          const serverMatch = content.match(/server:\s*(.+)/);
          if (nameMatch) {
            setFormData({ ...formData, name: nameMatch[1].trim() });
          }
          if (serverMatch) {
            setFormData({ ...formData, server: serverMatch[1].trim() });
          }
        } catch (error) {
          console.error('Failed to parse kubeconfig:', error);
        }
      };
      reader.readAsText(file);
    }
  };

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              集群配置
            </Typography>
            <Typography variant="body2" color="text.secondary">
              管理 Kubernetes 集群连接配置
            </Typography>
          </Stack>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => handleOpenDialog()}>
            添加集群
          </Button>
        </Stack>
      </Paper>

      {/* Info Alert */}
      <Alert severity="info">
        <AlertTitle>提示</AlertTitle>
        支持两种添加方式：
        <br />
        1. 上传 kubeconfig 文件（推荐）
        <br />
        2. 手动输入 API Server 地址和 Token
      </Alert>

      {/* Clusters Table */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
          已配置的集群 ({clusters.length})
        </Typography>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>名称</TableCell>
              <TableCell>API Server</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>版本</TableCell>
              <TableCell>节点数</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {clusters.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  <Typography color="text.secondary" sx={{ py: 4 }}>
                    暂无集群配置，点击右上角"添加集群"开始配置
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              clusters.map((cluster) => (
                <TableRow key={cluster.id}>
                  <TableCell>
                    <Typography fontWeight={600}>{cluster.name}</Typography>
                  </TableCell>
                  <TableCell>
                    <Typography fontFamily="monospace" fontSize="0.875rem">
                      {cluster.server}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      icon={cluster.status === 'connected' ? <CheckCircleIcon /> : <ErrorIcon />}
                      label={cluster.status === 'connected' ? '已连接' : '未连接'}
                      color={cluster.status === 'connected' ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{cluster.version || '-'}</TableCell>
                  <TableCell>{cluster.nodes || '-'}</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5}>
                      <IconButton
                        size="small"
                        onClick={() => handleOpenDialog(cluster)}
                        aria-label="Edit cluster"
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => handleDelete(cluster.id)}
                        color="error"
                        aria-label="Delete cluster"
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>

      {/* Add/Edit Dialog */}
      <Dialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingCluster ? '编辑集群' : '添加集群'}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            {/* Kubeconfig Upload */}
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                方式一：上传 kubeconfig 文件
              </Typography>
              <Button
                variant="outlined"
                component="label"
                startIcon={<UploadIcon />}
                fullWidth
              >
                选择文件
                <input
                  type="file"
                  hidden
                  accept=".yaml,.yml,.kubeconfig"
                  onChange={handleFileUpload}
                />
              </Button>
            </Box>

            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Box sx={{ flex: 1, height: 1, bgcolor: 'divider', mr: 2 }} />
              <Typography variant="body2" color="text.secondary">或</Typography>
              <Box sx={{ flex: 1, height: 1, bgcolor: 'divider', ml: 2 }} />
            </Box>

            {/* Manual Configuration */}
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                方式二：手动配置
              </Typography>
              <Stack spacing={2}>
                <TextField
                  fullWidth
                  label="集群名称"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="my-cluster"
                  required
                />

                <TextField
                  fullWidth
                  label="API Server 地址"
                  value={formData.server}
                  onChange={(e) => setFormData({ ...formData, server: e.target.value })}
                  placeholder="https://192.168.1.100:6443"
                  required
                />

                <TextField
                  fullWidth
                  label="Service Account Token"
                  value={formData.token}
                  onChange={(e) => setFormData({ ...formData, token: e.target.value })}
                  placeholder="eyJhbGciOiJSUzI1NiIsImtpZCI6Ii..."
                  multiline
                  rows={3}
                  required
                />
              </Stack>
            </Box>

            {/* Test Connection */}
            {formData.server && (
              <Box>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                  <Typography variant="subtitle2">连接测试</Typography>
                  <Button
                    size="small"
                    variant="outlined"
                    startIcon={<RefreshIcon />}
                    onClick={handleTestConnection}
                    disabled={!formData.server}
                  >
                    测试连接
                  </Button>
                </Stack>

                {testResult && (
                  <Alert severity={testResult.success ? 'success' : 'error'}>
                    {testResult.message}
                  </Alert>
                )}
              </Box>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>取消</Button>
          <Button
            onClick={handleSave}
            variant="contained"
            disabled={!formData.name || !formData.server}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
