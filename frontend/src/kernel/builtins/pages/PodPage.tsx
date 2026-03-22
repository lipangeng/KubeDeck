import { useEffect, useState, useRef } from 'react';
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
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import Alert from '@mui/material/Alert';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { AttachAddon } from '@xterm/addon-attach';
import '@xterm/xterm/css/xterm.css';
import { copy } from '../../../i18n/copy';
import { useKernelRuntime } from '../../runtime/KernelRuntimeContext';

interface PodDetail {
  name: string;
  namespace: string;
  status: string;
  node: string;
  ip: string;
  startTime: string;
  containers: ContainerDetail[];
}

interface ContainerDetail {
  name: string;
  image: string;
  status: string;
  restarts: number;
  ready: boolean;
}

interface LogLine {
  timestamp: string;
  message: string;
}

export function PodPage() {
  const { activeCluster, namespaceScope } = useKernelRuntime();
  const [pod, setPod] = useState<PodDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState(0);
  const [selectedContainer, setSelectedContainer] = useState('');
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [logFollow, setLogFollow] = useState(true);
  const [logFilter, setLogFilter] = useState('');
  const [terminalOpen, setTerminalOpen] = useState(false);
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);

  useEffect(() => {
    async function fetchPodDetail() {
      try {
        const params = new URLSearchParams(window.location.search);
        const name = params.get('name') || 'api-7c9d8';
        const namespace = params.get('namespace') || 'default';
        
        // Mock data for demo
        const mockPod: PodDetail = {
          name: name,
          namespace: namespace,
          status: 'Running',
          node: 'node-1',
          ip: '10.244.1.15',
          startTime: new Date().toISOString(),
          containers: [
            { name: 'api', image: 'myapp:v1.2.3', status: 'Running', restarts: 0, ready: true },
            { name: 'sidecar', image: 'fluent-bit', status: 'Running', restarts: 1, ready: true },
          ],
        };
        setPod(mockPod);
        if (mockPod.containers.length > 0) {
          setSelectedContainer(mockPod.containers[0].name);
        }
      } catch (error) {
        console.error('Failed to fetch pod detail:', error);
      } finally {
        setLoading(false);
      }
    }

    void fetchPodDetail();
  }, [activeCluster]);

  useEffect(() => {
    if (activeTab === 1 && selectedContainer) {
      // Load mock logs
      const mockLogs: LogLine[] = [
        { timestamp: new Date().toISOString(), message: 'INFO Starting application...' },
        { timestamp: new Date().toISOString(), message: 'INFO Connected to database' },
        { timestamp: new Date().toISOString(), message: 'WARN High memory usage detected' },
        { timestamp: new Date().toISOString(), message: 'ERROR Failed to connect to cache' },
      ];
      setLogs(mockLogs);
    }
  }, [activeTab, selectedContainer]);

  useEffect(() => {
    if (activeTab === 2 && terminalRef.current && !terminalOpen) {
      const term = new Terminal({
        cursorBlink: true,
        theme: {
          background: '#1e1e1e',
          foreground: '#ffffff',
        },
      });
      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);
      term.open(terminalRef.current);
      fitAddon.fit();
      
      // Mock terminal content
      term.writeln('\x1b[32mroot@pod:/app#\x1b[0m ls -la');
      term.writeln('total 48');
      term.writeln('drwxr-xr-x  4 root root 4096 Mar 23 10:15 .');
      term.writeln('drwxr-xr-x  3 root root 4096 Mar 23 10:15 ..');
      term.writeln('-rw-r--r--  1 root root  234 Mar 23 10:15 main.go');
      term.writeln('\x1b[32mroot@pod:/app#\x1b[0m ');
      
      terminalInstance.current = term;
      setTerminalOpen(true);

      return () => {
        term.dispose();
        terminalInstance.current = null;
        setTerminalOpen(false);
      };
    }
  }, [activeTab]);

  if (loading) {
    return <Typography>Loading...</Typography>;
  }

  if (!pod) {
    return <Typography>Pod not found</Typography>;
  }

  const filteredLogs = logs.filter(log => 
    log.message.toLowerCase().includes(logFilter.toLowerCase())
  );

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6">
              {pod.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Namespace: {pod.namespace}
            </Typography>
          </Stack>
          <Stack direction="row" spacing={1}>
            <Chip label={`Status: ${pod.status}`} color={pod.status === 'Running' ? 'success' : 'warning'} />
            <Chip label={`Node: ${pod.node}`} variant="outlined" />
            <Chip label={`IP: ${pod.ip}`} variant="outlined" />
          </Stack>
        </Stack>
      </Paper>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)}>
          <Tab label={copy('pod.tabs.overview') || 'Overview'} />
          <Tab label={copy('pod.tabs.logs') || 'Logs'} />
          <Tab label={copy('pod.tabs.terminal') || 'Terminal'} />
          <Tab label={copy('pod.tabs.files') || 'Files'} />
          <Tab label={copy('pod.tabs.events') || 'Events'} />
          <Tab label="YAML" />
        </Tabs>
      </Box>

      {/* Overview Tab */}
      {activeTab === 0 && (
        <Stack spacing={2}>
          {/* Containers Table */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
              Containers ({pod.containers.length})
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Image</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Restarts</TableCell>
                  <TableCell>Ready</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {pod.containers.map((container) => (
                  <TableRow key={container.name}>
                    <TableCell>{container.name}</TableCell>
                    <TableCell>{container.image}</TableCell>
                    <TableCell>
                      <Chip label={container.status} size="small" color={container.status === 'Running' ? 'success' : 'default'} />
                    </TableCell>
                    <TableCell>{container.restarts}</TableCell>
                    <TableCell>
                      <Chip label={container.ready ? 'Ready' : 'Not Ready'} size="small" color={container.ready ? 'success' : 'error'} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Paper>

          {/* Environment Variables */}
          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 1 }}>
              Environment Variables
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Name</TableCell>
                  <TableCell>Value</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                <TableRow>
                  <TableCell>DATABASE_URL</TableCell>
                  <TableCell>postgres://db:5432/app</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>LOG_LEVEL</TableCell>
                  <TableCell>info</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </Paper>
        </Stack>
      )}

      {/* Logs Tab */}
      {activeTab === 1 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={2}>
            <Stack direction="row" spacing={1} alignItems="center" justifyContent="space-between">
              <Stack direction="row" spacing={1}>
                <FormControl size="small" sx={{ minWidth: 150 }}>
                  <InputLabel>Container</InputLabel>
                  <Select
                    value={selectedContainer}
                    label="Container"
                    onChange={(e) => setSelectedContainer(e.target.value)}
                  >
                    {pod.containers.map((c) => (
                      <MenuItem key={c.name} value={c.name}>{c.name}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <Button
                  variant={logFollow ? 'contained' : 'outlined'}
                  size="small"
                  onClick={() => setLogFollow(!logFollow)}
                >
                  {logFollow ? '⏸ Pause' : '▶ Follow'}
                </Button>
                <Button variant="outlined" size="small">⬇ Download</Button>
              </Stack>
              <TextField
                size="small"
                placeholder="Filter logs..."
                value={logFilter}
                onChange={(e) => setLogFilter(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">🔍</InputAdornment>
                  ),
                }}
                sx={{ width: 250 }}
              />
            </Stack>
            <Box
              sx={{
                bgcolor: '#1e1e1e',
                color: '#ffffff',
                p: 2,
                fontFamily: 'monospace',
                fontSize: '0.875rem',
                height: 400,
                overflow: 'auto',
              }}
            >
              {filteredLogs.map((log, index) => (
                <Box key={index} sx={{ mb: 0.5 }}>
                  <Box component="span" sx={{ color: '#6a9955' }}>{log.timestamp}</Box>{' '}
                  <Box component="span" sx={{ color: log.message.includes('ERROR') ? '#f48771' : log.message.includes('WARN') ? '#dcdcaa' : '#d4d4d4' }}>
                    {log.message}
                  </Box>
                </Box>
              ))}
            </Box>
          </Stack>
        </Paper>
      )}

      {/* Terminal Tab */}
      {activeTab === 2 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Stack spacing={2}>
            <Stack direction="row" spacing={1} alignItems="center" justifyContent="space-between">
              <Stack direction="row" spacing={1}>
                <FormControl size="small" sx={{ minWidth: 150 }}>
                  <InputLabel>Container</InputLabel>
                  <Select
                    value={selectedContainer}
                    label="Container"
                    onChange={(e) => setSelectedContainer(e.target.value)}
                  >
                    {pod.containers.map((c) => (
                      <MenuItem key={c.name} value={c.name}>{c.name}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <FormControl size="small" sx={{ minWidth: 100 }}>
                  <InputLabel>Shell</InputLabel>
                  <Select value="/bin/bash" label="Shell">
                    <MenuItem value="/bin/bash">/bin/bash</MenuItem>
                    <MenuItem value="/bin/sh">/bin/sh</MenuItem>
                  </Select>
                </FormControl>
              </Stack>
              <Stack direction="row" spacing={1}>
                <Button variant="outlined" size="small">Disconnect</Button>
                <Button variant="outlined" size="small">Clear</Button>
                <Button variant="outlined" size="small">Fullscreen</Button>
              </Stack>
            </Stack>
            <Box
              ref={terminalRef}
              sx={{
                height: 400,
                '& .xterm': { height: '100%' },
              }}
            />
          </Stack>
        </Paper>
      )}

      {/* Files Tab */}
      {activeTab === 3 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography>File management will be implemented in Phase 3</Typography>
        </Paper>
      )}

      {/* Events Tab */}
      {activeTab === 4 && (
        <Paper variant="outlined" sx={{ p: 2 }}>
          <Typography>Events will be implemented in Phase 2</Typography>
        </Paper>
      )}
    </Stack>
  );
}
