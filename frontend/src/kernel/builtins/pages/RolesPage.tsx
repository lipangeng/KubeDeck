import { useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';

const resourceTypes = [
  'Pods', 'Deployments', 'Services', 'ConfigMaps', 'Secrets',
  'Ingresses', 'PersistentVolumes', 'PersistentVolumeClaims',
  'Nodes', 'Namespaces', 'Roles', 'RoleBindings',
];

const operations = [
  { id: 'get', label: 'get' },
  { id: 'list', label: 'list' },
  { id: 'watch', label: 'watch' },
  { id: 'create', label: 'create' },
  { id: 'update', label: 'update' },
  { id: 'patch', label: 'patch' },
  { id: 'delete', label: 'delete' },
];

interface Role {
  id: string;
  name: string;
  description: string;
  isBuiltin: boolean;
  resources: string[];
  verbs: string[];
}

export function RolesPage() {
  const [rolesDialogOpen, setRolesDialogOpen] = useState(false);
  const [yamlMode, setYamlMode] = useState(false);
  const [yamlContent, setYamlContent] = useState('');
  const [selectedResources, setSelectedResources] = useState<string[]>([]);
  const [selectedVerbs, setSelectedVerbs] = useState<string[]>([]);
  const [roleName, setRoleName] = useState('');
  const [roleDescription, setRoleDescription] = useState('');

  const mockRoles: Role[] = [
    {
      id: '1',
      name: 'admin',
      description: 'Cluster administrator with full access',
      isBuiltin: true,
      resources: ['*'],
      verbs: ['*'],
    },
    {
      id: '2',
      name: 'developer',
      description: 'Developer with read/write access to workloads',
      isBuiltin: true,
      resources: ['Pods', 'Deployments', 'Services', 'ConfigMaps'],
      verbs: ['get', 'list', 'watch', 'create', 'update', 'delete'],
    },
    {
      id: '3',
      name: 'viewer',
      description: 'Read-only access to all resources',
      isBuiltin: true,
      resources: ['*'],
      verbs: ['get', 'list', 'watch'],
    },
  ];

  const handleResourceToggle = (resource: string) => {
    setSelectedResources(prev =>
      prev.includes(resource)
        ? prev.filter(r => r !== resource)
        : [...prev, resource]
    );
  };

  const handleVerbToggle = (verb: string) => {
    setSelectedVerbs(prev =>
      prev.includes(verb)
        ? prev.filter(v => v !== verb)
        : [...prev, verb]
    );
  };

  const handleSelectAllResources = () => {
    setSelectedResources(resourceTypes);
  };

  const handleClearResources = () => {
    setSelectedResources([]);
  };

  const handleSelectAllVerbs = () => {
    setSelectedVerbs(operations.map(op => op.id));
  };

  const handleClearVerbs = () => {
    setSelectedVerbs([]);
  };

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h5" sx={{ fontWeight: 700 }}>
              Roles
            </Typography>
            <Typography color="text.secondary">
              Manage RBAC roles and permissions across clusters
            </Typography>
          </Stack>
          <Button
            variant="contained"
            onClick={() => {
              setRoleName('');
              setRoleDescription('');
              setSelectedResources([]);
              setSelectedVerbs([]);
              setYamlMode(false);
              setRolesDialogOpen(true);
            }}
          >
            Create Role
          </Button>
        </Stack>
      </Paper>

      {/* Roles Table */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Description</TableCell>
              <TableCell>Resources</TableCell>
              <TableCell>Verbs</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {mockRoles.map((role) => (
              <TableRow key={role.id}>
                <TableCell>
                  <Typography sx={{ fontWeight: 600 }}>{role.name}</Typography>
                </TableCell>
                <TableCell>{role.description}</TableCell>
                <TableCell>
                  <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
                    {role.resources.slice(0, 3).map(r => (
                      <Chip key={r} label={r} size="small" variant="outlined" />
                    ))}
                    {role.resources.length > 3 && (
                      <Chip label={`+${role.resources.length - 3}`} size="small" />
                    )}
                  </Stack>
                </TableCell>
                <TableCell>
                  <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
                    {role.verbs.slice(0, 3).map(v => (
                      <Chip key={v} label={v} size="small" color="primary" variant="outlined" />
                    ))}
                    {role.verbs.length > 3 && (
                      <Chip label={`+${role.verbs.length - 3}`} size="small" />
                    )}
                  </Stack>
                </TableCell>
                <TableCell>
                  {role.isBuiltin ? (
                    <Chip label="Builtin" size="small" color="info" />
                  ) : (
                    <Chip label="Custom" size="small" color="default" />
                  )}
                </TableCell>
                <TableCell>
                  <Stack direction="row" spacing={0.5}>
                    <Button size="small">Edit</Button>
                    <Button size="small" color="error" disabled={role.isBuiltin}>Delete</Button>
                  </Stack>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Paper>

      {/* Create/Edit Role Dialog */}
      <Dialog open={rolesDialogOpen} onClose={() => setRolesDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>Create Role</DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Role Name"
              value={roleName}
              onChange={(e) => setRoleName(e.target.value)}
              placeholder="custom-role"
            />
            <TextField
              fullWidth
              label="Description"
              value={roleDescription}
              onChange={(e) => setRoleDescription(e.target.value)}
              placeholder="Role description"
              multiline
              rows={2}
            />

            <Box sx={{ borderBottom: 1, borderColor: 'divider', pb: 1 }}>
              <Stack direction="row" spacing={1} alignItems="center">
                <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                  Permission Rules
                </Typography>
                <Box sx={{ flexGrow: 1 }} />
                <Button size="small" onClick={() => setYamlMode(!yamlMode)}>
                  {yamlMode ? 'Use Visual Editor' : 'Switch to YAML'}
                </Button>
              </Stack>
            </Box>

            {yamlMode ? (
              <TextField
                fullWidth
                multiline
                rows={15}
                value={yamlContent}
                onChange={(e) => setYamlContent(e.target.value)}
                placeholder="apiVersion: rbac.authorization.k8s.io/v1&#10;kind: ClusterRole&#10;metadata:&#10;  name: custom-role..."
                sx={{
                  '& .MuiInputBase-input': {
                    fontFamily: 'monospace',
                    fontSize: '0.875rem',
                  },
                }}
              />
            ) : (
              <Stack spacing={3}>
                <Stack direction="row" spacing={3}>
                  {/* Resources */}
                  <Box sx={{ flex: 1 }}>
                    <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1 }}>
                      Resource Types
                    </Typography>
                    <Paper variant="outlined" sx={{ p: 2, maxHeight: 250, overflow: 'auto' }}>
                      <Stack spacing={0.5}>
                        {resourceTypes.map((resource) => (
                          <FormControlLabel
                            key={resource}
                            control={
                              <Checkbox
                                checked={selectedResources.includes(resource)}
                                onChange={() => handleResourceToggle(resource)}
                                size="small"
                              />
                            }
                            label={resource}
                          />
                        ))}
                      </Stack>
                    </Paper>
                    <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                      <Button size="small" onClick={handleSelectAllResources}>Select All</Button>
                      <Button size="small" onClick={handleClearResources}>Clear</Button>
                    </Stack>
                  </Box>

                  {/* Operations */}
                  <Box sx={{ flex: 1 }}>
                    <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1 }}>
                      Operations
                    </Typography>
                    <Paper variant="outlined" sx={{ p: 2 }}>
                      <Stack spacing={0.5}>
                        {operations.map((op) => (
                          <FormControlLabel
                            key={op.id}
                            control={
                              <Checkbox
                                checked={selectedVerbs.includes(op.id)}
                                onChange={() => handleVerbToggle(op.id)}
                                size="small"
                              />
                            }
                            label={op.label}
                          />
                        ))}
                      </Stack>
                    </Paper>
                    <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
                      <Button size="small" onClick={handleSelectAllVerbs}>Select All</Button>
                      <Button size="small" onClick={handleClearVerbs}>Clear</Button>
                    </Stack>
                  </Box>
                </Stack>
              </Stack>
            )}

            {/* YAML Preview */}
            {!yamlMode && (
              <Box>
                <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1 }}>
                  YAML Preview
                </Typography>
                <Box
                  sx={{
                    bgcolor: '#1e1e1e',
                    color: '#d4d4d4',
                    p: 2,
                    fontFamily: 'monospace',
                    fontSize: '0.875rem',
                    borderRadius: 1,
                  }}
                >
                  <pre>{`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${roleName || 'custom-role'}
rules:
- apiGroups: [""]
  resources: [${selectedResources.map(r => `"${r.toLowerCase()}"`).join(', ')}]
  verbs: [${selectedVerbs.map(v => `"${v}"`).join(', ')}]`}</pre>
                </Box>
              </Box>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRolesDialogOpen(false)}>Cancel</Button>
          <Button variant="contained">Create Role</Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
