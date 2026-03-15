import { useState } from 'react';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Typography from '@mui/material/Typography';
import type { ReactNode } from 'react';
import type { ResourcePageTab } from './types';

interface ResourcePageShellProps {
  title: string;
  summary?: ReactNode;
  actions?: ReactNode;
  tabs?: ResourcePageTab[];
  takeoverContent?: ReactNode;
}

export function ResourcePageShell({
  title,
  summary,
  actions,
  tabs = [],
  takeoverContent,
}: ResourcePageShellProps) {
  const [activeTabId, setActiveTabId] = useState(tabs[0]?.id ?? 'overview');
  const activeTab = tabs.find((tab) => tab.id === activeTabId) ?? tabs[0] ?? null;

  return (
    <Stack spacing={2}>
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack spacing={summary || actions ? 1 : 0}>
          <Typography variant="h6" component="h2">
            {title}
          </Typography>
          {summary ? <Box>{summary}</Box> : null}
          {actions ? <Box>{actions}</Box> : null}
        </Stack>
      </Paper>
      <Paper variant="outlined" sx={{ p: 2 }}>
        {takeoverContent ? (
          <Box>{takeoverContent}</Box>
        ) : (
          <>
            <Tabs
              value={activeTab?.id ?? false}
              onChange={(_, nextValue: string) => setActiveTabId(nextValue)}
            >
              {tabs.map((tab) => (
                <Tab key={tab.id} value={tab.id} label={tab.title} />
              ))}
            </Tabs>
            <Box sx={{ pt: 2 }}>{activeTab?.content}</Box>
          </>
        )}
      </Paper>
    </Stack>
  );
}
