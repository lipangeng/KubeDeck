import { useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import { copy } from '../../../i18n/copy';

interface IngressRule {
  host: string;
  path: string;
  pathType: string;
  service: string;
  port: number;
}

interface IngressTLS {
  hosts: string[];
  secret: string;
  status: string;
}

export function IngressPage() {
  const [activeTab, setActiveTab] = useState(0);
  const [rulesDialogOpen, setRulesDialogOpen] = useState(false);
  const [tlsDialogOpen, setTlsDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<IngressRule | null>(null);

  const mockRules: IngressRule[] = [
    { host: 'api.example.com', path: '/api', pathType: 'Prefix', service: 'api-service', port: 8080 },
    { host: '*.example.com', path: '/', pathType: 'Prefix', service: 'web-service', port: 80 },
  ];

  const mockTLS: IngressTLS[] = [
    { hosts: ['api.example.com'], secret: 'api-tls', status: 'Valid' },
  ];

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6">
              api-ingress
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: default
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip label="Class: nginx" variant="outlined" />
            <Chip label="Address: 192.168.1.100" variant="outlined" />
          </Stack>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label="Rules" />
          <Tab label="TLS" />
          <Tab label="Events" />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Rules Tab */}
      {activeTab === 0 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={2}>
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                Routing Rules
              </Typography>
              <Button
                variant="contained"
                onClick={() => {
                  setEditingRule(null);
                  setRulesDialogOpen(true);
                }}
              >
                + Add Rule
              </Button>
            </Stack>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Host</TableCell>
                  <TableCell>Path</TableCell>
                  <TableCell>Path Type</TableCell>
                  <TableCell>Service</TableCell>
                  <TableCell>Port</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {mockRules.map((rule, index) => (
                  <TableRow key={index}>
                    <TableCell>{rule.host}</TableCell>
                    <TableCell>{rule.path}</TableCell>
                    <TableCell>{rule.pathType}</TableCell>
                    <TableCell>{rule.service}</TableCell>
                    <TableCell>{rule.port}</TableCell>
                    <TableCell>
                      <Stack direction="row" spacing={0.5}>
                        <Button
                          size="small"
                          onClick={() => {
                            setEditingRule(rule);
                            setRulesDialogOpen(true);
                          }}
                        >
                          Edit
                        </Button>
                        <Button size="small" color="error">Delete</Button>
                      </Stack>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Stack>
        </Paper>
      )}

      {/* TLS Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={2}>
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                TLS Configuration
              </Typography>
              <Button
                variant="contained"
                onClick={() => setTlsDialogOpen(true)}
              >
                + Add TLS
              </Button>
            </Stack>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Hosts</TableCell>
                  <TableCell>Secret</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {mockTLS.map((tls, index) => (
                  <TableRow key={index}>
                    <TableCell>{tls.hosts.join(', ')}</TableCell>
                    <TableCell>{tls.secret}</TableCell>
                    <TableCell>
                      <Chip label={tls.status} size="small" color="success" />
                    </TableCell>
                    <TableCell>
                      <Button size="small" color="error">Delete</Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Stack>
        </Paper>
      )}

      {/* Events Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography>Events will be implemented</Typography>
        </Paper>
      )}

      {/* YAML Tab */}
      {activeTab === 3 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
            Ingress YAML
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
            }}
          >
            <pre>{`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 8080
  tls:
  - hosts:
    - api.example.com
    secretName: api-tls`}</pre>
          </Box>
        </Paper>
      )}

      {/* Add/Edit Rule Dialog */}
      <Dialog open={rulesDialogOpen} onClose={() => setRulesDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingRule ? 'Edit' : 'Add'} Ingress Rule</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Host"
              defaultValue={editingRule?.host || ''}
              placeholder="api.example.com"
            />
            <TextField
              fullWidth
              label="Path"
              defaultValue={editingRule?.path || '/'}
              placeholder="/api"
            />
            <FormControl fullWidth>
              <InputLabel>Path Type</InputLabel>
              <Select defaultValue={editingRule?.pathType || 'Prefix'} label="Path Type">
                <MenuItem value="Prefix">Prefix</MenuItem>
                <MenuItem value="Exact">Exact</MenuItem>
                <MenuItem value="ImplementationSpecific">ImplementationSpecific</MenuItem>
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="Service"
              defaultValue={editingRule?.service || ''}
              placeholder="api-service"
            />
            <TextField
              fullWidth
              label="Port"
              type="number"
              defaultValue={editingRule?.port || 80}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRulesDialogOpen(false)}>Cancel</Button>
          <Button variant="contained">Save</Button>
        </DialogActions>
      </Dialog>

      {/* Add TLS Dialog */}
      <Dialog open={tlsDialogOpen} onClose={() => setTlsDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add TLS Configuration</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Hosts (comma separated)"
              placeholder="api.example.com, www.example.com"
            />
            <TextField
              fullWidth
              label="TLS Secret"
              placeholder="api-tls"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTlsDialogOpen(false)}>Cancel</Button>
          <Button variant="contained">Save</Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
