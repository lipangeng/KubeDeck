import { useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import RefreshIcon from '@mui/icons-material/Refresh';
import ScaleIcon from '@mui/icons-material/AspectRatio';
import HistoryIcon from '@mui/icons-material/History';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import WarningIcon from '@mui/icons-material/Warning';

interface DeploymentDetail {
  name: string;
  namespace: string;
  replicas: number;
  readyReplicas: number;
  availableReplicas: number;
  updatedReplicas: number;
  strategy: string;
  image: string;
  conditions: DeploymentCondition[];
}

interface DeploymentCondition {
  type: string;
  status: string;
  reason: string;
  message: string;
}

export function DeploymentPage() {
  const [activeTab, setActiveTab] = useState(0);
  const [deployment] = useState<DeploymentDetail>({
    name: 'api-server',
    namespace: 'default',
    replicas: 3,
    readyReplicas: 3,
    availableReplicas: 3,
    updatedReplicas: 3,
    strategy: 'RollingUpdate',
    image: 'myapp:v1.2.3',
    conditions: [
      { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable', message: 'Deployment has minimum availability' },
      { type: 'Progressing', status: 'True', reason: 'NewReplicaSetAvailable', message: 'ReplicaSet "api-server-xxx" has successfully progressed' },
    ],
  });

  const handleScale = (delta: number) => {
    const newReplicas = Math.max(0, deployment.replicas + delta);
    console.log(`Scaling ${deployment.name} to ${newReplicas} replicas`);
    // TODO: Implement scaling API call
  };

  const handleRollback = () => {
    console.log(`Rolling back ${deployment.name}`);
    // TODO: Implement rollback API call
  };

  const handleRestart = () => {
    console.log(`Restarting ${deployment.name}`);
    // TODO: Implement restart API call
  };

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {deployment.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: {deployment.namespace}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip 
              label={`${deployment.readyReplicas}/${deployment.replicas} Ready`} 
              color={deployment.readyReplicas === deployment.replicas ? 'success' : 'warning'}
            />
            <Chip label={deployment.strategy} variant="outlined" size="small" />
          </Stack>
        </Stack>
      </Paper>

      {/* Actions */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={1} justifyContent="space-between" alignItems="center">
          <Stack direction="row" spacing={1}>
            <Tooltip title="Scale Down">
              <IconButton onClick={() => handleScale(-1)} color="warning">
                <ScaleIcon sx={{ transform: 'scaleX(-1)' }} />
              </IconButton>
            </Tooltip>
            <Typography sx={{ pt: 1 }}>Replicas: {deployment.replicas}</Typography>
            <Tooltip title="Scale Up">
              <IconButton onClick={() => handleScale(1)} color="primary">
                <ScaleIcon />
              </IconButton>
            </Tooltip>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" startIcon={<RefreshIcon />} onClick={handleRestart}>
              Restart
            </Button>
            <Button variant="outlined" startIcon={<HistoryIcon />} onClick={handleRollback}>
              Rollback
            </Button>
          </Stack>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label="Overview" />
          <Tab label="Replica Sets" />
          <Tab label="Pods" />
          <Tab label="Events" />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Overview Tab */}
      {activeTab === 0 && (
        <Stack spacing={2}>
          {/* Replica Status */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
              Replica Status
            </Typography>
            <Stack direction="row" spacing={3}>
              <Box>
                <Typography variant="body2" color="text.secondary">Desired</Typography>
                <Typography variant="h4">{deployment.replicas}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Ready</Typography>
                <Typography variant="h4" color="success.main">{deployment.readyReplicas}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Available</Typography>
                <Typography variant="h4" color="success.main">{deployment.availableReplicas}</Typography>
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Updated</Typography>
                <Typography variant="h4">{deployment.updatedReplicas}</Typography>
              </Box>
            </Stack>
          </Paper>

          {/* Container Image */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
              Container Image
            </Typography>
            <Typography variant="body2" fontFamily="monospace">
              {deployment.image}
            </Typography>
          </Paper>

          {/* Conditions */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
              Conditions
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Type</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Reason</TableCell>
                  <TableCell>Message</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {deployment.conditions.map((condition) => (
                  <TableRow key={condition.type}>
                    <TableCell>{condition.type}</TableCell>
                    <TableCell>
                      {condition.status === 'True' ? (
                        <Chip icon={<CheckCircleIcon />} label="True" color="success" size="small" />
                      ) : condition.status === 'False' ? (
                        <Chip icon={<ErrorIcon />} label="False" color="error" size="small" />
                      ) : (
                        <Chip icon={<WarningIcon />} label="Unknown" color="warning" size="small" />
                      )}
                    </TableCell>
                    <TableCell>{condition.reason}</TableCell>
                    <TableCell>{condition.message}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Paper>
        </Stack>
      )}

      {/* Replica Sets Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Replica Sets
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Desired</TableCell>
                <TableCell>Current</TableCell>
                <TableCell>Ready</TableCell>
                <TableCell>Age</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <TableRow>
                <TableCell>api-server-7d8f9b6c5</TableCell>
                <TableCell>3</TableCell>
                <TableCell>3</TableCell>
                <TableCell>3</TableCell>
                <TableCell>2d</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Paper>
      )}

      {/* Pods Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Pods (3)
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Ready</TableCell>
                <TableCell>Restarts</TableCell>
                <TableCell>Age</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {[1, 2, 3].map((i) => (
                <TableRow key={i}>
                  <TableCell>api-server-7d8f9b6c5-abc{i}</TableCell>
                  <TableCell><Chip label="Running" size="small" color="success" /></TableCell>
                  <TableCell>1/1</TableCell>
                  <TableCell>0</TableCell>
                  <TableCell>2d</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>
      )}

      {/* Events Tab */}
      {activeTab === 3 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Events
          </Typography>
          <Typography color="text.secondary">Events will be implemented</Typography>
        </Paper>
      )}

      {/* YAML Tab */}
      {activeTab === 4 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
            Deployment YAML
          </Typography>
          <Box
            sx={{
              bgcolor: '#1e1e1e',
              color: '#d4d4d4',
              p: 2,
              fontFamily: 'monospace',
              fontSize: '0.875rem',
              overflow: 'auto',
              maxHeight: 400,
              borderRadius: 1,
            }}
          >
            <pre>{`apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${deployment.name}
  namespace: ${deployment.namespace}
spec:
  replicas: ${deployment.replicas}
  selector:
    matchLabels:
      app: ${deployment.name}
  template:
    metadata:
      labels:
        app: ${deployment.name}
    spec:
      containers:
      - name: ${deployment.name}
        image: ${deployment.image}`}</pre>
          </Box>
        </Paper>
      )}
    </Stack>
  );
}
