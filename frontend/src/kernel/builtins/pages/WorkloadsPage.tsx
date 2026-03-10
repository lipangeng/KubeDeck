import { useEffect, useState } from 'react';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
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
import { resolveDefaultTabs } from '../../resource-pages/tabs';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';
import type { WorkloadItem } from '../../runtime/fetchWorkloads';

export function WorkloadsPage() {
  const {
    activePage,
    activeSummarySlots,
    currentResource,
    enterResource,
    fetchWorkloadsForDomain,
  } = useKernelRuntime();
  const [items, setItems] = useState<WorkloadItem[]>([]);
  const [loading, setLoading] = useState(true);
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
    return (
      <ResourcePageShell
        title={`${currentResource.kind}/${currentResource.name}`}
        summary={
          <Typography color="text.secondary">
            Namespace: {currentResource.namespace ?? 'cluster'}
          </Typography>
        }
        tabs={resolveDefaultTabs({
          overviewContent: <Typography>Resource overview for {currentResource.name}</Typography>,
          yamlContent: (
            <Typography component="pre" sx={{ m: 0, fontFamily: 'monospace' }}>
              {`apiVersion: apps/v1\nkind: ${currentResource.kind}\nmetadata:\n  name: ${currentResource.name}\n  namespace: ${currentResource.namespace ?? 'default'}`}
            </Typography>
          ),
        })}
      />
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
