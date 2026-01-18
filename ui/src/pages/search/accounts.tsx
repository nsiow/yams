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
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { IconSearch, IconCloud, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Account } from '../../lib/api';
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

// TODO(nsiow): Add tag filtering when tags are available in the API

interface AccountListItem {
  id: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

export function AccountsPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': idFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [accountIds, setAccountIds] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedId, setSelectedId] = useState<string | null>(idFromUrl || null);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Initialize filters from URL params
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  // TODO(nsiow): Add OU filter when we have a way to efficiently get OU data for all accounts

  // Sync filters to URL
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    setSearchParams(params, { replace: true });
  }, [searchQuery, setSearchParams]);

  const hasActiveFilters = Boolean(searchQuery);

  const clearAllFilters = (): void => {
    setSearchQuery('');
  };

  // Build account list with names
  const accountList = useMemo((): AccountListItem[] => {
    return accountIds.map(id => ({
      id,
      name: accountNames[id] || id,
    }));
  }, [accountIds, accountNames]);

  // Filter accounts based on search
  const filteredAccounts = useMemo(() => {
    return accountList.filter(a => {
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return a.name.toLowerCase().includes(query) || a.id.toLowerCase().includes(query);
      }
      return true;
    });
  }, [accountList, debouncedSearch]);

  // Fetch all account IDs and names on mount
  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [ids, names] = await Promise.all([
          yamsApi.listAccounts(),
          yamsApi.accountNames(),
        ]);
        setAccountIds(ids);
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

  // Fetch account detail when selected
  const fetchAccountDetail = useCallback(async (id: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const account = await yamsApi.getAccount(id);
      setSelectedAccount(account);
    } catch (err) {
      console.error('Failed to fetch account detail:', err);
      setSelectedAccount(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectAccount = (id: string): void => {
    setSelectedId(id);
    fetchAccountDetail(id);
    navigate(`/search/accounts/${id}?${searchParams.toString()}`, { replace: true });
  };

  // Load account from URL on mount or when URL changes
  useEffect(() => {
    if (idFromUrl && idFromUrl !== selectedId) {
      setSelectedId(idFromUrl);
      fetchAccountDetail(idFromUrl);
    }
  }, [idFromUrl, fetchAccountDetail, selectedId]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredAccounts.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch]);

  const paginatedAccounts = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredAccounts.slice(start, start + itemsPerPage);
  }, [filteredAccounts, page]);

  // Get the leaf OU name for an account (last OU before the account itself)
  const getLeafOuName = (account: Account): string | null => {
    if (!account.OrgNodes) return null;
    const ous = account.OrgNodes.filter(n => n.Type === 'ORGANIZATIONAL_UNIT');
    return ous.length > 0 ? ous[ous.length - 1].Name : null;
  };

  // Breadcrumb items
  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Accounts', href: '/search/accounts' },
    ...(selectedAccount ? [{ title: selectedAccount.Name, href: '' }] : []),
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
            <Title order={2}>Accounts</Title>

            <TextInput
              placeholder="Search accounts..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Text size="sm" c="dimmed">Search by name or ID</Text>
            </FilterBar>

            <Text size="sm" c="dimmed">
              {formatNumber(filteredAccounts.length)} of {formatNumber(accountList.length)} accounts
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredAccounts.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="account"
                    />
                  </Box>
                ) : (
                  paginatedAccounts.map((a) => (
                    <div
                      key={a.id}
                      onClick={() => handleSelectAccount(a.id)}
                      style={{
                        cursor: 'pointer',
                        padding: '8px 12px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        backgroundColor: selectedId === a.id
                          ? 'var(--mantine-color-primary-light)'
                          : undefined,
                        borderBottom: '1px solid var(--mantine-color-default-border)',
                      }}
                      onMouseEnter={(e) => {
                        if (selectedId !== a.id) {
                          e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (selectedId !== a.id) {
                          e.currentTarget.style.backgroundColor = '';
                        }
                      }}
                    >
                      <div style={{ flexShrink: 0 }}>
                        <IconCloud size={16} color="var(--mantine-color-indigo-6)" />
                      </div>
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <Text size="sm" fw={500} truncate>{a.name}</Text>
                        <Text size="xs" c="dimmed" truncate>{a.id}</Text>
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
            {!selectedId ? (
              <EmptyState variant="no-selection" entityName="account" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedAccount ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedAccount.Name}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedAccount} />
                      <ExportButton data={selectedAccount} filename={`account-${selectedAccount.Name}`} />
                      <Badge color="indigo" size="lg">Account</Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={120}>Account ID:</Text>
                        <Text size="sm" ff="monospace" style={{ flex: 1 }}>{selectedAccount.Id}</Text>
                        <CopyButton value={selectedAccount.Id} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={120}>Name:</Text>
                        <Text size="sm" ff="monospace">{selectedAccount.Name}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={120}>Organization:</Text>
                        <Text size="sm" ff="monospace" style={{ flex: 1 }}>{selectedAccount.OrgId}</Text>
                        <CopyButton value={selectedAccount.OrgId} />
                      </Group>
                      {getLeafOuName(selectedAccount) && (
                        <Group gap="xs">
                          <Text size="sm" fw={600} c="dimmed" w={120}>Parent OU:</Text>
                          <Text size="sm" ff="monospace">{getLeafOuName(selectedAccount)}</Text>
                        </Group>
                      )}
                    </Stack>
                  </CollapsibleCard>

                  {selectedAccount.OrgNodes && selectedAccount.OrgNodes.length > 0 && (
                    <CollapsibleCard title="Organization Hierarchy">
                      <Stack gap="md">
                        {selectedAccount.OrgNodes.map((node, idx) => (
                          <Box
                            key={node.Id}
                            pl={idx * 16}
                            style={{
                              borderLeft: idx > 0 ? '2px solid var(--mantine-color-gray-3)' : undefined,
                            }}
                          >
                            <Group gap="xs" mb={4}>
                              <Badge
                                size="xs"
                                color={
                                  node.Type === 'ROOT' ? 'red' :
                                  node.Type === 'ORGANIZATIONAL_UNIT' ? 'blue' :
                                  'green'
                                }
                                variant="light"
                              >
                                {node.Type === 'ORGANIZATIONAL_UNIT' ? 'OU' : node.Type}
                              </Badge>
                              <Text size="sm" fw={500}>{node.Name}</Text>
                              <Text size="xs" c="dimmed" ff="monospace">({node.Id})</Text>
                            </Group>

                            {/* SCPs */}
                            {node.SCPs && node.SCPs.length > 0 && (
                              <Box ml="md" mb={4}>
                                <Text size="xs" c="dimmed" mb={2}>SCPs:</Text>
                                <Stack gap={2}>
                                  {node.SCPs.map((scp) => (
                                    <Anchor
                                      key={scp}
                                      component={Link}
                                      to={`/search/policies/${scp}`}
                                      size="xs"
                                      ff="monospace"
                                      style={{ wordBreak: 'break-all' }}
                                    >
                                      {scp.split('/').pop()}
                                    </Anchor>
                                  ))}
                                </Stack>
                              </Box>
                            )}

                            {/* RCPs */}
                            {node.RCPs && node.RCPs.length > 0 && (
                              <Box ml="md">
                                <Text size="xs" c="dimmed" mb={2}>RCPs:</Text>
                                <Stack gap={2}>
                                  {node.RCPs.map((rcp) => (
                                    <Anchor
                                      key={rcp}
                                      component={Link}
                                      to={`/search/policies/${rcp}`}
                                      size="xs"
                                      ff="monospace"
                                      style={{ wordBreak: 'break-all' }}
                                    >
                                      {rcp.split('/').pop()}
                                    </Anchor>
                                  ))}
                                </Stack>
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
              <EmptyState variant="error" message="Failed to load account details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
