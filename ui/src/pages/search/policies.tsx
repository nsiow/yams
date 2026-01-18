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
import { IconSearch, IconFileText, IconHierarchy, IconChevronRight } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Policy } from '../../lib/api';
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

type PolicyType = 'iam' | 'scp' | 'rcp';

interface PolicyListItem {
  arn: string;
  accountId: string;
  name: string;
  policyType: PolicyType;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parsePolicyArn(arn: string): PolicyListItem {
  const parts = arn.split(':');
  const service = parts[2] || '';
  const accountId = parts[4] || '';
  const resourcePart = parts.slice(5).join(':');

  let policyType: PolicyType = 'iam';
  let name = resourcePart;

  if (service === 'organizations') {
    // arn:aws:organizations::ACCOUNT:policy/ORG/service_control_policy/ID
    // arn:aws:organizations::ACCOUNT:policy/ORG/resource_control_policy/ID
    if (resourcePart.includes('service_control_policy')) {
      policyType = 'scp';
      name = resourcePart.split('/').pop() || resourcePart;
    } else if (resourcePart.includes('resource_control_policy')) {
      policyType = 'rcp';
      name = resourcePart.split('/').pop() || resourcePart;
    }
  } else {
    // arn:aws:iam::ACCOUNT:policy/NAME
    name = resourcePart.replace(/^policy\//, '');
  }

  return { arn, accountId, name, policyType };
}

const policyTypeLabels: Record<PolicyType, string> = {
  iam: 'IAM Policy',
  scp: 'Service Control Policy',
  rcp: 'Resource Control Policy',
};

const policyTypeColors: Record<PolicyType, string> = {
  iam: 'grape',
  scp: 'orange',
  rcp: 'cyan',
};

// Convert API type (AWS::IAM::Policy, Yams::Organizations::ServiceControlPolicy) to PolicyType
function apiTypeToPolicyType(apiType: string): PolicyType {
  if (apiType.includes('ServiceControlPolicy')) return 'scp';
  if (apiType.includes('ResourceControlPolicy')) return 'rcp';
  return 'iam';
}

export function PoliciesPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': arnFromUrl } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [policyArns, setPolicyArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedPolicy, setSelectedPolicy] = useState<Policy | null>(null);
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

  // Parse all policy ARNs into list items
  const policyList = useMemo(() => {
    return policyArns.map(parsePolicyArn);
  }, [policyArns]);

  // Extract unique types for filter dropdown
  const typeOptions = useMemo(() => {
    const types = new Set(policyList.map(p => p.policyType));
    return Array.from(types).sort().map(t => ({
      value: t,
      label: policyTypeLabels[t],
    }));
  }, [policyList]);

  // Extract unique accounts for filter dropdown
  const accountOptions = useMemo(() => {
    const ids = new Set(policyList.map(p => p.accountId));
    return Array.from(ids)
      .sort()
      .map(id => {
        const name = accountNames[id];
        const label = name ? `${name} (${id})` : id;
        return { value: id, label };
      });
  }, [policyList, accountNames]);

  // Helper to format account display
  const formatAccount = (accountId: string): string => {
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  // Filter policies based on search and filters
  const filteredPolicies = useMemo(() => {
    return policyList.filter(p => {
      if (typeFilter && p.policyType !== typeFilter) return false;
      if (accountFilter && p.accountId !== accountFilter) return false;
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return p.name.toLowerCase().includes(query) || p.arn.toLowerCase().includes(query);
      }
      return true;
    });
  }, [policyList, typeFilter, accountFilter, debouncedSearch]);

  // Fetch all policy ARNs and account names on mount
  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [arns, names] = await Promise.all([
          yamsApi.listPolicies(),
          yamsApi.accountNames(),
        ]);
        setPolicyArns(arns);
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

  // Fetch policy detail when selected
  const fetchPolicyDetail = useCallback(async (arn: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const policy = await yamsApi.getPolicy(arn);
      setSelectedPolicy(policy);
    } catch (err) {
      console.error('Failed to fetch policy detail:', err);
      setSelectedPolicy(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectPolicy = (arn: string): void => {
    setSelectedArn(arn);
    fetchPolicyDetail(arn);
    navigate(`/search/policies/${arn}?${searchParams.toString()}`, { replace: true });
  };

  // Load policy from URL on mount or when URL changes
  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchPolicyDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchPolicyDetail, selectedArn]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredPolicies.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, typeFilter, accountFilter]);

  const paginatedPolicies = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredPolicies.slice(start, start + itemsPerPage);
  }, [filteredPolicies, page]);

  // Breadcrumb items
  const getPolicyDisplayName = (): string => {
    if (!selectedPolicy) return '';
    return selectedPolicy.Name || parsePolicyArn(selectedPolicy.Arn).name;
  };

  const breadcrumbItems = [
    { title: 'Search', href: '/' },
    { title: 'Policies', href: '/search/policies' },
    ...(selectedPolicy ? [{ title: getPolicyDisplayName(), href: '' }] : []),
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
            <Title order={2}>Policies</Title>

            <TextInput
              placeholder="Search policies..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            <FilterBar hasActiveFilters={hasActiveFilters} onClearAll={clearAllFilters}>
              <Select
                placeholder="All types"
                size="sm"
                data={typeOptions}
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
              {formatNumber(filteredPolicies.length)} of {formatNumber(policyList.length)} policies
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {loading ? (
                  <ListSkeleton />
                ) : filteredPolicies.length === 0 ? (
                  <Box p="xl">
                    <EmptyState
                      variant={hasActiveFilters ? 'no-results' : 'no-data'}
                      entityName="policy"
                    />
                  </Box>
                ) : (
                  paginatedPolicies.map((p) => (
                    <div
                      key={p.arn}
                      onClick={() => handleSelectPolicy(p.arn)}
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
                        {p.policyType === 'iam' ? (
                          <IconFileText size={16} color={`var(--mantine-color-${policyTypeColors[p.policyType]}-6)`} />
                        ) : (
                          <IconHierarchy size={16} color={`var(--mantine-color-${policyTypeColors[p.policyType]}-6)`} />
                        )}
                      </div>
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <Text size="sm" fw={500} truncate>{p.name}</Text>
                        <Text size="xs" c="dimmed" truncate>
                          {policyTypeLabels[p.policyType]} · {formatAccount(p.accountId)}
                        </Text>
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
              <EmptyState variant="no-selection" entityName="policy" />
            ) : loadingDetail ? (
              <DetailSkeleton />
            ) : selectedPolicy ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedPolicy.Name || parsePolicyArn(selectedPolicy.Arn).name}</Title>
                    <Group gap="xs">
                      <CopyEntityButton data={selectedPolicy} />
                      <ExportButton data={selectedPolicy} filename={`policy-${selectedPolicy.Name || parsePolicyArn(selectedPolicy.Arn).name}`} />
                      <Badge color={policyTypeColors[apiTypeToPolicyType(selectedPolicy.Type)]} size="lg">
                        {policyTypeLabels[apiTypeToPolicyType(selectedPolicy.Type)]}
                      </Badge>
                    </Group>
                  </Group>

                  <CollapsibleCard title="Metadata">
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all', flex: 1 }}>
                          {selectedPolicy.Arn}
                        </Text>
                        <CopyButton value={selectedPolicy.Arn} />
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Anchor component={Link} to={`/search/accounts/${selectedPolicy.AccountId}`} size="sm" ff="monospace">
                          {formatAccount(selectedPolicy.AccountId)}
                        </Anchor>
                        <CopyButton value={selectedPolicy.AccountId} />
                      </Group>
                    </Stack>
                  </CollapsibleCard>

                  <CollapsibleCard title="Policy Document">
                    <CodeHighlight
                      code={JSON.stringify(selectedPolicy.Policy, null, 2)}
                      language="json"
                      withCopyButton
                    />
                  </CollapsibleCard>
                </Stack>
              </ScrollArea>
            ) : (
              <EmptyState variant="error" message="Failed to load policy details" />
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
