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
  Table,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import { IconSearch, IconUser, IconMask, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Principal } from '../../lib/api';
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

interface PrincipalListItem {
  arn: string;
  type: 'user' | 'role';
  accountId: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parsePrincipalArn(arn: string): PrincipalListItem {
  const parts = arn.split(':');
  const accountId = parts[4] || '';
  const resourcePart = parts[5] || '';
  const [typeStr, ...nameParts] = resourcePart.split('/');
  const name = nameParts.join('/') || '';
  const type = typeStr === 'role' ? 'role' : 'user';
  return { arn, type, accountId, name };
}

export function PrincipalsPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': arnFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [principalArns, setPrincipalArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedPrincipal, setSelectedPrincipal] = useState<Principal | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Initialize filters from URL params
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [typeFilter, setTypeFilter] = useState<string | null>(searchParams.get('type'));
  const [accountFilter, setAccountFilter] = useState<string | null>(searchParams.get('account'));

  // Sync filters to URL
  useEffect(() => {
    const params = new URLSearchParams();
    if (searchQuery) params.set('q', searchQuery);
    if (typeFilter) params.set('type', typeFilter);
    if (accountFilter) params.set('account', accountFilter);
    setSearchParams(params, { replace: true });
  }, [searchQuery, typeFilter, accountFilter, setSearchParams]);

  const hasActiveFilters = Boolean(searchQuery || typeFilter || accountFilter);

  const clearAllFilters = (): void => {
    setSearchQuery('');
    setTypeFilter(null);
    setAccountFilter(null);
  };

  const principalList = useMemo(() => {
    return principalArns.map(parsePrincipalArn);
  }, [principalArns]);

  const accountOptions = useMemo(() => {
    const ids = new Set(principalList.map(p => p.accountId));
    return Array.from(ids)
      .sort()
      .map(id => {
        const name = accountNames[id];
        const label = name ? `${name} (${id})` : id;
        return { value: id, label };
      });
  }, [principalList, accountNames]);

  const formatAccount = (accountId: string): string => {
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  const filteredPrincipals = useMemo(() => {
    return principalList.filter(p => {
      if (typeFilter && p.type !== typeFilter) return false;
      if (accountFilter && p.accountId !== accountFilter) return false;
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return p.name.toLowerCase().includes(query) || p.arn.toLowerCase().includes(query);
      }
      return true;
    });
  }, [principalList, typeFilter, accountFilter, debouncedSearch]);

  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [arns, names] = await Promise.all([
          yamsApi.listPrincipals(),
          yamsApi.accountNames(),
        ]);
        setPrincipalArns(arns);
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

