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
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';

interface ConfigMapDetail {
  name: string;
  namespace: string;
  type: 'ConfigMap' | 'Secret';
  data: Record<string, string>;
  labels: Record<string, string>;
  createdAt: string;
}

interface ConfigMapEditorProps {
  configMap: ConfigMapDetail;
  onSave: (data: Record<string, string>) => void;
  onClose: () => void;
}

function ConfigMapEditor({ configMap, onSave, onClose }: ConfigMapEditorProps) {
  const [data, setData] = useState(configMap.data);

  const handleAddKey = () => {
    setData({ ...data, '': '' });
  };

  const handleUpdateKey = (oldKey: string, newKey: string, value: string) => {
    const newData = { ...data };
    delete newData[oldKey];
    newData[newKey] = value;
    setData(newData);
  };

  const handleDeleteKey = (key: string) => {
    const newData = { ...data };
    delete newData[key];
    setData(newData);
  };

  const handleSave = () => {
    onSave(data);
  };

  return (
    <Dialog open maxWidth="md" fullWidth>
      <DialogTitle>Edit {configMap.type}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          {Object.entries(data).map(([key, value]) => (
            <Stack key={key} direction="row" spacing={1} alignItems="flex-start">
              <TextField
                fullWidth
                placeholder="Key"
                defaultValue={key}
                onChange={(e) => handleUpdateKey(key, e.target.value, value)}
                size="small"
              />
              <TextField
                fullWidth
                placeholder="Value"
                defaultValue={value}
                multiline
                rows={value.includes('\n') ? 3 : 1}
                onChange={(e) => handleUpdateKey(key, key, e.target.value)}
                size="small"
              />
              <IconButton onClick={() => handleDeleteKey(key)} color="error">
                <DeleteIcon />
              </IconButton>
            </Stack>
          ))}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleAddKey} startIcon={<AddIcon />}>Add Key</Button>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSave} variant="contained">Save</Button>
      </DialogActions>
    </Dialog>
  );
}

export function ConfigMapPage() {
  const [activeTab, setActiveTab] = useState(0);
  const [editorOpen, setEditorOpen] = useState(false);
  const [configMap] = useState<ConfigMapDetail>({
    name: 'app-config',
    namespace: 'default',
    type: 'ConfigMap',
    data: {
      DATABASE_URL: 'postgres://db:5432/app',
      LOG_LEVEL: 'info',
      CACHE_TTL: '3600',
    },
    labels: { app: 'myapp', tier: 'backend' },
    createdAt: '2024-01-15T10:30:00Z',
  });

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const handleSave = (data: Record<string, string>) => {
    console.log('Saving config:', data);
    setEditorOpen(false);
  };

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {configMap.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: {configMap.namespace}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip label={configMap.type} color={configMap.type === 'Secret' ? 'warning' : 'primary'} />
            <Chip label={`${Object.keys(configMap.data).length} keys`} variant="outlined" size="small" />
          </Stack>
        </Stack>
      </Paper>

      {/* Actions */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" startIcon={<EditIcon />} onClick={() => setEditorOpen(true)}>
              Edit {configMap.type}
            </Button>
            <Button variant="outlined" startIcon={<ContentCopyIcon />}>
              Duplicate
            </Button>
          </Stack>
          <Button variant="outlined" color="error" startIcon={<DeleteIcon />}>
            Delete
          </Button>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label="Data" />
          <Tab label="Labels" />
          <Tab label="Usage" />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Data Tab */}
      {activeTab === 0 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Data ({Object.keys(configMap.data).length})
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Key</TableCell>
                <TableCell>Value</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {Object.entries(configMap.data).map(([key, value]) => (
                <TableRow key={key}>
                  <TableCell>
                    <Typography fontWeight={600}>{key}</Typography>
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Typography fontFamily="monospace" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {configMap.type === 'Secret' ? '••••••••' : value}
                      </Typography>
                      <Tooltip title="Copy">
                        <IconButton size="small" onClick={() => copyToClipboard(value)}>
                          <ContentCopyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Stack direction="row" spacing={0.5}>
                      <IconButton size="small" onClick={() => setEditorOpen(true)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton size="small" color="error">
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>
      )}

      {/* Labels Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Labels
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            {Object.entries(configMap.labels).map(([key, value]) => (
              <Chip key={key} label={`${key}: ${value}`} variant="outlined" />
            ))}
          </Stack>
        </Paper>
      )}

      {/* Usage Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
            Referenced By
          </Typography>
          <Typography color="text.secondary">
            This {configMap.type.toLowerCase()} is not currently referenced by any resources.
          </Typography>
        </Paper>
      )}

      {/* YAML Tab */}
      {activeTab === 3 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
            {configMap.type} YAML
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
kind: ${configMap.type}
metadata:
  name: ${configMap.name}
  namespace: ${configMap.namespace}
  labels:
${Object.entries(configMap.labels).map(([k, v]) => `    ${k}: ${v}`).join('\n')}
  createdAt: ${configMap.createdAt}
${configMap.type === 'Secret' ? 'type: Opaque' : ''}
data:
${Object.entries(configMap.data).map(([k, v]) => `  ${k}: ${v}`).join('\n')}`}</pre>
          </Box>
        </Paper>
      )}

      {/* Editor Dialog */}
      {editorOpen && (
        <ConfigMapEditor
          configMap={configMap}
          onSave={handleSave}
          onClose={() => setEditorOpen(false)}
        />
      )}
    </Stack>
  );
}
