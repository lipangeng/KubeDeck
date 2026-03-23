import { useState } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import { useNotification } from './NotificationProvider';

interface BatchActionsProps {
  open: boolean;
  onClose: () => void;
  items: Array<{ id: string; name: string }>;
  onBatchAction: (action: string, names: string[]) => Promise<void>;
}

export function BatchActions({ open, onClose, items, onBatchAction }: BatchActionsProps) {
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [action, setAction] = useState<string>('restart');
  const [loading, setLoading] = useState(false);
  const notification = useNotification();

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelected(new Set(items.map(item => item.id)));
    } else {
      setSelected(new Set());
    }
  };

  const handleSelect = (id: string, checked: boolean) => {
    const newSelected = new Set(selected);
    if (checked) {
      newSelected.add(id);
    } else {
      newSelected.delete(id);
    }
    setSelected(newSelected);
  };

  const handleExecute = async () => {
    if (selected.size === 0) {
      notification.showError('请选择至少一个项目');
      return;
    }

    setLoading(true);
    try {
      const names = items.filter(item => selected.has(item.id)).map(item => item.name);
      await onBatchAction(action, names);
      notification.showSuccess(`成功执行 ${action} 操作，共 ${names.length} 个`);
      handleClose();
    } catch (error) {
      notification.showError('批量操作失败：' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setSelected(new Set());
    setAction('restart');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>批量操作</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {/* Action Selection */}
          <Stack spacing={1}>
            <Typography variant="subtitle2">选择操作</Typography>
            <Stack direction="row" spacing={2}>
              <Button
                variant={action === 'restart' ? 'contained' : 'outlined'}
                onClick={() => setAction('restart')}
                size="small"
              >
                重启
              </Button>
              <Button
                variant={action === 'pause' ? 'contained' : 'outlined'}
                onClick={() => setAction('pause')}
                size="small"
              >
                暂停
              </Button>
              <Button
                variant={action === 'resume' ? 'contained' : 'outlined'}
                onClick={() => setAction('resume')}
                size="small"
              >
                恢复
              </Button>
            </Stack>
          </Stack>

          {/* Item Selection */}
          <Stack spacing={1}>
            <Typography variant="subtitle2">
              已选择 {selected.size} / {items.length} 个
            </Typography>
            <FormControlLabel
              control={
                <Checkbox
                  checked={selected.size === items.length && items.length > 0}
                  indeterminate={selected.size > 0 && selected.size < items.length}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              }
              label="全选"
            />
            <Stack spacing={0.5} sx={{ maxHeight: 300, overflow: 'auto' }}>
              {items.map((item) => (
                <FormControlLabel
                  key={item.id}
                  control={
                    <Checkbox
                      checked={selected.has(item.id)}
                      onChange={(e) => handleSelect(item.id, e.target.checked)}
                    />
                  }
                  label={item.name}
                />
              ))}
            </Stack>
          </Stack>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>取消</Button>
        <Button
          onClick={handleExecute}
          variant="contained"
          disabled={loading || selected.size === 0}
        >
          执行
        </Button>
      </DialogActions>
    </Dialog>
  );
}