  const fetchPrincipalDetail = useCallback(async (arn: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const principal = await yamsApi.getPrincipal(arn);
      setSelectedPrincipal(principal);
    } catch (err) {
      console.error('Failed to fetch principal detail:', err);
      setSelectedPrincipal(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectPrincipal = (arn: string): void => {
    setSelectedArn(arn);
    fetchPrincipalDetail(arn);
    navigate(`/search/principals/${arn}?${searchParams.toString()}`, { replace: true });
  };

  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchPrincipalDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchPrincipalDetail, selectedArn]);

  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredPrincipals.length / itemsPerPage);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, typeFilter, accountFilter]);

  const paginatedPrincipals = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredPrincipals.slice(start, start + itemsPerPage);
  }, [filteredPrincipals, page]);

  // Breadcrumb items
  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Principals', href: '/search/principals' },
    ...(selectedPrincipal ? [{ title: selectedPrincipal.Name, href: '' }] : []),
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
            <Title order={2}>Principals</Title>

            <TextInput
              placeholder="Search principals..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Select
                placeholder="All types"
                size="sm"
                data={[
                  { value: 'user', label: 'Users' },
                  { value: 'role', label: 'Roles' },
                ]}
                value={typeFilter}
                onChange={setTypeFilter}
                clearable
                style={{ flex: 1 }}
              />
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
              {formatNumber(filteredPrincipals.length)} of {formatNumber(principalList.length)} principals
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredPrincipals.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="principal"
                    />
                  </Box>
                ) : (
                  paginatedPrincipals.map((p) => (
                    <div
                      key={p.arn}
                      onClick={() => handleSelectPrincipal(p.arn)}
                      style={{
                        cursor: 'pointer',
                        padding: '8px 12px',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        backgroundColor: selectedArn === p.arn
                          ? 'var(--mantine-color-primary-light)'
                          : undefined,
                        borderBottom: '1px solid var(--mantine-color-default-border)',
                      }}
                      onMouseEnter={(e) => {
                        if (selectedArn !== p.arn) {
                          e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (selectedArn !== p.arn) {
                          e.currentTarget.style.backgroundColor = '';
                        }
                      }}
                    >
                      <div style={{ flexShrink: 0 }}>
                        {p.type === 'role' ? (
                          <IconMask size={16} color="var(--mantine-color-orange-6)" />
                        ) : (
                          <IconUser size={16} color="var(--mantine-color-blue-6)" />
                        )}
                      </div>
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <Text size="sm" fw={500} truncate>{p.name}</Text>
                        <Text size="xs" c="dimmed" truncate>{formatAccount(p.accountId)}</Text>
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
              <EmptyState variant="no-selection" entityName="principal" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedPrincipal ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedPrincipal.Name}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedPrincipal} />
                      <ExportButton data={selectedPrincipal} filename={`principal-${selectedPrincipal.Name}`} />
                      <Badge color={selectedPrincipal.Type === 'role' ? 'orange' : 'blue'} size="lg">
                        {selectedPrincipal.Type}
                      </Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all', flex: 1 }}>
                          {selectedPrincipal.Arn}
                        </Text>
                        <CopyButton value={selectedPrincipal.Arn} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Anchor component={Link} to={`/search/accounts/${selectedPrincipal.AccountId}`} size="sm" ff="monospace">
                          {formatAccount(selectedPrincipal.AccountId)}
                        </Anchor>
                        <CopyButton value={selectedPrincipal.AccountId} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Type:</Text>
                        <Text size="sm" ff="monospace">{selectedPrincipal.Type}</Text>
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  {selectedPrincipal.Tags && selectedPrincipal.Tags.length > 0 && (
                    <CollapsibleCard title="Tags">
                      <Table withRowBorders={false}>
                        <Table.Tbody>
                          {selectedPrincipal.Tags.map((tag) => (
                            <Table.Tr key={tag.Key}>
                              <Table.Td w={200} style={{ verticalAlign: 'top' }}>
                                <Text size="sm" fw={500} c="dimmed">{tag.Key}</Text>
                              </Table.Td>
                              <Table.Td>
                                <Text size="sm" style={{ wordBreak: 'break-word' }}>{tag.Value}</Text>
                              </Table.Td>
                            </Table.Tr>
                          ))}
                        </Table.Tbody>
                      </Table>
                    </CollapsibleCard>
                  )}

                  {selectedPrincipal.Groups && selectedPrincipal.Groups.length > 0 && (
                    <CollapsibleCard title="Groups">
                      <Group gap="xs">
                        {selectedPrincipal.Groups.map((group) => (
                          <Badge key={group} variant="light" size="sm">{group}</Badge>
                        ))}
                      </Group>
                    </CollapsibleCard>
                  )}

                  {selectedPrincipal.AttachedPolicies && selectedPrincipal.AttachedPolicies.length > 0 && (
                    <CollapsibleCard title="Attached Policies">
                      <Stack gap="xs">
                        {selectedPrincipal.AttachedPolicies.map((policy) => (
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

                  {selectedPrincipal.PermissionsBoundary && (
                    <CollapsibleCard title="Permission Boundary">
                      <Group gap="xs">
                        <Anchor component={Link} to={`/search/policies/${selectedPrincipal.PermissionsBoundary}`} size="sm" ff="monospace" style={{ flex: 1 }}>
                          {selectedPrincipal.PermissionsBoundary}
                        </Anchor>
                        <CopyButton value={selectedPrincipal.PermissionsBoundary} />
                      </Group>
                    </CollapsibleCard>
                  )}

                  {selectedPrincipal.InlinePolicies && selectedPrincipal.InlinePolicies.length > 0 && (
                    <CollapsibleCard title="Inline Policies">
                      <Stack gap="md">
                        {selectedPrincipal.InlinePolicies.map((policy, index) => {
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
              <EmptyState variant="error" message="Failed to load principal details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
