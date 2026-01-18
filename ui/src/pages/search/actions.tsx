import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Alert,
  Badge,
  Box,
  Card,
  Code,
  Grid,
  Group,
  Loader,
  Pagination,
  ScrollArea,
  Select,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { IconAlertCircle, IconSearch, IconBolt } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Action } from '../../lib/api';

interface ActionListItem {
  key: string;
  service: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parseActionKey(key: string): ActionListItem {
  // Format: service:ActionName
  const colonIdx = key.indexOf(':');
  const service = colonIdx > 0 ? key.substring(0, colonIdx) : '';
  const name = colonIdx > 0 ? key.substring(colonIdx + 1) : key;
  return { key, service, name };
}

export function ActionsPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': keyFromUrl } = useParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionKeys, setActionKeys] = useState<string[]>([]);
  const [selectedKey, setSelectedKey] = useState<string | null>(keyFromUrl || null);
  const [selectedAction, setSelectedAction] = useState<Action | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [serviceFilter, setServiceFilter] = useState<string | null>(null);

  // Parse all action keys into list items
  const actionList = useMemo(() => {
    return actionKeys.map(parseActionKey);
  }, [actionKeys]);

  // Extract unique services for filter dropdown
  const serviceOptions = useMemo(() => {
    const services = new Set(actionList.map(a => a.service));
    return Array.from(services).sort().map(s => ({ value: s, label: s }));
  }, [actionList]);

