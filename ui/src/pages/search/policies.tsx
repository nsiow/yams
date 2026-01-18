import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Alert,
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
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import { IconAlertCircle, IconSearch, IconFileText, IconHierarchy } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Policy } from '../../lib/api';

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

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [policyArns, setPolicyArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedPolicy, setSelectedPolicy] = useState<Policy | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [accountFilter, setAccountFilter] = useState<string | null>(null);

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
      // Type filter
      if (typeFilter && p.policyType !== typeFilter) {
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
    navigate(`/search/policies/${arn}`, { replace: true });
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

  if (loading) {
    return (
      <Box p="xl">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Loading policies...</Text>
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
            <Title order={2}>Policies</Title>

            {/* Search box */}
            <TextInput
              placeholder="Search policies..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            {/* Type and Account filters */}
            <Group gap="sm" grow>
              <Select
                placeholder="All types"
                size="sm"
                data={typeOptions}
                value={typeFilter}
                onChange={setTypeFilter}
                clearable
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
              {formatNumber(filteredPolicies.length)} of {formatNumber(policyList.length)} policies
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            {/* Policy list - paginated */}
            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {paginatedPolicies.map((p) => (
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
                      <Text size="sm" fw={500} truncate>
                        {p.name}
                      </Text>
                      <Text size="xs" c="dimmed" truncate>
                        {policyTypeLabels[p.policyType]} · {formatAccount(p.accountId)}
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
                <Text c="dimmed">Select a policy to view details</Text>
              </Stack>
            ) : loadingDetail ? (
              <Stack align="center" justify="center" h="100%">
                <Loader size="md" />
                <Text c="dimmed">Loading details...</Text>
              </Stack>
            ) : selectedPolicy ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  {/* Header */}
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedPolicy.Name || parsePolicyArn(selectedPolicy.Arn).name}</Title>
                    <Badge color={policyTypeColors[apiTypeToPolicyType(selectedPolicy.Type)]} size="lg">
                      {policyTypeLabels[apiTypeToPolicyType(selectedPolicy.Type)]}
                    </Badge>
                  </Group>

                  {/* Metadata */}
                  <Card withBorder p="sm">
                    <Title order={5} mb="xs">Metadata</Title>
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all' }}>{selectedPolicy.Arn}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Text size="sm" ff="monospace">{formatAccount(selectedPolicy.AccountId)}</Text>
                      </Group>
                    </Stack>
                  </Card>

                  {/* Policy Document */}
                  <Card withBorder p="sm">
                    <Title order={5} mb="xs">Policy Document</Title>
                    <CodeHighlight
                      code={JSON.stringify(selectedPolicy.Policy, null, 2)}
                      language="json"
                      withCopyButton
                    />
                  </Card>
                </Stack>
              </ScrollArea>
            ) : (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Failed to load policy details</Text>
              </Stack>
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
