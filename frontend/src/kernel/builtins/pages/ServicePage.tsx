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
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Link from '@mui/material/Link';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';

interface ServiceDetail {
  name: string;
  namespace: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';
  clusterIP: string;
  externalIP: string;
  loadBalancerIP?: string;
  ports: ServicePort[];
  selector: Record<string, string>;
  endpoints: Endpoint[];
}

interface ServicePort {
  name: string;
  port: number;
  targetPort: number;
  protocol: string;
  nodePort?: number;
}

interface Endpoint {
  ip: string;
  port: number;
  ready: boolean;
  nodeName?: string;
}

export function ServicePage() {
  const [activeTab, setActiveTab] = useState(0);
  const [service] = useState<ServiceDetail>({
    name: 'api-service',
    namespace: 'default',
    type: 'LoadBalancer',
    clusterIP: '10.96.100.50',
    externalIP: '203.0.113.50',
    loadBalancerIP: '203.0.113.50',
    ports: [
      { name: 'http', port: 80, targetPort: 8080, protocol: 'TCP', nodePort: 30080 },
      { name: 'https', port: 443, targetPort: 8443, protocol: 'TCP', nodePort: 30443 },
    ],
    selector: { app: 'api-server', tier: 'backend' },
    endpoints: [
      { ip: '10.244.1.10', port: 8080, ready: true, nodeName: 'node-1' },
      { ip: '10.244.2.15', port: 8080, ready: true, nodeName: 'node-2' },
      { ip: '10.244.3.20', port: 8080, ready: false, nodeName: 'node-3' },
    ],
  });

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {service.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: {service.namespace}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip label={service.type} color="primary" />
            <Chip 
              label={service.externalIP ? 'External' : 'Internal'} 
              color={service.externalIP ? 'success' : 'default'}
            />
          </Stack>
        </Stack>
      </Paper>

      {/* Quick Info */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Stack direction="row" spacing={3} alignItems="center">
            <Box>
              <Typography variant="body2" color="text.secondary">Cluster IP</Typography>
              <Stack direction="row" spacing={0.5} alignItems="center">
                <Typography variant="h6" fontFamily="monospace">{service.clusterIP}</Typography>
                <Tooltip title="Copy">
                  <IconButton size="small" onClick={() => copyToClipboard(service.clusterIP)}>
                    <ContentCopyIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
            </Box>
            {service.externalIP && (
              <Box>
                <Typography variant="body2" color="text.secondary">External IP</Typography>
                <Stack direction="row" spacing={0.5} alignItems="center">
                  <Typography variant="h6" fontFamily="monospace">{service.externalIP}</Typography>
                  <Tooltip title="Copy">
                    <IconButton size="small" onClick={() => copyToClipboard(service.externalIP)}>
                      <ContentCopyIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Stack>
              </Box>
            )}
            {service.loadBalancerIP && (
              <Box>
                <Typography variant="body2" color="text.secondary">LoadBalancer IP</Typography>
                <Typography variant="h6" fontFamily="monospace">{service.loadBalancerIP}</Typography>
              </Box>
            )}
          </Stack>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label="Ports" />
          <Tab label="Endpoints" />
          <Tab label="Selector" />
          <Tab label="Events" />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Ports Tab */}
      {activeTab === 0 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Service Ports
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Protocol</TableCell>
                <TableCell>Port</TableCell>
                <TableCell>Target Port</TableCell>
                {service.type === 'NodePort' && <TableCell>Node Port</TableCell>}
              </TableRow>
            </TableHead>
            <TableBody>
              {service.ports.map((port) => (
                <TableRow key={port.name}>
                  <TableCell>{port.name}</TableCell>
                  <TableCell>{port.protocol}</TableCell>
                  <TableCell>{port.port}</TableCell>
                  <TableCell>{port.targetPort}</TableCell>
                  {service.type === 'NodePort' && (
                    <TableCell>{port.nodePort}</TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>
      )}

      {/* Endpoints Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Endpoints ({service.endpoints.length})
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>IP</TableCell>
                <TableCell>Port</TableCell>
                <TableCell>Ready</TableCell>
                <TableCell>Node</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {service.endpoints.map((endpoint) => (
                <TableRow key={endpoint.ip}>
                  <TableCell>
                    <Typography fontFamily="monospace">{endpoint.ip}</Typography>
                  </TableCell>
                  <TableCell>{endpoint.port}</TableCell>
                  <TableCell>
                    <Chip 
                      label={endpoint.ready ? 'Ready' : 'Not Ready'} 
                      size="small"
                      color={endpoint.ready ? 'success' : 'error'}
                    />
                  </TableCell>
                  <TableCell>{endpoint.nodeName || '-'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>
      )}

      {/* Selector Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Pod Selector
          </Typography>
          <Box sx={{ p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
            {Object.entries(service.selector).map(([key, value]) => (
              <Stack key={key} direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                <Typography component="span" fontWeight={600}>{key}:</Typography>
                <Chip label={value} size="small" />
              </Stack>
            ))}
          </Box>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            This service routes traffic to pods matching the above labels.
          </Typography>
        </Paper>
      )}

      {/* Events Tab */}
      {activeTab === 3 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Events
          </Typography>
          <Typography color="text.secondary">
            Events will be implemented in a future update.
          </Typography>
        </Paper>
      )}

      {/* YAML Tab */}
      {activeTab === 4 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
            Service YAML
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
            <pre>{`apiVersion: v1
kind: Service
metadata:
  name: ${service.name}
  namespace: ${service.namespace}
spec:
  type: ${service.type}
  clusterIP: ${service.clusterIP}${service.externalIP ? `
  externalIPs:
  - ${service.externalIP}` : ''}
  ports:
${service.ports.map(p => `  - name: ${p.name}
    port: ${p.port}
    targetPort: ${p.targetPort}
    protocol: ${p.protocol}${service.type === 'NodePort' ? `
    nodePort: ${p.nodePort}` : ''}`).join('\n')}
  selector:
${Object.entries(service.selector).map(([k, v]) => `    ${k}: ${v}`).join('\n')}`}</pre>
          </Box>
        </Paper>
      )}
    </Stack>
  );
}
