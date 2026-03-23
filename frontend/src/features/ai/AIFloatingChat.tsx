import { useState, useRef, useEffect } from 'react';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import IconButton from '@mui/material/IconButton';
import InputBase from '@mui/material/InputBase';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Avatar from '@mui/material/Avatar';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemAvatar from '@mui/material/ListItemAvatar';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';
import SendIcon from '@mui/icons-material/Send';
import ChatIcon from '@mui/icons-material/Chat';
import CloseIcon from '@mui/icons-material/Close';
import MinimizeIcon from '@mui/icons-material/Minimize';
import SettingsIcon from '@mui/icons-material/Settings';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import { useNavigate } from 'react-router-dom';

interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  action?: string;
  actionData?: Record<string, unknown>;
}

interface ApprovalDialogProps {
  open: boolean;
  command: string;
  security: string;
  onApprove: () => void;
  onReject: () => void;
}

function ApprovalDialog({ open, command, security, onApprove, onReject }: ApprovalDialogProps) {
  return (
    <Dialog 
      open={open} 
      onClose={onReject} 
      maxWidth="sm" 
      fullWidth
      aria-labelledby="approval-dialog-title"
      aria-describedby="approval-dialog-description"
    >
      <DialogTitle id="approval-dialog-title">Command Approval Required</DialogTitle>
      <DialogContent id="approval-dialog-description">
        <Alert severity={security === 'red' ? 'error' : 'warning'} role="alert">
          <AlertTitle>Security Level: {security.toUpperCase()}</AlertTitle>
          This command requires approval before execution:
          <Box component="pre" sx={{ mt: 1, p: 1, bgcolor: 'rgba(0,0,0,0.08)', borderRadius: 1 }}>
            <code>{command}</code>
          </Box>
        </Alert>
      </DialogContent>
      <DialogActions>
        <Button onClick={onReject} color="error">Reject</Button>
        <Button onClick={onApprove} variant="contained" color="success">Approve</Button>
      </DialogActions>
    </Dialog>
  );
}

