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
  Table,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { CodeHighlight } from '@mantine/code-highlight';
import { IconAlertCircle, IconSearch, IconDatabase } from '@tabler/icons-react';
import { yamsApi } from '../../lib/api';
import type { Resource } from '../../lib/api';

import '@mantine/code-highlight/styles.css';

interface ResourceListItem {
  arn: string;
  service: string;
  region: string;
  accountId: string;
  resourceType: string;
  name: string;
}

function formatNumber(n: number): string {
  return n.toLocaleString();
}

function parseResourceArn(arn: string): ResourceListItem {
  // ARN format: arn:aws:service:region:account:resource-type/resource-name
  // or: arn:aws:service:region:account:resource-type:resource-name
  const parts = arn.split(':');
  const service = parts[2] || '';
  const region = parts[3] || '';
  const accountId = parts[4] || '';

  // Resource part can be type/name, type:name, or just name
  const resourceParts = parts.slice(5).join(':');
  let resourceType = '';
  let name = resourceParts;

  if (resourceParts.includes('/')) {
    const idx = resourceParts.indexOf('/');
    resourceType = resourceParts.substring(0, idx);
    name = resourceParts.substring(idx + 1);
  } else if (resourceParts.includes(':')) {
    const idx = resourceParts.indexOf(':');
    resourceType = resourceParts.substring(0, idx);
    name = resourceParts.substring(idx + 1);
  }

  return { arn, service, region, accountId, resourceType, name };
}

// Format service:resourceType for display (lowercase)
function formatResourceType(item: ResourceListItem): string {
  if (item.resourceType) {
    return `${item.service}:${item.resourceType}`.toLowerCase();
  }
  return item.service.toLowerCase();
}

// Convert AWS CloudFormation type (AWS::S3::Bucket) to service:type format
function formatCloudFormationType(cfnType: string): string {
  // AWS::Service::ResourceType -> service:resourcetype
  const parts = cfnType.split('::');
  if (parts.length >= 3) {
    return `${parts[1]}:${parts[2]}`.toLowerCase();
  }
  return cfnType.toLowerCase();
}

