import { useState, useRef, useEffect } from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import Editor from '@monaco-editor/react';

interface YamlEditorProps {
  value: string;
  onChange?: (value: string) => void;
  onSave?: (value: string) => void;
  readOnly?: boolean;
  height?: string | number;
}

export function YamlEditor({ 
  value, 
  onChange, 
  onSave,
  readOnly = false,
  height = 400 
}: YamlEditorProps) {
  const [error, setError] = useState<string | null>(null);
  const editorRef = useRef<any>(null);

  const handleEditorMount = (editor: any) => {
    editorRef.current = editor;
  };

  const handleValidate = () => {
    // Basic YAML validation
    try {
      // Simple validation - check for basic YAML structure
      const lines = value.split('\n');
      let hasContent = false;
      
      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed && !trimmed.startsWith('#')) {
          hasContent = true;
          break;
        }
      }
      
      if (!hasContent) {
        setError('YAML 内容为空');
      } else {
        setError(null);
      }
    } catch (err) {
      setError('YAML 格式错误：' + (err as Error).message);
    }
  };

  const handleSave = () => {
    if (onSave && !error) {
      onSave(value);
    }
  };

  return (
    <Box>
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
        <Typography variant="subtitle2" fontWeight={600}>
          YAML 编辑器
        </Typography>
        <Box sx={{ flexGrow: 1 }} />
        {!readOnly && (
          <>
            <Button size="small" variant="outlined" onClick={handleValidate}>
              验证
            </Button>
            <Button 
              size="small" 
              variant="contained" 
              onClick={handleSave}
              disabled={!!error}
            >
              保存
            </Button>
          </>
        )}
      </Stack>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Box
        sx={{
          border: 1,
          borderColor: 'divider',
          borderRadius: 1,
          overflow: 'hidden',
        }}
      >
        <Editor
          height={height}
          language="yaml"
          value={value}
          onChange={(val) => onChange?.(val || '')}
          onMount={handleEditorMount}
          options={{
            readOnly,
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: 'on',
          }}
          theme="vs-dark"
        />
      </Box>
    </Box>
  );
}