export function AIFloatingChat() {
  const [open, setOpen] = useState(false);
  const [minimized, setMinimized] = useState(false);
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      role: 'assistant',
      content: 'Hi! I\'m your Kubernetes assistant. I can help you:\n\n• View cluster resources (pods, deployments, etc.)\n• Navigate to different pages\n• Execute kubectl commands (with approval for write operations)\n\nTry asking: "Show me the pods" or "Open the clusters page"',
      timestamp: new Date(),
    },
  ]);
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const [approvalOpen, setApprovalOpen] = useState(false);
  const [pendingCommand, setPendingCommand] = useState({ command: '', security: 'yellow' });
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const sendMessage = async (content: string) => {
    if (!content.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInputValue('');
    setLoading(true);

    try {
      const response = await fetch('/api/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: content }),
      });

      const data = await response.json();

      if (data.action === 'navigate' && data.action_data?.route) {
        navigate(data.action_data.route);
      } else if (data.action === 'approve') {
        setPendingCommand({
          command: content,
          security: 'yellow',
        });
        setApprovalOpen(true);
      }

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: data.message || 'I received your message.',
        timestamp: new Date(),
        action: data.action,
        actionData: data.action_data,
      };

      setMessages((prev) => [...prev, assistantMessage]);
    } catch (error) {
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = () => {
    setApprovalOpen(false);
    // Execute the approved command
    sendMessage(`Execute: ${pendingCommand.command}`);
  };

  const handleReject = () => {
    setApprovalOpen(false);
    const rejectMessage: Message = {
      id: Date.now().toString(),
      role: 'assistant',
      content: 'Command execution was rejected.',
      timestamp: new Date(),
    };
    setMessages((prev) => [...prev, rejectMessage]);
  };

  const quickActions = [
    { label: 'Show Pods', query: 'Show me the pods' },
    { label: 'Show Deployments', query: 'List deployments' },
    { label: 'Cluster Info', query: 'Cluster information' },
    { label: 'Open Clusters', query: 'Open clusters page' },
  ];

  return (
    <>
      {/* Floating Button */}
      {!open && (
        <IconButton
          color="primary"
          onClick={() => setOpen(true)}
          sx={{
            position: 'fixed',
            bottom: 24,
            right: 24,
            width: 64,
            height: 64,
            boxShadow: 3,
            zIndex: 1200,
          }}
          aria-label="Open AI assistant chat"
          aria-expanded="false"
        >
          <ChatIcon sx={{ fontSize: 32 }} />
        </IconButton>
      )}

      {/* Chat Window */}
      {open && (
        <Paper
          elevation={6}
          sx={{
            position: 'fixed',
            bottom: minimized ? 24 : 80,
            right: 24,
            width: minimized ? 300 : 400,
            height: minimized ? 56 : 550,
            zIndex: 1200,
            display: 'flex',
            flexDirection: 'column',
            transition: 'all 0.3s ease',
            overflow: 'hidden',
          }}
          role="dialog"
          aria-label="AI Assistant Chat"
          aria-modal={!minimized}
        >
          {/* Header */}
          <Box
            sx={{
              p: 1.5,
              bgcolor: 'primary.main',
              color: 'primary.contrastText',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
            role="banner"
          >
            <Stack direction="row" spacing={1} alignItems="center">
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.light' }}>
                <ChatIcon />
              </Avatar>
              <Typography variant="subtitle2" sx={{ fontWeight: 600 }} id="chat-title">
                AI Assistant
              </Typography>
            </Stack>
            <Stack direction="row" spacing={0.5}>
              <IconButton
                size="small"
                onClick={() => setMinimized(!minimized)}
                sx={{ color: 'inherit' }}
                aria-label={minimized ? 'Expand chat window' : 'Minimize chat window'}
                aria-controls="chat-content"
              >
                <MinimizeIcon />
              </IconButton>
              <IconButton
                size="small"
                onClick={() => setOpen(false)}
                sx={{ color: 'inherit' }}
                aria-label="Close chat window"
              >
                <CloseIcon />
              </IconButton>
            </Stack>
          </Box>

          {/* Messages */}
          {!minimized && (
            <>
              <Box
                sx={{
                  flex: 1,
                  overflow: 'auto',
                  p: 2,
                  bgcolor: 'background.default',
                }}
                id="chat-content"
                role="log"
                aria-label="Chat messages"
                aria-live="polite"
              >
                <List sx={{ p: 0 }} role="list">
                  {messages.map((msg) => (
                    <ListItem
                      key={msg.id}
                      alignItems="flex-start"
                      sx={{
                        px: 0,
                        flexDirection: msg.role === 'user' ? 'row-reverse' : 'row',
                      }}
                      role="listitem"
                    >
                      <ListItemAvatar>
                        <Avatar
                          sx={{
                            width: 32,
                            height: 32,
                            bgcolor: msg.role === 'user' ? 'primary.main' : 'secondary.main',
                          }}
                          aria-label={msg.role === 'user' ? 'User avatar' : 'AI assistant avatar'}
                        >
                          {msg.role === 'user' ? 'U' : 'AI'}
                        </Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={
                          <Box
                            sx={{
                              p: 1.5,
                              borderRadius: 2,
                              bgcolor: msg.role === 'user' ? 'primary.light' : 'grey.100',
                              display: 'inline-block',
                              maxWidth: '80%',
                              whiteSpace: 'pre-wrap',
                            }}
                          >
                            <Typography variant="body2">{msg.content}</Typography>
                          </Box>
                        }
                        secondary={
                          <Typography variant="caption" color="text.secondary">
                            {msg.timestamp.toLocaleTimeString()}
                          </Typography>
                        }
                      />
                    </ListItem>
                  ))}
                  {loading && (
                    <ListItem sx={{ px: 0 }}>
                      <ListItemAvatar>
                        <Avatar sx={{ width: 32, height: 32, bgcolor: 'secondary.main' }}>
                          AI
                        </Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={
                          <Box sx={{ p: 1.5, bgcolor: 'grey.100', borderRadius: 2, display: 'inline-block' }}>
                            <Typography variant="body2" role="status">Thinking...</Typography>
                          </Box>
                        }
                      />
                    </ListItem>
                  )}
                  <div ref={messagesEndRef} />
                </List>
              </Box>

              {/* Quick Actions */}
              <Divider />
              <Box sx={{ p: 1, bgcolor: 'background.paper' }} role="group" aria-label="Quick actions">
                <Stack direction="row" spacing={0.5} sx={{ overflow: 'auto' }}>
                  {quickActions.map((action) => (
                    <Button
                      key={action.label}
                      size="small"
                      variant="outlined"
                      onClick={() => sendMessage(action.query)}
                      sx={{ whiteSpace: 'nowrap' }}
                      aria-label={`Quick action: ${action.label}`}
                    >
                      {action.label}
                    </Button>
                  ))}
                </Stack>
              </Box>

              {/* Input */}
              <Divider />
              <Box
                component="form"
                onSubmit={(e) => {
                  e.preventDefault();
                  sendMessage(inputValue);
                }}
                sx={{
                  p: 1,
                  display: 'flex',
                  alignItems: 'center',
                  bgcolor: 'background.paper',
                }}
                role="searchbox"
                aria-label="Chat input form"
              >
                <InputBase
                  fullWidth
                  placeholder="Ask me anything..."
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  disabled={loading}
                  sx={{ ml: 1, flex: 1 }}
                  aria-label="Type your message"
                  aria-describedby="chat-input-help"
                />
                <IconButton
                  type="submit"
                  color="primary"
                  disabled={loading || !inputValue.trim()}
                  aria-label="Send message"
                >
                  <SendIcon />
                </IconButton>
                <span id="chat-input-help" hidden>Press Enter to send your message</span>
              </Box>
            </>
          )}
        </Paper>
      )}

      {/* Approval Dialog */}
      <ApprovalDialog
        open={approvalOpen}
        command={pendingCommand.command}
        security={pendingCommand.security}
        onApprove={handleApprove}
        onReject={handleReject}
      />
    </>
  );
}