export function ResourcesPage(): JSX.Element {
  const navigate = useNavigate();
  const { '*': arnFromUrl } = useParams();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [resourceArns, setResourceArns] = useState<string[]>([]);
  const [accountNames, setAccountNames] = useState<Record<string, string>>({});
  const [selectedArn, setSelectedArn] = useState<string | null>(arnFromUrl || null);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch] = useDebouncedValue(searchQuery, 200);
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [regionFilter, setRegionFilter] = useState<string | null>(null);
  const [accountFilter, setAccountFilter] = useState<string | null>(null);

  // Parse all resource ARNs into list items
  const resourceList = useMemo(() => {
    return resourceArns.map(parseResourceArn);
  }, [resourceArns]);

  // Extract unique types for filter dropdown
  const typeOptions = useMemo(() => {
    const types = new Set(resourceList.map(formatResourceType));
    return Array.from(types).sort().map(t => ({ value: t, label: t }));
  }, [resourceList]);

  // Extract unique regions for filter dropdown
  const regionOptions = useMemo(() => {
    const regions = new Set(resourceList.map(r => r.region).filter(Boolean));
    return Array.from(regions).sort().map(r => ({ value: r, label: r }));
  }, [resourceList]);

  // Extract unique accounts for filter dropdown
  const accountOptions = useMemo(() => {
    const ids = new Set(resourceList.map(r => r.accountId).filter(Boolean));
    return Array.from(ids)
      .sort()
      .map(id => {
        const name = accountNames[id];
        const label = name ? `${name} (${id})` : id;
        return { value: id, label };
      });
  }, [resourceList, accountNames]);

  // Helper to format account display
  const formatAccount = (accountId: string): string => {
    if (!accountId) return 'N/A';
    const name = accountNames[accountId];
    return name ? `${name} (${accountId})` : accountId;
  };

  // Filter resources based on search and filters
  const filteredResources = useMemo(() => {
    return resourceList.filter(r => {
      // Type filter
      if (typeFilter && formatResourceType(r) !== typeFilter) {
        return false;
      }
      // Region filter
      if (regionFilter && r.region !== regionFilter) {
        return false;
      }
      // Account filter
      if (accountFilter && r.accountId !== accountFilter) {
        return false;
      }
      // Search filter (searches name and ARN)
      if (debouncedSearch) {
        const query = debouncedSearch.toLowerCase();
        return (
          r.name.toLowerCase().includes(query) ||
          r.arn.toLowerCase().includes(query)
        );
      }
      return true;
    });
  }, [resourceList, typeFilter, regionFilter, accountFilter, debouncedSearch]);

  // Fetch all resource ARNs and account names on mount
  useEffect(() => {
    async function fetchData(): Promise<void> {
      try {
        const [arns, names] = await Promise.all([
          yamsApi.listResources(),
          yamsApi.accountNames(),
        ]);
        setResourceArns(arns);
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

  // Fetch resource detail when selected
  const fetchResourceDetail = useCallback(async (arn: string): Promise<void> => {
    setLoadingDetail(true);
    try {
      const resource = await yamsApi.getResource(arn);
      setSelectedResource(resource);
    } catch (err) {
      console.error('Failed to fetch resource detail:', err);
      setSelectedResource(null);
    } finally {
      setLoadingDetail(false);
    }
  }, []);

  const handleSelectResource = (arn: string): void => {
    setSelectedArn(arn);
    fetchResourceDetail(arn);
    navigate(`/search/resources/${arn}`, { replace: true });
  };

  // Load resource from URL on mount or when URL changes
  useEffect(() => {
    if (arnFromUrl && arnFromUrl !== selectedArn) {
      setSelectedArn(arnFromUrl);
      fetchResourceDetail(arnFromUrl);
    }
  }, [arnFromUrl, fetchResourceDetail, selectedArn]);

  // Pagination
  const [page, setPage] = useState(1);
  const itemsPerPage = 20;
  const totalPages = Math.ceil(filteredResources.length / itemsPerPage);

  // Reset to page 1 when filters change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, typeFilter, regionFilter, accountFilter]);

  const paginatedResources = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredResources.slice(start, start + itemsPerPage);
  }, [filteredResources, page]);

  if (loading) {
    return (
      <Box p="xl">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text c="dimmed">Loading resources...</Text>
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
            <Title order={2}>Resources</Title>

            {/* Search box */}
            <TextInput
              placeholder="Search resources..."
              leftSection={<IconSearch size={16} />}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.currentTarget.value)}
            />

            {/* Filters */}
            <Group gap="sm" grow>
              <Select
                placeholder="All types"
                size="sm"
                data={typeOptions}
                value={typeFilter}
                onChange={setTypeFilter}
                clearable
                searchable
              />
              <Select
                placeholder="All regions"
                size="sm"
                data={regionOptions}
                value={regionFilter}
                onChange={setRegionFilter}
                clearable
                searchable
              />
            </Group>
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

            {/* Results count */}
            <Text size="sm" c="dimmed">
              {formatNumber(filteredResources.length)} of {formatNumber(resourceList.length)} resources
              {totalPages > 1 && ` (page ${formatNumber(page)} of ${formatNumber(totalPages)})`}
            </Text>

            {/* Resource list - paginated */}
            <Card withBorder p={0} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
              <ScrollArea style={{ flex: 1 }}>
                {paginatedResources.map((r) => (
                  <div
                    key={r.arn}
                    onClick={() => handleSelectResource(r.arn)}
                    style={{
                      cursor: 'pointer',
                      padding: '8px 12px',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '12px',
                      backgroundColor: selectedArn === r.arn
                        ? 'var(--mantine-color-primary-light)'
                        : undefined,
                      borderBottom: '1px solid var(--mantine-color-default-border)',
                    }}
                    onMouseEnter={(e) => {
                      if (selectedArn !== r.arn) {
                        e.currentTarget.style.backgroundColor = 'var(--mantine-color-gray-light-hover)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (selectedArn !== r.arn) {
                        e.currentTarget.style.backgroundColor = '';
                      }
                    }}
                  >
                    <div style={{ flexShrink: 0 }}>
                      <IconDatabase size={16} color="var(--mantine-color-teal-6)" />
                    </div>
                    <div style={{ minWidth: 0, flex: 1 }}>
                      <Text size="sm" fw={500} truncate>
                        {r.name}
                      </Text>
                      <Text size="xs" c="dimmed" truncate>
                        {formatResourceType(r)} &middot; {r.region || 'global'}
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
                <Text c="dimmed">Select a resource to view details</Text>
              </Stack>
            ) : loadingDetail ? (
              <Stack align="center" justify="center" h="100%">
                <Loader size="md" />
                <Text c="dimmed">Loading details...</Text>
              </Stack>
            ) : selectedResource ? (
              <ScrollArea h="calc(100vh - 180px)">
                <Stack gap="md">
                  {/* Header */}
                  <Group justify="space-between" align="flex-start">
                    <Title order={3}>{selectedResource.Name}</Title>
                    <Badge color="teal" size="lg">
                      {formatCloudFormationType(selectedResource.Type)}
                    </Badge>
                  </Group>

                  {/* Metadata */}
                  <Card withBorder p="sm">
                    <Title order={5} mb="xs">Metadata</Title>
                    <Stack gap="xs">
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>ARN:</Text>
                        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all' }}>{selectedResource.Arn}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Account:</Text>
                        <Text size="sm" ff="monospace">{formatAccount(selectedResource.AccountId)}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Region:</Text>
                        <Text size="sm" ff="monospace">{selectedResource.Region || 'global'}</Text>
                      </Group>
                      <Group gap="xs">
                        <Text size="sm" fw={600} c="dimmed" w={100}>Type:</Text>
                        <Text size="sm" ff="monospace">{formatCloudFormationType(selectedResource.Type)}</Text>
                      </Group>
                    </Stack>
                  </Card>

                  {/* Tags */}
                  {selectedResource.Tags && selectedResource.Tags.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Tags</Title>
                      <Table withRowBorders={false}>
                        <Table.Tbody>
                          {selectedResource.Tags.map((tag) => (
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

                  {/* Resource Policy */}
                  {selectedResource.Policy && selectedResource.Policy.Statement && selectedResource.Policy.Statement.length > 0 && (
                    <Card withBorder p="sm">
                      <Title order={5} mb="xs">Resource Policy</Title>
                      <CodeHighlight
                        code={JSON.stringify(selectedResource.Policy, null, 2)}
                        language="json"
                        withCopyButton
                      />
                    </Card>
                  )}
                </Stack>
              </ScrollArea>
            ) : (
              <Stack align="center" justify="center" h="100%">
                <Text c="dimmed">Failed to load resource details</Text>
              </Stack>
            )}
          </Card>
        </Grid.Col>
      </Grid>
    </Box>
  );
}
