import { useEffect, useState } from 'react';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import { ListPageShell } from '../../../components/page-shell/ResourcePageShell';
import { copy } from '../../../i18n/copy';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';
import type { KernelActionExecutionResult } from '../../runtime/executeKernelAction';

export interface ClusterItem {
  id: string;
  name: string;
  status: string;
  health: string;
  version: string;
  nodes: number;
  updatedAt: string;
}

export function ClustersPage() {
  const { activeActions, activePage, executeAction, switchCluster } = useKernelRuntime();
  const [items, setItems] = useState<ClusterItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionResult, setActionResult] = useState<KernelActionExecutionResult | null>(null);
  const workflowDomainId = activePage?.workflowDomainId;

  useEffect(() => {
    if (!workflowDomainId) {
      return;
    }
    const currentWorkflowDomainId = workflowDomainId;

    let active = true;

    async function load() {
      try {
        const response = await fetch(`/api/clusters/items`);
        if (!response.ok) {
          throw new Error(`clusters request failed: ${response.status}`);
        }
        const data = await response.json();
        if (active) {
          setItems(data);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void load();
    return () => {
      active = false;
    };
  }, [workflowDomainId]);

  const handleExecuteAction = async (actionId: string, clusterName?: string) => {
    const result = await executeAction({
      actionId,
      workflowDomainId: activePage?.workflowDomainId ?? 'clusters',
      target: {
        cluster: clusterName || 'default',
        namespace: 'default',
        scope: 'cluster',
      },
      input: {
        name: clusterName || 'new-cluster',
      },
    });
    setActionResult(result);
  };

  const handleSwitchCluster = (clusterId: string) => {
    switchCluster(clusterId);
  };

  return (
    <ListPageShell
      title={copy('clusters.title')}
      toolbar={<Chip color="primary" label={copy('clusters.badge')} size="small" />}
    >
      <Stack spacing={1.5}>
        <Typography color="text.secondary">{copy('clusters.description')}</Typography>
        
        {actionResult ? (
          <Alert severity="success">
            <Stack spacing={0.5}>
              <Typography>{actionResult.Summary}</Typography>
              {actionResult.AffectedObjects.map((item) => (
                <Typography key={item} variant="body2">
                  {item}
                </Typography>
              ))}
            </Stack>
          </Alert>
        ) : null}

        {loading ? (
          <Stack direction="row" spacing={1} alignItems="center">
            <CircularProgress size={18} />
            <Typography variant="body2" color="text.secondary">
              {copy('clusters.loading')}
            </Typography>
          </Stack>
        ) : (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>{copy('clusters.columns.name')}</TableCell>
                <TableCell>{copy('clusters.columns.status')}</TableCell>
                <TableCell>{copy('clusters.columns.health')}</TableCell>
                <TableCell>{copy('clusters.columns.version')}</TableCell>
                <TableCell>{copy('clusters.columns.nodes')}</TableCell>
                <TableCell>{copy('clusters.columns.actions')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <Typography sx={{ fontWeight: 600 }}>{item.name}</Typography>
                  </TableCell>
                  <TableCell>
                    <Chip label={item.status} size="small" color={item.status === 'Ready' ? 'success' : 'warning'} />
                  </TableCell>
                  <TableCell>
                    <Chip label={item.health} size="small" color={item.health === 'Healthy' ? 'success' : item.health === 'Warning' ? 'warning' : 'error'} />
                  </TableCell>
                  <TableCell>{item.version}</TableCell>
                  <TableCell>{item.nodes}</TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5}>
                      <Button
                        variant="outlined"
                        size="small"
                        onClick={() => handleSwitchCluster(item.id)}
                      >
                        {copy('clusters.actions.switch')}
                      </Button>
                      <Button
                        variant="outlined"
                        size="small"
                        onClick={() => void handleExecuteAction('remove', item.name)}
                      >
                        {copy('actions.removeCluster')}
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}

        <Box sx={{ mt: 2 }}>
          <Stack direction="row" spacing={1}>
            {activeActions.map((action) => (
              <Button
                key={action.identity.contributionId}
                variant={action.actionId === 'add' ? 'contained' : 'outlined'}
                onClick={() => void handleExecuteAction(action.actionId)}
              >
                {action.title.fallback}
              </Button>
            ))}
          </Stack>
        </Box>

        <Typography variant="body2" color="text.secondary">
          {copy('clusters.placeholder')}
        </Typography>
      </Stack>
    </ListPageShell>
  );
}
