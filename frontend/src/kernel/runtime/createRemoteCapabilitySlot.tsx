import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import { copy } from '../../i18n/copy';
import type { LocalizedText } from '../contracts/types';

export function createRemoteCapabilitySlot(title?: LocalizedText) {
  function RemoteCapabilitySlot() {
    return (
      <Alert severity="info">
        {title ? <Typography sx={{ fontWeight: 700 }}>{title.fallback}</Typography> : null}
        <Typography variant="body2">{copy('remoteCapability.slotBody')}</Typography>
      </Alert>
    );
  }

  return RemoteCapabilitySlot;
}
