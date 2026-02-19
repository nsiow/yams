// ui/src/pages/config.tsx
import { useState } from 'react';
import {
  Box,
  Card,
  Stack,
  Switch,
  Text,
  Title,
} from '@mantine/core';

const SHARED_CONTEXT_KEY = 'yams.enable_shared_request_context';

export function ConfigPage(): JSX.Element {
  const [enabled, setEnabled] = useState<boolean>(
    () => localStorage.getItem(SHARED_CONTEXT_KEY) === 'true'
  );

  const handleToggle = (checked: boolean): void => {
    setEnabled(checked);
    localStorage.setItem(SHARED_CONTEXT_KEY, String(checked));
  };

  return (
    <Box p="md">
      <Stack gap="lg">
        <Box>
          <Title order={3} mb={4}>Configuration</Title>
          <Text size="sm" c="dimmed">
            Manage instance-wide settings for this yams UI.
          </Text>
        </Box>

        <Card withBorder p="lg">
          <Title order={5} mb="md">Shared Request Context</Title>
          <Switch
            label="Enable shared request context"
            description="When enabled, context key-value pairs configured by the server operator are automatically included in all simulation panes. User-defined context variables take precedence on key conflicts."
            checked={enabled}
            onChange={(e) => handleToggle(e.currentTarget.checked)}
          />
        </Card>
      </Stack>
    </Box>
  );
}
