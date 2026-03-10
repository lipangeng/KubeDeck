import { useEffect, useState } from 'react';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import { ListPageShell } from '../../../components/page-shell/ResourcePageShell';
import { copy } from '../../../i18n/copy';
import { fetchWorkloads, type WorkloadItem } from '../../runtime/fetchWorkloads';

export function WorkloadsPage() {
  const [items, setItems] = useState<WorkloadItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        const nextItems = await fetchWorkloads('default');
        if (active) {
          setItems(nextItems);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void load();
    return () => {
      active = false;
    };
  }, []);

  return (
    <ListPageShell
      title={copy('workloads.title')}
      toolbar={<Chip color="primary" label={copy('workloads.badge')} size="small" />}
    >
      <Stack spacing={1.5}>
        <Typography color="text.secondary">{copy('workloads.description')}</Typography>
        {loading ? (
          <Stack direction="row" spacing={1} alignItems="center">
            <CircularProgress size={18} />
            <Typography variant="body2" color="text.secondary">
              {copy('workloads.loading')}
            </Typography>
          </Stack>
        ) : (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Kind</TableCell>
                <TableCell>Namespace</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Health</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{item.kind}</TableCell>
                  <TableCell>{item.namespace}</TableCell>
                  <TableCell>{item.status}</TableCell>
                  <TableCell>{item.health}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        <Typography variant="body2" color="text.secondary">
          {copy('workloads.placeholder')}
        </Typography>
      </Stack>
    </ListPageShell>
  );
}
