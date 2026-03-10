import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Drawer from '@mui/material/Drawer';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import type { ActionContext, NamespaceScope } from '../../state/work-context/types';

interface ActionDrawerProps {
  open: boolean;
  actionLabel: string;
  activeClusterId: string;
  namespaceScopeLabel: string;
  namespaceScope: NamespaceScope;
  actionContext: ActionContext;
  actionNamespace: string;
  actionManifest: string;
  actionFormError: string | null;
  onClose: () => void;
  onSubmit: () => void;
  onBackToEdit: () => void;
  onReturnToWorkloads: () => void;
  onNamespaceChange: (value: string) => void;
  onManifestChange: (value: string) => void;
  resultSeverity: (outcome: string) => 'success' | 'warning' | 'error';
}

export function ActionDrawer({
  open,
  actionLabel,
  activeClusterId,
  namespaceScopeLabel,
  namespaceScope,
  actionContext,
  actionNamespace,
  actionManifest,
  actionFormError,
  onClose,
  onSubmit,
  onBackToEdit,
  onReturnToWorkloads,
  onNamespaceChange,
  onManifestChange,
  resultSeverity,
}: ActionDrawerProps) {
  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={onClose}
      PaperProps={{ sx: { width: { xs: '100%', sm: 480 }, p: 2 } }}
    >
      <Stack spacing={2}>
        <Typography variant="overline" color="primary.main">
          {actionLabel} Workflow
        </Typography>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          {actionLabel} in {activeClusterId}
        </Typography>
        <Typography color="text.secondary">
          Browsing scope: {namespaceScopeLabel}
        </Typography>

        {actionContext.resultSummary ? (
          <>
            <Alert severity={resultSeverity(actionContext.resultSummary.outcome)}>
              {actionContext.resultSummary.outcome === 'success'
                ? `${actionLabel} succeeded`
                : actionContext.resultSummary.outcome === 'partial_failure'
                  ? `${actionLabel} partially failed`
                  : `${actionLabel} failed`}
            </Alert>

            <Paper variant="outlined" sx={{ p: 1.5 }}>
              <Stack spacing={1}>
                <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                  Result summary
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Execution target:{' '}
                  {actionContext.executionTarget?.kind === 'namespace'
                    ? actionContext.executionTarget.namespace
                    : 'cluster-scoped'}
                </Typography>
                {actionContext.resultSummary.affectedObjects?.length ? (
                  <Typography variant="body2">
                    Affected: {actionContext.resultSummary.affectedObjects.join(', ')}
                  </Typography>
                ) : null}
                {actionContext.resultSummary.failedObjects?.length ? (
                  <Typography variant="body2" color="error">
                    Failed: {actionContext.resultSummary.failedObjects.join(', ')}
                  </Typography>
                ) : null}
              </Stack>
            </Paper>

            <Stack direction="row" spacing={1} justifyContent="flex-end">
              {actionContext.resultSummary.outcome === 'failure' ? (
                <Button variant="outlined" onClick={onBackToEdit}>
                  Back to Edit
                </Button>
              ) : null}
              <Button variant="contained" onClick={onReturnToWorkloads}>
                Back to Workloads
              </Button>
            </Stack>
          </>
        ) : (
          <>
            {actionContext.needsRevalidation ? (
              <Alert severity="warning">
                Namespace browsing scope changed. Review the execution target before submit.
              </Alert>
            ) : null}

            {actionFormError ? <Alert severity="error">{actionFormError}</Alert> : null}

            <TextField
              label="Execution namespace"
              size="small"
              value={actionNamespace}
              onChange={(event) => onNamespaceChange(event.target.value)}
              disabled={namespaceScope.mode === 'single'}
              helperText={
                namespaceScope.mode === 'single'
                  ? 'Derived from the current single-namespace browsing scope.'
                  : 'Required because all namespaces is not a valid write target.'
              }
            />

            <TextField
              label="Manifest"
              multiline
              minRows={12}
              value={actionManifest}
              onChange={(event) => onManifestChange(event.target.value)}
            />

            <Stack direction="row" spacing={1} justifyContent="flex-end">
              <Button variant="text" onClick={onClose}>
                Cancel
              </Button>
              <Button
                variant="contained"
                onClick={onSubmit}
                disabled={actionContext.status === 'submitting'}
              >
                {actionContext.status === 'submitting' ? 'Submitting...' : `Submit ${actionLabel}`}
              </Button>
            </Stack>
          </>
        )}
      </Stack>
    </Drawer>
  );
}
