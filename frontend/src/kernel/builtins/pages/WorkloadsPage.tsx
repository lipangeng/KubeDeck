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
import { ListPageShell } from '../../../components/page-shell/ResourcePageShell';
import { copy } from '../../../i18n/copy';
import { ResourcePageShell } from '../../resource-pages/ResourcePageShell';
import { resolveResourcePage } from '../../resource-pages/resolveResourcePage';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';
import type { KernelActionExecutionResult } from '../../runtime/executeKernelAction';
import type { WorkloadItem } from '../../runtime/fetchWorkloads';

export function WorkloadsPage() {
  const {
    activePage,
    activeSummarySlots,
    currentResource,
    executeAction,
    enterResource,
    exitResource,
    fetchWorkloadsForDomain,
    resourcePageExtensions,
  } = useKernelRuntime();
  const [items, setItems] = useState<WorkloadItem[]>([]);
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
        const nextItems = await fetchWorkloadsForDomain(currentWorkflowDomainId, 'default');
        if (active) {
          setItems(nextItems);
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
  }, [fetchWorkloadsForDomain, workflowDomainId]);

  if (currentResource) {
    const resolvedPage = resolveResourcePage({
      resource: currentResource,
      overviewContent: <Typography>Resource overview for {currentResource.name}</Typography>,
      yamlContent: (
        <Typography component="pre" sx={{ m: 0, fontFamily: 'monospace' }}>
          {`apiVersion: apps/v1\nkind: ${currentResource.kind}\nmetadata:\n  name: ${currentResource.name}\n  namespace: ${currentResource.namespace ?? 'default'}`}
        </Typography>
      ),
      yamlVariantContent: (
        <Typography component="pre" sx={{ m: 0, fontFamily: 'monospace' }}>
          {`Deployment YAML v2 for ${currentResource.name}`}
        </Typography>
      ),
      runtimeContent: (
        <Typography color="text.secondary">
          Runtime status and rollout details for {currentResource.name}
        </Typography>
      ),
      logsContent: (
        <Typography component="pre" sx={{ m: 0, fontFamily: 'monospace' }}>
          {`${currentResource.name}: application logs stream preview`}
        </Typography>
      ),
      extensions: resourcePageExtensions,
    });

    const handleAction = async (actionId: string) => {
      const result = await executeAction({
        actionId,
        workflowDomainId: activePage?.workflowDomainId ?? 'workloads',
        target: {
          cluster: 'default',
          namespace: currentResource.namespace ?? 'default',
          scope: currentResource.namespace ? 'namespace' : 'cluster',
        },
        input: {
          name: currentResource.name,
        },
      });
      setActionResult(result);
    };

    return (
      <Stack spacing={2}>
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
        <ResourcePageShell
          title={`${currentResource.kind}/${currentResource.name}`}
          summary={
            <Typography color="text.secondary">
              Namespace: {currentResource.namespace ?? 'cluster'}
            </Typography>
          }
          actions={
            <Stack direction="row" spacing={1}>
              <Button variant="contained" onClick={() => void handleAction('apply')}>
                Apply
              </Button>
              <Button variant="outlined" onClick={() => void handleAction('create')}>
                Create
              </Button>
              <Button variant="text" onClick={exitResource}>
                Back to Workloads
              </Button>
            </Stack>
          }
          tabs={resolvedPage.tabs}
          takeoverContent={resolvedPage.takeoverContent}
        />
      </Stack>
    );
  }

  return (
    <ListPageShell
      title={copy('workloads.title')}
      toolbar={<Chip color="primary" label={copy('workloads.badge')} size="small" />}
      summary={
        activeSummarySlots.length > 0 ? (
          <Stack spacing={1}>
            {activeSummarySlots.map((slot) => {
              const SlotComponent = slot.component;
              return <SlotComponent key={slot.identity.contributionId} />;
            })}
          </Stack>
        ) : null
      }
    >
      <Stack spacing={1.5}>
        <Typography color="text.secondary">{copy('workloads.description')}</Typography>
        {loading ? (
          <Stack direction="row" spacing={1} alignItems="center">
            <CircularProgress size={18} />
            <Typography variant="body2" color="text.secondary">
              {copy('workloads.loading')}
            </Typography>
          </Stack>
        ) : (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>{copy('workloads.columns.name')}</TableCell>
                <TableCell>{copy('workloads.columns.kind')}</TableCell>
                <TableCell>{copy('workloads.columns.namespace')}</TableCell>
                <TableCell>{copy('workloads.columns.status')}</TableCell>
                <TableCell>{copy('workloads.columns.health')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <Button
                      variant="text"
                      onClick={() =>
                        enterResource({
                          kind: item.kind,
                          name: item.name,
                          namespace: item.namespace,
                        })
                      }
                    >
                      {item.name}
                    </Button>
                  </TableCell>
                  <TableCell>{item.kind}</TableCell>
                  <TableCell>{item.namespace}</TableCell>
                  <TableCell>{item.status}</TableCell>
                  <TableCell>{item.health}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        <Typography variant="body2" color="text.secondary">
          {copy('workloads.placeholder')}
        </Typography>
      </Stack>
    </ListPageShell>
  );
}
