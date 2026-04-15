// ui/src/pages/shared-variables.tsx
import { useEffect, useState } from 'react';
import {
  Anchor,
  Badge,
  Box,
  Card,
  Group,
  Loader,
  Stack,
  Table,
  Text,
  Title,
} from '@mantine/core';
import { Link } from 'react-router-dom';
import { yamsApi } from '../lib/api';

const SHARED_CONTEXT_KEY = 'yams.enable_shared_request_context';

export function SharedVariablesPage(): JSX.Element {
  const [loading, setLoading] = useState(true);
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const enabled = localStorage.getItem(SHARED_CONTEXT_KEY) === 'true';

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
            When enabled, these are automatically included in all simulation panes.
          </Text>
        </Box>

        {/* Feature status */}
        <Card withBorder p="lg">
          <Group justify="space-between" align="center">
            <Group gap="xs" align="center">
              <Badge color={enabled ? 'green' : 'red'} size="xs" circle />
              <Text size="sm" fw={500}>
                Shared request context is {enabled ? 'enabled' : 'disabled'}
              </Text>
            </Group>
            <Anchor component={Link} to="/config" size="sm">
              {enabled ? 'Manage in Config' : 'Enable in Config'}
            </Anchor>
          </Group>
        </Card>

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