  // Filter actions based on search and filters
  const filteredActions = useMemo(() => {
    return actionList.filter(a => {
      // Service filter
      if (serviceFilter && a.service !== serviceFilter) {
        return false;
      }
      // Search filter (searches name and key)
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return (
          a.name.toLowerCase().includes(query) ||
          a.key.toLowerCase().includes(query)
        );
      }
      return true;
    });
  }, [actionList, serviceFilter, debouncedSearch]);

  // Fetch all action keys on mount
  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const keys = await yamsApi.listActions();
        setActionKeys(keys);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch data:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  // Fetch action detail when selected
  const fetchActionDetail = useCallback(async (key: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const action = await yamsApi.getAction(key);
      setSelectedAction(action);
    } catch (err) {
      console.error('Failed to fetch action detail:', err);
      setSelectedAction(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectAction = (key: string): void => {
    setSelectedKey(key);
    fetchActionDetail(key);
    navigate(`/search/actions/${key}`, { replace: true });
  };

  // Load action from URL on mount or when URL changes
  useEffect(() => {
    if (keyFromUrl && keyFromUrl !== selectedKey) {
      setSelectedKey(keyFromUrl);
      fetchActionDetail(keyFromUrl);
    }
  }, [keyFromUrl, fetchActionDetail, selectedKey]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredActions.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, serviceFilter]);

  const paginatedActions = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredActions.slice(start, start + itemsPerPage);
  }, [filteredActions, page]);

  if (loading) {
    return (
      <Box p="xl">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Loading actions...</Text>
        </Stack>
      </Box>
    );
  }

  if (error) {
    return (
      <Box p="xl">
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Error"
          color="red"
        >
          {error}
        </Alert>
      </Box>
    );
  }

  return (
    <Box p="md" h="100%">
      <Grid gutter="md" h="100%">
        {/* Left column - Search and list */}
        <Grid.Col span={5}>
          <Stack gap="md" h="100%">
            <Title order={2}>Actions</Title>

            {/* Search box */}
            <TextInput
              placeholder="Search actions..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            {/* Service filter */}
            <Select
              placeholder="All services"
              size="sm"
              data={serviceOptions}
              value={serviceFilter}
              onChange={setServiceFilter}
              clearable
              searchable
            />

            {/* Results count */}
            <Text size="sm" c="dimmed">
              {formatNumber(filteredActions.length)} of {formatNumber(actionList.length)} actions
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            {/* Action list - paginated */}
            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {paginatedActions.map((a) => (
                  <div
                    key={a.key}
                    onClick={() => handleSelectAction(a.key)}
                    style={{
                      cursor: 'pointer',
                      padding: '8px 12px',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      backgroundColor: selectedKey === a.key
                        ? 'var(--mantine-color-primary-light)'
                        : undefined,
                      borderBottom: '1px solid var(--mantine-color-default-border)',
                    }}
                    onMouseEnter={(e) => {
                      if (selectedKey !== a.key) {
                        e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (selectedKey !== a.key) {
                        e.currentTarget.style.backgroundColor = '';
                      }
                    }}
                  >
                    <div style={{ flexShrink: 0 }}>
                      <IconBolt size={16} color="var(--mantine-color-yellow-6)" />
                    </div>
                    <div style={{ minWidth: 0, flex: 1 }}>
                      <Text size="sm" fw={500} truncate>
                        {a.name}
                      </Text>
                      <Text size="xs" c="dimmed" truncate>
                        {a.service}
                      </Text>
                    </div>
                  </div>
                ))}
              </ScrollArea>
              {totalPages > 1 && (
                <Box p="xs" style={{ borderTop: '1px solid var(--mantine-color-default-border)' }}>
                  <Pagination
                    value={page}
                    onChange={setPage}
                    total={totalPages}
                    size="sm"
                    withEdges
                  />
                </Box>
              )}
            </Card>
          </Stack>
        </Grid.Col>

        {/* Right column - Detail panel */}
        <Grid.Col span={7}>
          <Card withBorder h="100%" p="md">
            {!selectedKey ? (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Select an action to view details</Text>
              </Stack>
            ) : loadingDetail ? (
              <Stack align="center" justify="center" h="100%">
                <Loader size="md" />
                <Text c="dimmed">Loading details...</Text>
              </Stack>
            ) : selectedAction ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  {/* Header */}
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedAction.Service}:{selectedAction.Name}</Title>
                    <Badge color="yellow" size="lg">
                      {selectedAction.Service}
                    </Badge>
                  </Group>

                  {/* Metadata */}
                  <Card withBorder p="sm">
                    <Title order={5} mb="xs">Metadata</Title>
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Action:</Text>
                        <Text size="sm" ff="monospace">{selectedAction.Name}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Service:</Text>
                        <Text size="sm" ff="monospace">{selectedAction.Service}</Text>
                      </Group>
                    </Stack>
                  </Card>

                  {/* Action Condition Keys */}
                  {selectedAction.ActionConditionKeys && selectedAction.ActionConditionKeys.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Action Condition Keys</Title>
                      <Group gap="xs">
                        {selectedAction.ActionConditionKeys.map((key) => (
                          <Code key={key}>{key}</Code>
                        ))}
                      </Group>
                    </Card>
                  )}

                  {/* Resource Types */}
                  {selectedAction.ResolvedResources && selectedAction.ResolvedResources.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Resource Types</Title>
                      <Stack gap="md">
                        {selectedAction.ResolvedResources.map((resource) => (
                          <Box key={resource.Name}>
                            <Text size="sm" fw={600} mb="xs">{resource.Name}</Text>

                            {resource.ARNFormats && resource.ARNFormats.length > 0 && (
                              <Box mb="xs">
                                <Text size="xs" c="dimmed" mb={4}>ARN Formats:</Text>
                                <Stack gap={4}>
                                  {resource.ARNFormats.map((arn, idx) => (
                                    <Code key={idx} block style={{ fontSize: '12px' }}>{arn}</Code>
                                  ))}
                                </Stack>
                              </Box>
                            )}

                            {resource.ConditionKeys && resource.ConditionKeys.length > 0 && (
                              <Box>
                                <Text size="xs" c="dimmed" mb={4}>Resource Condition Keys:</Text>
                                <Group gap="xs">
                                  {resource.ConditionKeys.map((key) => (
                                    <Code key={key} style={{ fontSize: '11px' }}>{key}</Code>
                                  ))}
                                </Group>
                              </Box>
                            )}
                          </Box>
                        ))}
                      </Stack>
                    </Card>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Failed to load action details</Text>
              </Stack>
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
