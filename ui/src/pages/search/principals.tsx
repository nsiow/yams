import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import {
  Alert,
  Anchor,
  Badge,
  Box,
  Card,
  Grid,
  Group,
  Loader,
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
import { IconAlertCircle, IconSearch, IconUser, IconMask } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Principal } from '../../lib/api';

import '@mantine/code-highlight/styles.css';

interface PrincipalListItem {
  arn: string;
  type: 'user' | 'role';
  accountId: string;
  name: string;
}

function parsePrincipalArn(arn: string): PrincipalListItem {
  // ARN format: arn:aws:iam::ACCOUNT_ID:user/NAME or arn:aws:iam::ACCOUNT_ID:role/NAME
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

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [principalArns, setPrincipalArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedPrincipal, setSelectedPrincipal] = useState<Principal | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [accountFilter, setAccountFilter] = useState<string | null>(null);

  // Parse all principal ARNs into list items
  const principalList = useMemo(() => {
    return principalArns.map(parsePrincipalArn);
  }, [principalArns]);

  // Extract unique account IDs and build display labels
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

  // Helper to format account display
  const formatAccount = (accountId: string): string => {
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  // Filter principals based on search and filters
  const filteredPrincipals = useMemo(() => {
    return principalList.filter(p => {
      // Type filter
      if (typeFilter && p.type !== typeFilter) {
        return false;
      }
      // Account filter
      if (accountFilter && p.accountId !== accountFilter) {
        return false;
      }
      // Search filter (searches name and ARN)
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return (
          p.name.toLowerCase().includes(query) ||
          p.arn.toLowerCase().includes(query)
        );
      }
      return true;
    });
  }, [principalList, typeFilter, accountFilter, debouncedSearch]);

  // Fetch all principal ARNs and account names on mount
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

  // Fetch principal detail when selected
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
    navigate(`/search/principals/${arn}`, { replace: true });
  };

  // Load principal from URL on mount or when URL changes
  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchPrincipalDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchPrincipalDetail, selectedArn]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredPrincipals.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, typeFilter, accountFilter]);

  const paginatedPrincipals = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredPrincipals.slice(start, start + itemsPerPage);
  }, [filteredPrincipals, page]);

  if (loading) {
    return (
      <Box p="xl">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Loading principals...</Text>
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
            <Title order={2}>Principals</Title>

            {/* Search box */}
            <TextInput
              placeholder="Search principals..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            {/* Type and Account filters */}
            <Group gap="sm" grow>
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
                searchable
              />
              <Select
                placeholder="All accounts"
                size="sm"
                data={accountOptions}
                value={accountFilter}
                onChange={setAccountFilter}
                clearable
                searchable
                renderOption={({ option }) => {
                  const name = accountNames[option.value];
                  if (name) {
                    return (
                      <Group gap="xs">
                        <Text size="sm">{name}</Text>
                        <Text size="xs" c="dimmed">({option.value})</Text>
                      </Group>
                    );
                  }
                  return <Text size="sm">{option.value}</Text>;
                }}
              />
            </Group>

            {/* Results count */}
            <Text size="sm" c="dimmed">
              {filteredPrincipals.length} of {principalList.length} principals
              {totalPages > 1 && ` (page ${page} of ${totalPages})`}
            </Text>

            {/* Principal list - paginated */}
            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {paginatedPrincipals.map((p) => (
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
                      <Text size="sm" fw={500} truncate>
                        {p.name}
                      </Text>
                      <Text size="xs" c="dimmed" truncate>
                        {formatAccount(p.accountId)}
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
            {!selectedArn ? (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Select a principal to view details</Text>
              </Stack>
            ) : loadingDetail ? (
              <Stack align="center" justify="center" h="100%">
                <Loader size="md" />
                <Text c="dimmed">Loading details...</Text>
              </Stack>
            ) : selectedPrincipal ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  {/* Header */}
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedPrincipal.Name}</Title>
                    <Badge
                      color={selectedPrincipal.Type === 'role' ? 'orange' : 'blue'}
                      size="lg"
                    >
                      {selectedPrincipal.Type}
                    </Badge>
                  </Group>

                  {/* Metadata */}
                  <Card withBorder p="sm">
                    <Title order={5} mb="xs">Metadata</Title>
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" style={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>{selectedPrincipal.Arn}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Text size="sm">{formatAccount(selectedPrincipal.AccountId)}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Type:</Text>
                        <Text size="sm">{selectedPrincipal.Type}</Text>
                      </Group>
                    </Stack>
                  </Card>

                  {/* Tags */}
                  {selectedPrincipal.Tags && selectedPrincipal.Tags.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Tags</Title>
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
                    </Card>
                  )}

                  {/* Groups */}
                  {selectedPrincipal.Groups && selectedPrincipal.Groups.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Groups</Title>
                      <Group gap="xs">
                        {selectedPrincipal.Groups.map((group) => (
                          <Badge key={group} variant="light" size="sm">
                            {group}
                          </Badge>
                        ))}
                      </Group>
                    </Card>
                  )}

                  {/* Attached Policies */}
                  {selectedPrincipal.AttachedPolicies && selectedPrincipal.AttachedPolicies.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Attached Policies</Title>
                      <Stack gap="xs">
                        {selectedPrincipal.AttachedPolicies.map((policy) => (
                          <Anchor
                            key={policy}
                            component={Link}
                            to={`/search/policies/${policy}`}
                            size="sm"
                            style={{ fontFamily: 'monospace' }}
                          >
                            {policy}
                          </Anchor>
                        ))}
                      </Stack>
                    </Card>
                  )}

                  {/* Permission Boundary */}
                  {selectedPrincipal.PermissionsBoundary && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Permission Boundary</Title>
                      <Anchor
                        component={Link}
                        to={`/search/policies/${selectedPrincipal.PermissionsBoundary}`}
                        size="sm"
                        style={{ fontFamily: 'monospace' }}
                      >
                        {selectedPrincipal.PermissionsBoundary}
                      </Anchor>
                    </Card>
                  )}

                  {/* Inline Policies */}
                  {selectedPrincipal.InlinePolicies && selectedPrincipal.InlinePolicies.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Inline Policies</Title>
                      <Stack gap="md">
                        {selectedPrincipal.InlinePolicies.map((policy, index) => (
                          <Box key={index}>
                            <Text size="sm" fw={600} mb="xs">
                              {policy.Id || `Policy ${index + 1}`}
                            </Text>
                            <CodeHighlight
                              code={JSON.stringify(policy, null, 2)}
                              language="json"
                              withCopyButton
                            />
                          </Box>
                        ))}
                      </Stack>
                    </Card>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Failed to load principal details</Text>
              </Stack>
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
