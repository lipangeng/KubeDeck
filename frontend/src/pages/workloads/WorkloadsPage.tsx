import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { ListPageShell } from '../../components/page-shell/ResourcePageShell';
import type { WorkloadItem } from '../../sdk/types';
import type { ActionType, NamespaceScope } from '../../state/work-context/types';

interface WorkloadsPageProps {
  activeClusterId: string;
  namespaceScope: NamespaceScope;
  namespaceScopeLabel: string;
  searchText: string;
  onSearchTextChange: (next: string) => void;
  onReturnHomepage: () => void;
  onRefresh: () => void;
  onNamespaceScopeChange: (value: string) => void;
  onOpenAction: (actionType: ActionType) => void;
  loading: boolean;
  metadataError: string | null;
  workloads: WorkloadItem[];
  resultBanner: string | null;
  resultBannerSeverity: 'success' | 'error';
  onDismissResult: () => void;
}

function resolveSingleNamespace(scope: NamespaceScope): string {
  return scope.mode === 'single' ? (scope.values[0] ?? 'default') : '';
}

export function WorkloadsPage({
  activeClusterId,
  namespaceScope,
  namespaceScopeLabel,
  searchText,
  onSearchTextChange,
  onReturnHomepage,
  onRefresh,
  onNamespaceScopeChange,
  onOpenAction,
  loading,
  metadataError,
  workloads,
  resultBanner,
  resultBannerSeverity,
  onDismissResult,
}: WorkloadsPageProps) {
  return (
    <ListPageShell
      title="Workloads"
      toolbar={
        <Stack direction="row" spacing={1}>
          <Button variant="text" onClick={onReturnHomepage}>
            Homepage
          </Button>
          <Button variant="outlined" onClick={onRefresh}>
            Refresh
          </Button>
          <Button variant="contained" onClick={() => onOpenAction('apply')}>
            Apply
          </Button>
          <Button variant="outlined" onClick={() => onOpenAction('create')}>
            Create
          </Button>
        </Stack>
      }
    >
      <Stack spacing={2}>
        <Paper variant="outlined" sx={{ p: 1.5 }}>
          <Stack
            direction={{ xs: 'column', md: 'row' }}
            spacing={1.5}
            alignItems={{ md: 'center' }}
            justifyContent="space-between"
          >
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
              <Chip color="primary" label={`Cluster: ${activeClusterId}`} />
              <Chip variant="outlined" label={`Namespace scope: ${namespaceScopeLabel}`} />
            </Stack>
            <FormControl size="small" sx={{ minWidth: 220 }}>
              <InputLabel htmlFor="namespace-scope-select">Namespace Scope</InputLabel>
              <Select
                native
                value={namespaceScope.mode === 'all' ? 'all' : resolveSingleNamespace(namespaceScope)}
                onChange={(event) => onNamespaceScopeChange(event.target.value)}
                label="Namespace Scope"
                inputProps={{ id: 'namespace-scope-select' }}
              >
                <option value="default">default</option>
                <option value="all">All namespaces</option>
              </Select>
            </FormControl>
          </Stack>
        </Paper>

        <Paper variant="outlined" sx={{ p: 1.5 }}>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
            <TextField
              label="Search workloads"
              size="small"
              fullWidth
              value={searchText}
              onChange={(event) => onSearchTextChange(event.target.value)}
            />
            <Chip
              variant="outlined"
              label={`Visible workloads: ${workloads.length}`}
              data-testid="workload-row-count"
            />
          </Stack>
        </Paper>

        {loading ? <Alert severity="info">Loading workloads...</Alert> : null}
        {metadataError ? <Alert severity="error">Failed to load workloads: {metadataError}</Alert> : null}

        {resultBanner ? (
          <Alert
            severity={resultBannerSeverity}
            action={
              <Button color="inherit" size="small" onClick={onDismissResult}>
                Dismiss
              </Button>
            }
          >
            {resultBanner}
          </Alert>
        ) : null}

        <Paper variant="outlined">
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Kind</TableCell>
                <TableCell>Namespace</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Health</TableCell>
                <TableCell>Updated</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {workloads.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Typography color="text.secondary">
                      No workloads available in the current cluster and namespace context.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                workloads.map((workload) => (
                  <TableRow key={workload.id}>
                    <TableCell sx={{ fontWeight: 700 }}>{workload.name}</TableCell>
                    <TableCell>{workload.kind}</TableCell>
                    <TableCell>{workload.namespace}</TableCell>
                    <TableCell>{workload.status}</TableCell>
                    <TableCell>{workload.health}</TableCell>
                    <TableCell>{workload.updatedAt}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </Paper>
      </Stack>
    </ListPageShell>
  );
}
