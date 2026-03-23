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
import Box from '@mui/material/Box';

interface HelmRelease {
  name: string;
  namespace: string;
  revision: number;
  updated: string;
  status: string;
  chart: string;
  app_version: string;
}

export function HelmPage() {
  const [releases, setReleases] = useState<HelmRelease[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadReleases();
  }, []);

  const loadReleases = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/helm?action=list&namespace=default');
      const data = await response.json();
      setReleases(data.releases || []);
      if (data.error) {
      }
    } catch (error) {
    } finally {
      setLoading(false);
    }
  };

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
              Helm Charts
            </Typography>
            <Typography variant="body2" color="text.secondary">
              管理 Helm Chart 发布
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" onClick={loadReleases}>刷新</Button>
            <Button variant="contained">安装 Chart</Button>
          </Stack>
        </Stack>
      </Paper>

      {/* Releases Table */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
          已安装的发布 ({releases.length})
        </Typography>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>名称</TableCell>
              <TableCell>命名空间</TableCell>
              <TableCell>版本</TableCell>
              <TableCell>Chart</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {releases.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  <Typography color="text.secondary" sx={{ py: 4 }}>
                    暂无 Helm 发布
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              releases.map((release, index) => (
                <TableRow key={index} hover>
                  <TableCell>
                    <Typography fontWeight={600}>{release.name}</Typography>
                  </TableCell>
                  <TableCell>{release.namespace}</TableCell>
                  <TableCell>{release.revision}</TableCell>
                  <TableCell>
                    <Typography variant="body2" fontFamily="monospace">
                      {release.chart}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={release.status} 
                      size="small" 
                      color={release.status === 'deployed' ? 'success' : 'default'} 
                    />
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5}>
                      <Button size="small">升级</Button>
                      <Button size="small" color="error">卸载</Button>
                    </Stack>
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
