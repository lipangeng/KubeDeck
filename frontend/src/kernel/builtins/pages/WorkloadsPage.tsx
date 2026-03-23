import { useEffect, useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Chip from '@mui/material/Chip';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import AddIcon from '@mui/icons-material/Add';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';

interface WorkloadItem {
  id: string;
  name: string;
  kind: string;
  namespace: string;
  status: string;
  health: string;
  image?: string;
  updated_at: string;
}

export function WorkloadsPage() {
  const { navigate, activeCluster } = useKernelRuntime();
  const [workloads, setWorkloads] = useState<WorkloadItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');

  useEffect(() => {
    loadWorkloads();
  }, [activeCluster]);

  const loadWorkloads = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/workflows/workloads/items?workflowDomainId=workloads&cluster=${activeCluster}`);
      const data = await response.json();
      setWorkloads(data || []);
    } catch (error) {
    } finally {
      setLoading(false);
    }
  };

  const getHealthColor = (health: string) => {
    switch (health) {
      case 'Healthy': return 'success';
      case 'Warning': return 'warning';
      case 'Error': return 'error';
      default: return 'default';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Running': return 'success';
      case 'Pending': return 'warning';
      case 'Failed': return 'error';
      default: return 'default';
    }
  };

  const filteredWorkloads = workloads.filter(workload =>
    workload.name?.toLowerCase().includes(filter.toLowerCase()) ||
    workload.namespace?.toLowerCase().includes(filter.toLowerCase()) ||
    workload.kind?.toLowerCase().includes(filter.toLowerCase())
  );

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              工作负载
            </Typography>
            <Typography variant="body2" color="text.secondary">
              管理 Kubernetes 工作负载
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <IconButton onClick={loadWorkloads} color="primary">
              <RefreshIcon />
            </IconButton>
            <Button variant="contained" startIcon={<AddIcon />}>
              创建
            </Button>
          </Stack>
        </Stack>
      </Paper>

      {/* Filters */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center">
          <TextField
            size="small"
            placeholder="搜索工作负载..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
            sx={{ flexGrow: 1 }}
          />
          <Typography variant="body2" color="text.secondary">
            共 {filteredWorkloads.length} 个
          </Typography>
        </Stack>
      </Paper>

      {/* Workloads Table */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>名称</TableCell>
              <TableCell>类型</TableCell>
              <TableCell>命名空间</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>健康</TableCell>
              <TableCell>更新时间</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredWorkloads.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center">
                  <Typography color="text.secondary" sx={{ py: 4 }}>
                    暂无工作负载
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              filteredWorkloads.map((workload) => (
                <TableRow key={workload.id} hover>
                  <TableCell>
                    <Typography fontWeight={600}>{workload.name}</Typography>
                  </TableCell>
                  <TableCell>
                    <Chip label={workload.kind} size="small" variant="outlined" />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2">{workload.namespace}</Typography>
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={workload.status} 
                      size="small" 
                      color={getStatusColor(workload.status)} 
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={workload.health} 
                      size="small" 
                      color={getHealthColor(workload.health)} 
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="caption" color="text.secondary">
                      {new Date(workload.updated_at).toLocaleString('zh-CN')}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Button 
                      size="small" 
                      onClick={() => navigate(`/pods/${workload.namespace}/${workload.name}`)}
                    >
                      详情
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>
    </Stack>
  );
}
