// ui/src/pages/shared-variables.tsx
import { useEffect, useState } from 'react';
import {
  Box,
  Card,
  Group,
  Loader,
  Stack,
  Table,
  Text,
  Title,
} from '@mantine/core';
import { yamsApi } from '../lib/api';

export function SharedVariablesPage(): JSX.Element {
  const [loading, setLoading] = useState(true);
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    yamsApi.sharedContext()
      .then((ctx) => {
        setVariables(ctx);
      })
      .catch((err) => {
        console.error('Failed to fetch shared context:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch shared context');
      })
      .finally(() => setLoading(false));
  }, []);

  const entries = Object.entries(variables);

  return (
    <Box p="md">
      <Stack gap="lg">
        <Box>
          <Title order={3} mb={4}>Shared Variables</Title>
          <Text size="sm" c="dimmed">
            Request context variables configured by the server operator via <Text component="span" ff="monospace" fw={500}>-c</Text> flags.
            These are included in simulation requests by default.
          </Text>
        </Box>

        {/* Variables table */}
        <Card withBorder p="lg">
          <Title order={5} mb="md">Variables</Title>

          {loading && (
            <Group justify="center" p="xl">
              <Loader size="md" />
              <Text c="dimmed">Loading shared context...</Text>
            </Group>
          )}

          {error && (
            <Text c="red" size="sm">{error}</Text>
          )}

          {!loading && !error && entries.length === 0 && (
            <Text size="sm" c="dimmed">
              No shared variables configured. Use <Text component="span" ff="monospace">-c key=value</Text> when starting the server.
            </Text>
          )}

          {!loading && !error && entries.length > 0 && (
            <Table striped highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Key</Table.Th>
                  <Table.Th>Value</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {entries.map(([key, value]) => (
                  <Table.Tr key={key}>
                    <Table.Td>
                      <Text size="sm" ff="monospace">{key}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" ff="monospace">{value}</Text>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          )}
        </Card>
      </Stack>
    </Box>
  );
}
