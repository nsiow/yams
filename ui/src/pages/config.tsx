// ui/src/pages/config.tsx
import {
  Box,
  Card,
  Stack,
  Text,
  Title,
} from '@mantine/core';

export function ConfigPage(): JSX.Element {
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
          <Text size="sm" c="dimmed">
            Server-configured request context is included in simulation requests by default.
            Request-specific context values take precedence on key conflicts.
          </Text>
        </Card>
      </Stack>
    </Box>
  );
}
