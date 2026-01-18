import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams, Link } from 'react-router-dom';
import {
  Anchor,
  Badge,
  Box,
  Breadcrumbs,
  Card,
  Code,
  Grid,
  Group,
  Pagination,
  ScrollArea,
  Select,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { IconSearch, IconBolt, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Action } from '../../lib/api';
import {
  CollapsibleCard,
  CopyButton,
  CopyEntityButton,
  EmptyState,
  ExportButton,
  FilterBar,
  ListSkeleton,
  DetailSkeleton,
} from '../../components';

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
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionKeys, setActionKeys] = useState<string[]>([]);
  const [selectedKey, setSelectedKey] = useState<string | null>(keyFromUrl || null);
  const [selectedAction, setSelectedAction] = useState<Action | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Initialize filters from URL params
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [serviceFilter, setServiceFilter] = useState<string | null>(searchParams.get('service'));

  // Sync filters to URL
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    if (serviceFilter) params.set('service', serviceFilter);
    setSearchParams(params, { replace: true });
  }, [searchQuery, serviceFilter, setSearchParams]);

  const hasActiveFilters = Boolean(searchQuery || serviceFilter);

  const clearAllFilters = (): void => {
    setSearchQuery('');
    setServiceFilter(null);
  };

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
      if (serviceFilter && a.service !== serviceFilter) return false;
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return a.name.toLowerCase().includes(query) || a.key.toLowerCase().includes(query);
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
    navigate(`/search/actions/${key}?${searchParams.toString()}`, { replace: true });
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

  // Breadcrumb items
  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Actions', href: '/search/actions' },
    ...(selectedAction ? [{ title: `${selectedAction.Service}:${selectedAction.Name}`, href: '' }] : []),
  ].map((item, index, arr) => {
    const isLast = index === arr.length - 1;
    if (isLast) {
      return <Text key={item.title} size="sm" c="dimmed">{item.title}</Text>;
    }
    return (
      <Anchor key={item.title} component={Link} to={item.href} size="sm">
        {item.title}
      </Anchor>
    );
  });

  if (error) {
    return (
      <Box p="md">
        <EmptyState variant="error" message={error} />
      </Box>
    );
  }

  return (
    <Box p="md" h="100%">
      <Grid gutter="md" h="100%">
        {/* Left column */}
        <Grid.Col span={5}>
          <Stack gap="md" h="100%">
            <Breadcrumbs separator={<IconChevronRight size={14} />}>{breadcrumbItems}</Breadcrumbs>
            <Title order={2}>Actions</Title>

            <TextInput
              placeholder="Search actions..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Select
                placeholder="All services"
                size="sm"
                data={serviceOptions}
                value={serviceFilter}
                onChange={setServiceFilter}
                clearable
                searchable
                style={{ flex: 1 }}
              />
            </FilterBar>

            <Text size="sm" c="dimmed">
              {formatNumber(filteredActions.length)} of {formatNumber(actionList.length)} actions
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredActions.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="action"
                    />
                  </Box>
                ) : (
                  paginatedActions.map((a) => (
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
                        <Text size="sm" fw={500} truncate>{a.name}</Text>
                        <Text size="xs" c="dimmed" truncate>{a.service}</Text>
                      </div>
                    </div>
                  ))
                )}
              </ScrollArea>
              {totalPages > 1 && (
                <Box p="xs" style={{ borderTop: '1px solid var(--mantine-color-default-border)' }}>
                  <Pagination value={page} onChange={setPage} total={totalPages} size="sm" withEdges />
                </Box>
              )}
            </Card>
          </Stack>
        </Grid.Col>

        {/* Right column */}
        <Grid.Col span={7}>
          <Card withBorder h="100%" p="md">
            {!selectedKey ? (
              <EmptyState variant="no-selection" entityName="action" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedAction ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedAction.Service}:{selectedAction.Name}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedAction} />
                      <ExportButton data={selectedAction} filename={`action-${selectedAction.Service}-${selectedAction.Name}`} />
                      <Badge color="yellow" size="lg">{selectedAction.Service}</Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Action:</Text>
                        <Text size="sm" ff="monospace" style={{ flex: 1 }}>{selectedAction.Name}</Text>
                        <CopyButton value={`${selectedAction.Service}:${selectedAction.Name}`} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Service:</Text>
                        <Text size="sm" ff="monospace">{selectedAction.Service}</Text>
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  {selectedAction.ActionConditionKeys && selectedAction.ActionConditionKeys.length > 0 && (
                    <CollapsibleCard title="Action Condition Keys">
                      <Group gap="xs">
                        {selectedAction.ActionConditionKeys.map((key) => (
                          <Code key={key}>{key}</Code>
                        ))}
                      </Group>
                    </CollapsibleCard>
                  )}

                  {selectedAction.ResolvedResources && selectedAction.ResolvedResources.length > 0 && (
                    <CollapsibleCard title="Resource Types">
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
                    </CollapsibleCard>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <EmptyState variant="error" message="Failed to load action details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
