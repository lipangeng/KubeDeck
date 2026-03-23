import { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  Typography,
  Alert,
  Stack,
  MenuItem,
} from '@mui/material';
import { YamlEditor } from './YamlEditor';

interface ResourceCreateFormProps {
  open: boolean;
  onClose: () => void;
  resourceType?: string;
  namespace?: string;
}

const resourceTypes = [
  { value: 'Deployment', label: 'Deployment' },
  { value: 'Service', label: 'Service' },
  { value: 'ConfigMap', label: 'ConfigMap' },
  { value: 'Secret', label: 'Secret' },
  { value: 'Ingress', label: 'Ingress' },
  { value: 'Pod', label: 'Pod' },
];

export function ResourceCreateForm({ open, onClose, resourceType, namespace }: ResourceCreateFormProps) {
  const [mode, setMode] = useState<'form' | 'yaml'>('yaml');
  const [selectedType, setSelectedType] = useState(resourceType || 'Deployment');
  const [name, setName] = useState('');
  const [namespaceName, setNamespaceName] = useState(namespace || 'default');
  const [yamlContent, setYamlContent] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/resources/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: selectedType,
          namespace: namespaceName,
          yaml: yamlContent,
        }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.message || '创建失败');
      }

      handleClose();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setError(null);
    setYamlContent('');
    onClose();
  };

  const generateTemplate = () => {
    const templates: Record<string, string> = {
      Deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${name || 'my-app'}
  namespace: ${namespaceName}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${name || 'my-app'}
  template:
    metadata:
      labels:
        app: ${name || 'my-app'}
    spec:
      containers:
      - name: ${name || 'my-app'}
        image: nginx:latest
        ports:
        - containerPort: 80`,

      Service: `apiVersion: v1
kind: Service
metadata:
  name: ${name || 'my-service'}
  namespace: ${namespaceName}
spec:
  selector:
    app: ${name || 'my-app'}
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP`,

      ConfigMap: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${name || 'my-config'}
  namespace: ${namespaceName}
data:
  key1: value1
  key2: value2`,

      Secret: `apiVersion: v1
kind: Secret
metadata:
  name: ${name || 'my-secret'}
  namespace: ${namespaceName}
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm`,

      Ingress: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${name || 'my-ingress'}
  namespace: ${namespaceName}
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80`,

      Pod: `apiVersion: v1
kind: Pod
metadata:
  name: ${name || 'my-pod'}
  namespace: ${namespaceName}
  labels:
    app: ${name || 'my-app'}
spec:
  containers:
  - name: ${name || 'my-app'}
    image: nginx:latest
    ports:
    - containerPort: 80`,
    };

    setYamlContent(templates[selectedType] || '');
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>创建资源</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {error && (
            <Alert severity="error">{error}</Alert>
          )}

          {/* Resource Type Selection */}
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              资源类型
            </Typography>
            <TextField
              select
              fullWidth
              size="small"
              value={selectedType}
              onChange={(e) => {
                setSelectedType(e.target.value);
                setYamlContent('');
              }}
            >
              {resourceTypes.map((type) => (
                <MenuItem key={type.value} value={type.value}>
                  {type.label}
                </MenuItem>
              ))}
            </TextField>
          </Box>

          {/* Basic Info */}
          <Stack direction="row" spacing={2}>
            <TextField
              fullWidth
              size="small"
              label="名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onBlur={generateTemplate}
            />
            <TextField
              fullWidth
              size="small"
              label="命名空间"
              value={namespaceName}
              onChange={(e) => setNamespaceName(e.target.value)}
            />
          </Stack>

          {/* Mode Selection */}
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              编辑模式
            </Typography>
            <Stack direction="row" spacing={1}>
              <Button
                size="small"
                variant={mode === 'form' ? 'contained' : 'outlined'}
                onClick={() => setMode('form')}
              >
                表单
              </Button>
              <Button
                size="small"
                variant={mode === 'yaml' ? 'contained' : 'outlined'}
                onClick={() => setMode('yaml')}
              >
                YAML
              </Button>
              <Button
                size="small"
                variant="outlined"
                onClick={generateTemplate}
              >
                生成模板
              </Button>
            </Stack>
          </Box>

          {/* YAML Editor */}
          {mode === 'yaml' && (
            <YamlEditor
              value={yamlContent}
              onChange={setYamlContent}
              height={400}
            />
          )}

          {/* Form Mode (simplified) */}
          {mode === 'form' && (
            <Alert severity="info">
              表单模式正在开发中，请使用 YAML 模式创建资源。
            </Alert>
          )}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>取消</Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={loading || !yamlContent}
        >
          创建
        </Button>
      </DialogActions>
    </Dialog>
  );
}
