import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams, Link } from 'react-router-dom';
import {
  Anchor,
  Badge,
  Box,
  Breadcrumbs,
  Card,
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
import { CodeHighlight } from '@mantine/code-highlight';
import { IconSearch, IconUsers, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Group as GroupEntity } from '../../lib/api';
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

import '@mantine/code-highlight/styles.css';

interface GroupListItem {
  arn: string;
  accountId: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parseGroupArn(arn: string): GroupListItem {
  const parts = arn.split(':');
  const accountId = parts[4] || '';
  const resourcePart = parts[5] || '';
  const name = resourcePart.split('/').slice(1).join('/') || '';
  return { arn, accountId, name };
}

export function GroupsPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': arnFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [groupArns, setGroupArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedGroup, setSelectedGroup] = useState<GroupEntity | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Filters
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [accountFilter, setAccountFilter] = useState<string | null>(searchParams.get('account'));

  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    if (accountFilter) params.set('account', accountFilter);
    setSearchParams(params, { replace: true });
  }, [searchQuery, accountFilter, setSearchParams]);

  const hasActiveFilters = Boolean(searchQuery || accountFilter);

  const clearAllFilters = (): void => {
    setSearchQuery('');
    setAccountFilter(null);
  };

  const groupList = useMemo(() => {
    return groupArns.map(parseGroupArn);
  }, [groupArns]);

  const accountOptions = useMemo(() => {
    const ids = new Set(groupList.map(g => g.accountId));
    return Array.from(ids)
      .sort()
      .map(id => {
        const name = accountNames[id];
        const label = name ? `${name} (${id})` : id;
        return { value: id, label };
      });
  }, [groupList, accountNames]);

  const formatAccount = (accountId: string): string => {
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  const filteredGroups = useMemo(() => {
    return groupList.filter(g => {
      if (accountFilter && g.accountId !== accountFilter) return false;
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return g.name.toLowerCase().includes(query) || g.arn.toLowerCase().includes(query);
      }
      return true;
    });
  }, [groupList, accountFilter, debouncedSearch]);

  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [arns, names] = await Promise.all([
          yamsApi.listGroups(),
          yamsApi.accountNames(),
        ]);
        setGroupArns(arns);
        setAccountNames(names);
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

  const fetchGroupDetail = useCallback(async (arn: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const group = await yamsApi.getGroup(arn);
      setSelectedGroup(group);
    } catch (err) {
      console.error('Failed to fetch group detail:', err);
      setSelectedGroup(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectGroup = (arn: string): void => {
    setSelectedArn(arn);
    fetchGroupDetail(arn);
    navigate(`/search/groups/${arn}?${searchParams.toString()}`, { replace: true });
  };

  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchGroupDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchGroupDetail, selectedArn]);

  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredGroups.length / itemsPerPage);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, accountFilter]);

  const paginatedGroups = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredGroups.slice(start, start + itemsPerPage);
  }, [filteredGroups, page]);

  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Groups', href: '/search/groups' },
    ...(selectedGroup ? [{ title: selectedGroup.Arn.split('/').pop() || '', href: '' }] : []),
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
            <Title order={2}>Groups</Title>

            <TextInput
              placeholder="Search groups..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Select
                placeholder="All accounts"
                size="sm"
                data={accountOptions}
                value={accountFilter}
                onChange={setAccountFilter}
                clearable
                searchable
                style={{ flex: 1 }}
              />
            </FilterBar>

            <Text size="sm" c="dimmed">
              {formatNumber(filteredGroups.length)} of {formatNumber(groupList.length)} groups
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredGroups.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="group"
                    />
                  </Box>
                ) : (
                  paginatedGroups.map((g) => (
                    <div
                      key={g.arn}
                      onClick={() => handleSelectGroup(g.arn)}
                      style={{
                        cursor: 'pointer',
                        padding: '8px 12px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        backgroundColor: selectedArn === g.arn
                          ? 'var(--mantine-color-primary-light)'
                          : undefined,
                        borderBottom: '1px solid var(--mantine-color-default-border)',
                      }}
                      onMouseEnter={(e) => {
                        if (selectedArn !== g.arn) {
                          e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (selectedArn !== g.arn) {
                          e.currentTarget.style.backgroundColor = '';
                        }
                      }}
                    >
                      <div style={{ flexShrink: 0 }}>
                        <IconUsers size={16} color="var(--mantine-color-teal-6)" />
                      </div>
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <Text size="sm" fw={500} truncate>{g.name}</Text>
                        <Text size="xs" c="dimmed" truncate>{formatAccount(g.accountId)}</Text>
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
            {!selectedArn ? (
              <EmptyState variant="no-selection" entityName="group" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedGroup ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedGroup.Arn.split('/').pop()}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedGroup} />
                      <ExportButton data={selectedGroup} filename={`group-${selectedGroup.Arn.split('/').pop()}`} />
                      <Badge color="teal" size="lg">Group</Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all', flex: 1 }}>
                          {selectedGroup.Arn}
                        </Text>
                        <CopyButton value={selectedGroup.Arn} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Anchor component={Link} to={`/search/accounts/${selectedGroup.AccountId}`} size="sm" ff="monospace">
                          {formatAccount(selectedGroup.AccountId || '')}
                        </Anchor>
                        <CopyButton value={selectedGroup.AccountId || ''} />
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  {selectedGroup.AttachedPolicies && selectedGroup.AttachedPolicies.length > 0 && (
                    <CollapsibleCard title="Attached Policies">
                      <Stack gap="xs">
                        {selectedGroup.AttachedPolicies.map((policy) => (
                          <Group key={policy} gap="xs">
                            <Anchor component={Link} to={`/search/policies/${policy}`} size="sm" ff="monospace" style={{ flex: 1 }}>
                              {policy}
                            </Anchor>
                            <CopyButton value={policy} />
                          </Group>
                        ))}
                      </Stack>
                    </CollapsibleCard>
                  )}

                  {selectedGroup.InlinePolicies && selectedGroup.InlinePolicies.length > 0 && (
                    <CollapsibleCard title="Inline Policies">
                      <Stack gap="md">
                        {selectedGroup.InlinePolicies.map((policy, index) => {
                          const { _Name, ...policyWithoutName } = policy;
                          const displayName = _Name || policy.Id || `Policy ${index + 1}`;
                          return (
                            <Box key={index}>
                              <Text size="sm" fw={600} mb="xs">{displayName}</Text>
                              <CodeHighlight code={JSON.stringify(policyWithoutName, null, 2)} language="json" withCopyButton />
                            </Box>
                          );
                        })}
                      </Stack>
                    </CollapsibleCard>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <EmptyState variant="error" message="Failed to load group details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
