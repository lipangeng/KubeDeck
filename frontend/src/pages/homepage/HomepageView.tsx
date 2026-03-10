import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import type { MenuItem } from '../../sdk/types';
import type { NamespaceScope } from '../../state/work-context/types';

interface HomepageViewProps {
  activeClusterId: string;
  namespaceScopeLabel: string;
  primaryEntryTitle: string;
  onEnterWorkloads: () => void;
  blockingSummary: string | null;
  healthStatus: 'checking' | 'ok' | 'error';
  readyStatus: 'checking' | 'ok' | 'error';
  apiTargetHint: string;
  lastCheckedAt: string | null;
  additionalEntries: MenuItem[];
  statusColor: (status: 'checking' | 'ok' | 'error') => 'success' | 'warning' | 'error';
}

function PrimaryEntryCard({
  title,
  description,
  actionLabel,
  onClick,
}: {
  title: string;
  description: string;
  actionLabel: string;
  onClick: () => void;
}) {
  return (
    <Paper variant="outlined" sx={{ p: 2 }}>
      <Stack spacing={1}>
        <Typography variant="overline" color="primary.main">
          Primary Workflow
        </Typography>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          {title}
        </Typography>
        <Typography color="text.secondary">{description}</Typography>
        <Button variant="contained" onClick={onClick} sx={{ alignSelf: 'flex-start' }}>
          {actionLabel}
        </Button>
      </Stack>
    </Paper>
  );
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

export function HomepageView({
  activeClusterId,
  namespaceScopeLabel,
  primaryEntryTitle,
  onEnterWorkloads,
  blockingSummary,
  healthStatus,
  readyStatus,
  apiTargetHint,
  lastCheckedAt,
  additionalEntries,
  statusColor,
}: HomepageViewProps) {
  return (
    <Stack spacing={2}>
      <Paper elevation={3} sx={{ p: 2.2, border: 1, borderColor: 'divider' }}>
        <Typography variant="overline" color="primary.main" sx={{ letterSpacing: 1.1 }}>
          Current Context
        </Typography>
        <Typography variant="h5" sx={{ fontWeight: 700, mb: 0.8 }}>
          Cluster {activeClusterId}
        </Typography>
        <Typography color="text.secondary">
          Namespace scope: {namespaceScopeLabel}
        </Typography>
      </Paper>

      <PrimaryEntryCard
        title={primaryEntryTitle}
        description="Browse workloads and start create/apply actions in the current cluster context."
        actionLabel="Enter Workloads"
        onClick={onEnterWorkloads}
      />

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
          Default Task
        </Typography>
        <Typography color="text.secondary" sx={{ mb: 1.5 }}>
          Start in Workloads to inspect live resources and continue with create or apply in the same cluster and namespace context.
        </Typography>
        <Button variant="text" onClick={onEnterWorkloads}>
          Continue to Workloads
        </Button>
      </Paper>

      {blockingSummary ? <Alert severity="warning">Blocking summary: {blockingSummary}</Alert> : null}

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 1 }}>
          Runtime diagnostics
        </Typography>
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mb: 1 }}>
          <Chip
            size="small"
            color={statusColor(healthStatus)}
            label={`healthz: ${healthStatus}`}
          />
          <Chip
            size="small"
            color={statusColor(readyStatus)}
            label={`readyz: ${readyStatus}`}
          />
        </Stack>
        <Typography variant="body2" color="text.secondary">
          API target ({apiTargetHint}) · Last checked: {lastCheckedAt ?? 'never'}
        </Typography>
      </Paper>

      <AdditionalEntryList items={additionalEntries} />
    </Stack>
  );
}
