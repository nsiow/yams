import { useCallback, useEffect, useState } from 'react';
import {
  ActionIcon,
  Alert,
  Badge,
  Card,
  Container,
  Group,
  Loader,
  SimpleGrid,
  Stack,
  Text,
  Title,
  Tooltip,
} from '@mantine/core';
import { IconAlertCircle, IconRefresh } from '@tabler/icons-react';
import { yamsApi } from '../lib/api';
import type { StatusResponse } from '../lib/api';

interface StatCardProps {
  label: string;
  value: number;
}

function StatCard({ label, value }: StatCardProps): JSX.Element {
  return (
    <Card shadow="sm" padding="lg" radius="md" withBorder>
      <Text size="xs" c="dimmed" tt="uppercase" fw={700}>
        {label}
      </Text>
      <Text size="xl" fw={700} mt="xs">
        {(value ?? 0).toLocaleString()}
      </Text>
    </Card>
  );
}

interface StatusIndicatorProps {
  healthy: boolean;
  label?: string;
}

function StatusIndicator({ healthy, label }: StatusIndicatorProps): JSX.Element {
  return (
    <Badge
      color={healthy ? 'green' : 'red'}
      variant="filled"
      size="lg"
    >
      {label ?? (healthy ? 'Healthy' : 'Unhealthy')}
    </Badge>
  );
}

function isSourceFresh(updatedTime: string): boolean {
  try {
    const updated = new Date(updatedTime);
    if (isNaN(updated.getTime())) return false;
    const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000);
    return updated > oneHourAgo;
  } catch {
    return false;
  }
}

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return 'Unknown';
    return date.toLocaleString();
  } catch {
    return 'Unknown';
  }
}

export function HomePage(): JSX.Element {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [healthy, setHealthy] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchData = useCallback(async (isRefresh = false): Promise<void> => {
    if (isRefresh) {
      setRefreshing(true);
    }
    try {
      // Fetch health first
      const healthResult = await yamsApi.healthcheck()
        .then(() => true)
        .catch(() => false);
      setHealthy(healthResult);

      // Then fetch status
      const statusResult = await yamsApi.status();
      setStatus(statusResult);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch status:', err);
      setError(err instanceof Error ? err.message : 'Failed to connect to server');
      setHealthy(false);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleRefresh = (): void => {
    fetchData(true);
  };

  if (loading) {
    return (
      <Container size="md" py="xl">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Connecting to yams server...</Text>
        </Stack>
      </Container>
    );
  }

  if (error) {
    return (
      <Container size="md" py="xl">
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Connection Error"
          color="red"
        >
          {error}
        </Alert>
      </Container>
    );
  }

  return (
    <Container size="md" py="xl">
      <Stack gap="lg">
        <Group justify="space-between" align="center">
          <Group>
            <Title order={1}>Dashboard</Title>
            <Tooltip label="Refresh">
              <ActionIcon
                variant="subtle"
                onClick={handleRefresh}
                loading={refreshing}
                aria-label="Refresh dashboard"
              >
                <IconRefresh size={20} />
              </ActionIcon>
            </Tooltip>
          </Group>
          <StatusIndicator healthy={healthy ?? false} />
        </Group>

        <Text c="dimmed">
          yams server status and entity counts
        </Text>

        {status && (
          <>
            <Title order={3} mt="md">Entity Counts</Title>
            <SimpleGrid cols={{ base: 2, sm: 3 }}>
              <StatCard label="Principals" value={status.principals} />
              <StatCard label="Resources" value={status.resources} />
              <StatCard label="Policies" value={status.policies} />
              <StatCard label="Groups" value={status.groups} />
              <StatCard label="Accounts" value={status.accounts} />
              <StatCard label="Total Entities" value={status.entities} />
            </SimpleGrid>

            <Title order={3} mt="xl">Data Sources</Title>
            {!status.sources || status.sources.length === 0 ? (
              <Text c="dimmed">No data sources configured</Text>
            ) : (
              <Stack gap="sm">
                {status.sources.map((src, index) => {
                  const fresh = isSourceFresh(src.updated);
                  return (
                    <Card key={src.source || index} shadow="sm" padding="md" radius="md" withBorder>
                      <Group justify="space-between">
                        <Stack gap={4}>
                          <Text fw={500}>{src.source}</Text>
                          <Text size="sm" c="dimmed">
                            Last updated: {formatTimestamp(src.updated)}
                          </Text>
                        </Stack>
                        <StatusIndicator
                          healthy={fresh}
                          label={fresh ? 'Fresh' : 'Stale'}
                        />
                      </Group>
                    </Card>
                  );
                })}
              </Stack>
            )}
          </>
        )}
      </Stack>
    </Container>
  );
}
