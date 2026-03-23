import { useEffect, useState } from 'react';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Chip from '@mui/material/Chip';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';

interface Event {
  name: string;
  namespace: string;
  type: string;
  reason: string;
  message: string;
  count: number;
  first_seen: string;
  last_seen: string;
  involved_object: string;
}

export function EventsPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');

  useEffect(() => {
    loadEvents();
  }, []);

  const loadEvents = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/events?namespace=default');
      const data = await response.json();
      setEvents(data.events || []);
      if (data.error) {
      }
    } catch (error) {
    } finally {
      setLoading(false);
    }
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'Warning':
        return <ErrorIcon color="error" />;
      case 'Normal':
        return <InfoIcon color="info" />;
      default:
        return <WarningIcon color="warning" />;
    }
  };

  const filteredEvents = events.filter(event => {
    const matchesSearch = filter === '' || 
      event.message.toLowerCase().includes(filter.toLowerCase()) ||
      event.reason.toLowerCase().includes(filter.toLowerCase()) ||
      event.involved_object.toLowerCase().includes(filter.toLowerCase());
    
    const matchesType = typeFilter === 'all' || event.type === typeFilter;
    
    return matchesSearch && matchesType;
  });

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <Stack spacing={3}>
      {/* Header */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center" justifyContent="space-between">
          <Stack spacing={1}>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              事件查看器
            </Typography>
            <Typography variant="body2" color="text.secondary">
              查看 Kubernetes 集群事件
            </Typography>
          </Stack>
          <IconButton onClick={loadEvents} color="primary">
            <RefreshIcon />
          </IconButton>
        </Stack>
      </Paper>

      {/* Filters */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack direction="row" spacing={2} alignItems="center">
          <TextField
            size="small"
            placeholder="搜索事件..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
            sx={{ flexGrow: 1 }}
          />
          <TextField
            select
            size="small"
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            SelectProps={{ native: true }}
            sx={{ minWidth: 150 }}
          >
            <option value="all">所有类型</option>
            <option value="Normal">Normal</option>
            <option value="Warning">Warning</option>
          </TextField>
        </Stack>
      </Paper>

      {/* Events Table */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 700, mb: 2 }}>
          事件列表 ({filteredEvents.length})
        </Typography>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>类型</TableCell>
              <TableCell>原因</TableCell>
              <TableCell>对象</TableCell>
              <TableCell>消息</TableCell>
              <TableCell>计数</TableCell>
              <TableCell>最后发生</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredEvents.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  <Typography color="text.secondary" sx={{ py: 4 }}>
                    {loading ? '加载中...' : '暂无事件'}
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              filteredEvents.map((event, index) => (
                <TableRow key={index} hover>
                  <TableCell>
                    <Chip
                      icon={getEventIcon(event.type)}
                      label={event.type}
                      color={event.type === 'Warning' ? 'error' : 'info'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" fontWeight={600}>
                      {event.reason}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" fontFamily="monospace">
                      {event.involved_object}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ maxWidth: 400, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {event.message}
                    </Typography>
                  </TableCell>
                  <TableCell>{event.count}</TableCell>
                  <TableCell>
                    <Typography variant="caption" color="text.secondary">
                      {new Date(event.last_seen).toLocaleString('zh-CN')}
                    </Typography>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>
    </Stack>
  );
}
