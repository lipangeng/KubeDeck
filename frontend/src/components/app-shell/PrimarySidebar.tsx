import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Paper from '@mui/material/Paper';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import type { MenuItem } from '../../sdk/types';

interface PrimarySidebarProps {
  clusters: string[];
  selectedCluster: string;
  namespaceScopeLabel: string;
  clusterStatus: string;
  isWorkloadsPage: boolean;
  secondaryEntries: MenuItem[];
  onClusterChange: (nextCluster: string) => void;
  onEnterWorkloads: () => void;
}

function AdditionalEntryList({ items }: { items: MenuItem[] }) {
  return (
    <Paper variant="outlined" sx={{ p: 2 }}>
      <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
        Additional Entries
      </Typography>
      <Stack spacing={1}>
        {items.length === 0 ? (
          <Typography color="text.disabled">No additional entries yet</Typography>
        ) : (
          items.map((menu) => (
            <Paper key={menu.id} variant="outlined" sx={{ p: 1.25 }}>
              <Typography sx={{ fontWeight: 600 }}>{menu.title}</Typography>
              <Typography variant="body2" color="text.secondary">
                {menu.targetType} · available later
              </Typography>
            </Paper>
          ))
        )}
      </Stack>
    </Paper>
  );
}

export function PrimarySidebar({
  clusters,
  selectedCluster,
  namespaceScopeLabel,
  clusterStatus,
  isWorkloadsPage,
  secondaryEntries,
  onClusterChange,
  onEnterWorkloads,
}: PrimarySidebarProps) {
  return (
    <Paper
      component="nav"
      aria-label="Primary Sidebar"
      elevation={2}
      sx={{
        p: 2,
        minHeight: { md: 'calc(100vh - 120px)' },
        border: 1,
        borderColor: 'divider',
      }}
    >
      <Stack spacing={2}>
        <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
          Working Context
        </Typography>

        <FormControl size="small" fullWidth>
          <InputLabel htmlFor="cluster-select">Cluster</InputLabel>
          <Select
            native
            value={selectedCluster}
            onChange={(event) => onClusterChange(event.target.value)}
            label="Cluster"
            inputProps={{ id: 'cluster-select' }}
          >
            {clusters.map((clusterId) => (
              <option key={clusterId} value={clusterId}>
                {clusterId}
              </option>
            ))}
          </Select>
        </FormControl>

        <Paper variant="outlined" sx={{ p: 1.5 }}>
          <Typography variant="caption" color="text.secondary">
            Active namespace scope
          </Typography>
          <Typography sx={{ fontWeight: 700 }}>{namespaceScopeLabel}</Typography>
          <Typography variant="body2" color="text.secondary">
            Cluster status: {clusterStatus}
          </Typography>
        </Paper>

        <Divider />

        <Button
          variant={isWorkloadsPage ? 'contained' : 'outlined'}
          onClick={onEnterWorkloads}
        >
          Enter Workloads
        </Button>

        <AdditionalEntryList items={secondaryEntries} />
      </Stack>
    </Paper>
  );
}
