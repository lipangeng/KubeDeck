import { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';

type NotificationType = 'success' | 'error' | 'warning' | 'info';

interface Notification {
  id: string;
  type: NotificationType;
  message: string;
  title?: string;
}

interface NotificationContextType {
  showSuccess: (message: string) => void;
  showError: (message: string) => void;
  showWarning: (message: string) => void;
  showInfo: (message: string) => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [current, setCurrent] = useState<Notification | null>(null);

  const showNotification = useCallback((type: NotificationType, message: string) => {
    const id = Date.now().toString();
    const notification: Notification = { id, type, message };
    setNotifications(prev => [...prev, notification]);
    setCurrent(notification);
  }, []);

  const handleClose = useCallback(() => {
    setCurrent(null);
    setNotifications(prev => prev.slice(1));
  }, []);

  const showSuccess = useCallback((message: string) => {
    showNotification('success', message);
  }, [showNotification]);

  const showError = useCallback((message: string) => {
    showNotification('error', message);
  }, [showNotification]);

  const showWarning = useCallback((message: string) => {
    showNotification('warning', message);
  }, [showNotification]);

  const showInfo = useCallback((message: string) => {
    showNotification('info', message);
  }, [showNotification]);

  return (
    <NotificationContext.Provider value={{ showSuccess, showError, showWarning, showInfo }}>
      {children}
      <Snackbar
        open={!!current}
        autoHideDuration={6000}
        onClose={handleClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert
          onClose={handleClose}
          severity={current?.type}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {current?.message}
        </Alert>
      </Snackbar>
    </NotificationContext.Provider>
  );
}

export function useNotification() {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotification must be used within a NotificationProvider');
  }
  return context;
}
